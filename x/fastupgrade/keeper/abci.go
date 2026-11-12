package keeper

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"cosmossdk.io/collections"
	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/xrplevm/node/v12/x/fastupgrade/types"
)

// EndBlock runs the module in a scratch copy of state and saves it only when
// everything succeeded. On an error or a panic it discards that copy, drops
// only the timer, and returns nil, so the module's checks do not halt the chain.
// On every path it then clears this block's nominations from the real state.
func (k Keeper) EndBlock(goCtx context.Context) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	defer k.clearNominations(ctx)

	work, save := ctx.CacheContext()
	err := k.endBlockRecovered(work)
	if err == nil {
		save()
		return nil
	}
	k.Logger(ctx).Error("fast upgrade end block failed; discarding its changes and dropping the timer", "error", err)

	if err := k.dropTimerAfterFailure(ctx); err != nil {
		k.Logger(ctx).Error("failed to drop fast upgrade timer", "error", err)
	}
	return nil
}

// dropTimerAfterFailure announces the drop like every other one. Without a
// timer there is nothing to retract, so it emits nothing.
func (k Keeper) dropTimerAfterFailure(ctx sdk.Context) error {
	timer, err := k.Timer.Get(ctx)
	switch {
	case errors.Is(err, collections.ErrNotFound):
		return nil
	case err != nil:
		// The timer cannot be read, so there is no event to build.
		return k.Timer.Remove(ctx)
	}
	return k.cancelTimer(ctx, timer, types.ReasonInternalError)
}

func (k Keeper) clearNominations(ctx sdk.Context) {
	if err := k.Nominations.Clear(ctx, nil); err != nil {
		k.Logger(ctx).Error("failed to clear fast upgrade nominations", "error", err)
	}
}

// endBlockRecovered turns a panic into an error. Gov's Tally, for example,
// divides by the delegator shares of every bonded validator that voted.
func (k Keeper) endBlockRecovered(ctx sdk.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return k.endBlock(ctx)
}

func (k Keeper) endBlock(ctx sdk.Context) error {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}
	timer, err := k.Timer.Get(ctx)
	running := err == nil
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return err
	}
	if !params.Enabled {
		if !running {
			return nil
		}
		return k.cancelTimer(ctx, timer, types.ReasonDisabled)
	}
	threshold, err := params.ThresholdDec()
	if err != nil {
		return err
	}
	if running {
		return k.resolveTimer(ctx, timer, params.PostponeBlocks, threshold)
	}
	return k.checkNominations(ctx, params, threshold)
}

// resolveTimer waits until H-1, then tries to close the vote early. A timer
// does not outlive its proposal's deadline: gov ends that vote later in this
// block, and stock gov decides.
func (k Keeper) resolveTimer(ctx sdk.Context, timer types.Timer, postpone uint64, threshold sdkmath.LegacyDec) error {
	p, voting, err := k.votingProposal(ctx, timer.ProposalId)
	if err != nil {
		return err
	}
	finalCheck := timer.UpgradeHeight - 1
	switch {
	case !voting:
		return k.cancelTimer(ctx, timer, types.ReasonNotVoting)
	case ctx.BlockHeight() > finalCheck:
		return k.cancelTimer(ctx, timer, types.ReasonMissedWindow)
	case !p.VotingEndTime.After(ctx.BlockTime()):
		return k.cancelTimer(ctx, timer, types.ReasonNaturalDeadline)
	case ctx.BlockHeight() < finalCheck:
		return nil
	}
	return k.closeVoteEarly(ctx, p, timer, postpone, threshold)
}

// cancelTimer removes the timer without firing and emits Dropped.
func (k Keeper) cancelTimer(ctx sdk.Context, timer types.Timer, reason types.Reason) error {
	if err := k.Timer.Remove(ctx); err != nil {
		return err
	}
	k.Logger(ctx).Info("fast upgrade timer dropped", "proposal", timer.ProposalId, "upgrade_height", timer.UpgradeHeight, "reason", string(reason))
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeDropped,
		sdk.NewAttribute(types.AttributeKeyProposalID, strconv.FormatUint(timer.ProposalId, 10)),
		sdk.NewAttribute(types.AttributeKeyUpgradeHeight, strconv.FormatInt(timer.UpgradeHeight, 10)),
		sdk.NewAttribute(types.AttributeKeyReason, string(reason)),
	))
	return nil
}

// checkNominations checks this block's nominations, newest first. The first
// proposal that passes the validator gate decides the block, so at most one
// full tally runs.
func (k Keeper) checkNominations(ctx sdk.Context, params types.Params, threshold sdkmath.LegacyDec) error {
	iter, err := k.Nominations.Iterate(ctx, new(collections.Range[uint64]).Descending())
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		id, err := iter.Key()
		if err != nil {
			return err
		}
		p, voting, err := k.votingProposal(ctx, id)
		if err != nil {
			return err
		}
		if !voting {
			continue
		}
		// A plan below H-1 would be in the past at the final check, so its
		// timer could only drop there while holding the timer slot.
		upgradeHeight, err := addBlocks(ctx.BlockHeight(), params.TimelockBlocks)
		if err != nil {
			return err
		}
		msg, reason, err := k.guards(ctx, p, upgradeHeight-1, threshold)
		if err != nil {
			return err
		}
		switch reason {
		case types.ReasonOK:
			return k.startTimer(ctx, id, msg.Plan, upgradeHeight)
		case types.ReasonGateBelowThreshold, types.ReasonNaturalDeadline, types.ReasonNotEligible:
			continue
		case types.ReasonPlanPending:
			return nil
		default:
			k.emitNotStarted(ctx, id, reason)
			return nil
		}
	}
	return nil
}

func (k Keeper) startTimer(ctx sdk.Context, proposalID uint64, plan upgradetypes.Plan, upgradeHeight int64) error {
	if err := k.Timer.Set(ctx, types.Timer{ProposalId: proposalID, UpgradeHeight: upgradeHeight}); err != nil {
		return err
	}
	k.Logger(ctx).Info("fast upgrade timer started", "proposal", proposalID, "plan", plan.Name, "upgrade_height", upgradeHeight)
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeStarted,
		sdk.NewAttribute(types.AttributeKeyProposalID, strconv.FormatUint(proposalID, 10)),
		sdk.NewAttribute(types.AttributeKeyPlanName, plan.Name),
		sdk.NewAttribute(types.AttributeKeyPlanInfo, plan.Info),
		sdk.NewAttribute(types.AttributeKeyUpgradeHeight, strconv.FormatInt(upgradeHeight, 10)),
	))
	return nil
}

func (k Keeper) emitNotStarted(ctx sdk.Context, proposalID uint64, reason types.Reason) {
	k.Logger(ctx).Info("fast upgrade timer not started", "proposal", proposalID, "reason", string(reason))
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeNotStarted,
		sdk.NewAttribute(types.AttributeKeyProposalID, strconv.FormatUint(proposalID, 10)),
		sdk.NewAttribute(types.AttributeKeyReason, string(reason)),
	))
}

// closeVoteEarly runs every guard before any gov write. If they all pass, it
// rewrites the plan height to H and ends the voting period now, so gov's
// EndBlocker executes the proposal in this block. Gov runs after this module,
// so a proposal it executes first in this block could change what its tally
// of ours reads. When another proposal is due, H moves later and every guard
// runs again on the settled state.
func (k Keeper) closeVoteEarly(ctx sdk.Context, p govv1.Proposal, timer types.Timer, postpone uint64, threshold sdkmath.LegacyDec) error {
	msg, reason, err := k.guards(ctx, p, ctx.BlockHeight(), threshold)
	if err != nil {
		return err
	}
	if reason != types.ReasonOK {
		return k.cancelTimer(ctx, timer, reason)
	}
	due, err := k.proposalDue(ctx)
	if err != nil {
		return err
	}
	if due {
		return k.postponeTimer(ctx, timer, postpone)
	}

	if err := k.Timer.Remove(ctx); err != nil {
		return err
	}
	if err := setPlanHeight(&p, msg, timer.UpgradeHeight); err != nil {
		return err
	}
	if err := k.endVotingNow(ctx, p); err != nil {
		return err
	}

	k.Logger(ctx).Info("fast upgrade vote closed early", "proposal", p.Id, "plan", msg.Plan.Name, "upgrade_height", timer.UpgradeHeight)
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeEarlyClose,
		sdk.NewAttribute(types.AttributeKeyProposalID, strconv.FormatUint(p.Id, 10)),
		sdk.NewAttribute(types.AttributeKeyPlanName, msg.Plan.Name),
		sdk.NewAttribute(types.AttributeKeyUpgradeHeight, strconv.FormatInt(timer.UpgradeHeight, 10)),
	))
	return nil
}

// proposalDue reports whether gov's EndBlocker ends a voting period in this
// block. It uses gov's own range over the queue.
func (k Keeper) proposalDue(ctx sdk.Context) (bool, error) {
	iter, err := k.gov.ActiveProposalsQueue.Iterate(ctx, collections.NewPrefixUntilPairRange[time.Time, uint64](ctx.BlockTime()))
	if err != nil {
		return false, err
	}
	defer iter.Close()
	return iter.Valid(), nil
}

func (k Keeper) postponeTimer(ctx sdk.Context, timer types.Timer, postpone uint64) error {
	upgradeHeight, err := addBlocks(timer.UpgradeHeight, postpone)
	if err != nil {
		return err
	}
	if err := k.Timer.Set(ctx, types.Timer{ProposalId: timer.ProposalId, UpgradeHeight: upgradeHeight}); err != nil {
		return err
	}
	k.Logger(ctx).Info("fast upgrade timer postponed", "proposal", timer.ProposalId, "upgrade_height", upgradeHeight)
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypePostponed,
		sdk.NewAttribute(types.AttributeKeyProposalID, strconv.FormatUint(timer.ProposalId, 10)),
		sdk.NewAttribute(types.AttributeKeyUpgradeHeight, strconv.FormatInt(upgradeHeight, 10)),
	))
	return nil
}

func addBlocks(height int64, blocks uint64) (int64, error) {
	if blocks > math.MaxInt64 || height > math.MaxInt64-int64(blocks) {
		return 0, fmt.Errorf("upgrade height overflows: height %d + %d blocks", height, blocks)
	}
	return height + int64(blocks), nil
}

// guards runs the fast-path checks shared by timer start and H-1.
func (k Keeper) guards(ctx sdk.Context, p govv1.Proposal, minPlanHeight int64, threshold sdkmath.LegacyDec) (*upgradetypes.MsgSoftwareUpgrade, types.Reason, error) {
	if !p.VotingEndTime.After(ctx.BlockTime()) {
		return nil, types.ReasonNaturalDeadline, nil
	}
	msg, err := k.eligibleMinorMessage(ctx, p, minPlanHeight)
	if err != nil {
		return nil, "", err
	}
	if msg == nil {
		return nil, types.ReasonNotEligible, nil
	}
	pending, err := k.upgradePlanExists(ctx)
	if err != nil {
		return nil, "", err
	}
	if pending {
		return nil, types.ReasonPlanPending, nil
	}
	reason, err := k.supportCheck(ctx, p, threshold)
	return msg, reason, err
}

// setPlanHeight repacks the stored message with the new height. Gov executes
// the stored Any, so editing the decoded message alone would save nothing.
func setPlanHeight(p *govv1.Proposal, msg *upgradetypes.MsgSoftwareUpgrade, height int64) error {
	updated := *msg
	updated.Plan.Height = height
	if err := updated.Plan.ValidateBasic(); err != nil {
		return err
	}
	packed, err := codectypes.NewAnyWithValue(&updated)
	if err != nil {
		return err
	}
	p.Messages = []*codectypes.Any{packed}
	return nil
}

// endVotingNow moves the proposal's queue key and VotingEndTime together:
// gov removes the key using VotingEndTime and processes any due key.
func (k Keeper) endVotingNow(ctx sdk.Context, p govv1.Proposal) error {
	now := ctx.BlockTime()
	if err := k.gov.ActiveProposalsQueue.Remove(ctx, collections.Join(*p.VotingEndTime, p.Id)); err != nil {
		return err
	}
	p.VotingEndTime = &now
	if err := k.gov.SetProposal(ctx, p); err != nil {
		return err
	}
	return k.gov.ActiveProposalsQueue.Set(ctx, collections.Join(now, p.Id), p.Id)
}
