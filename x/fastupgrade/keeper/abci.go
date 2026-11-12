package keeper

import (
	"context"
	"errors"
	"fmt"
	"math"
	"runtime/debug"
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

	cacheCtx, writeCache := ctx.CacheContext()
	err := k.endBlockNoPanic(cacheCtx)
	if err == nil {
		writeCache()
		return nil
	}
	k.Logger(ctx).Error("fast upgrade end block failed; discarding its changes and dropping the timer", "error", err)

	if err := k.cancelTimerOnInternalError(ctx); err != nil {
		k.Logger(ctx).Error("failed to drop fast upgrade timer", "error", err)
	}
	return nil
}

// cancelTimerOnInternalError removes the timer after a failed run. The failed
// run changed no state, so this is a safety choice, not a repair. A timer
// announces an upgrade at H, and after an internal failure the module cannot
// stand behind that: the same state would fail again, and a timer that misses
// its H-1 check is dropped in the next block anyway. Without a timer, stock
// gov decides the proposal with its original plan.
//
// It announces the drop like every other one. Without a timer there is
// nothing to retract, so it emits nothing.
func (k Keeper) cancelTimerOnInternalError(ctx sdk.Context) error {
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

// endBlockNoPanic runs endBlock and returns a panic as an error. A panic in
// an end block stops every node at that block, so it would halt the chain.
func (k Keeper) endBlockNoPanic(ctx sdk.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// The chain continues after this panic, so the log is the only
			// place that shows it. The stack gives the line that panicked,
			// as in the SDK's own recover for transactions.
			err = fmt.Errorf("panic: %v\nstack:\n%s", r, debug.Stack())
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
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return err
	}
	// Timer.Get returns a value, which cannot be nil, so the error tells the
	// two cases apart: nil means a timer is stored, ErrNotFound means none.
	running := err == nil
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
	// Here the proposal is in its voting period, its deadline is in the
	// future, and this block is exactly H-1.
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
	// H is the same for every nomination of this block.
	upgradeHeight, err := addBlocks(ctx.BlockHeight(), params.TimelockBlocks)
	if err != nil {
		return err
	}
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
		// The proposal has its own plan height, and at the final check in
		// block H-1 that height must not be in the past, or the timer is
		// dropped. A timer for a plan height below H-1 can never succeed. It
		// would only block other proposals for the whole timelock, so it is
		// not started.
		//
		// Example: block 100 and timelock 150 give H = 250, and a plan height
		// of 200 gets no timer, because 200 is below 249.
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

// closeVoteEarly is the final check at H-1. It runs every check before it
// writes anything to gov. If all checks pass, it sets the plan height to H
// and ends the voting period now. Gov's EndBlocker runs after this module and
// executes the proposal in this same block.
//
// Gov processes every proposal that is due in a block, one after the other.
// If gov executes another proposal before it tallies ours, that proposal can
// change the validators or gov's params, or schedule an upgrade plan. Gov's
// result for ours can then differ from the check here. So when another
// proposal is due in this block, closeVoteEarly does not close the vote. It
// moves H later and runs the final check again at the new H-1, after gov
// processed the other proposal.
func (k Keeper) closeVoteEarly(ctx sdk.Context, p govv1.Proposal, timer types.Timer, postpone uint64, threshold sdkmath.LegacyDec) error {
	msg, reason, err := k.guards(ctx, p, ctx.BlockHeight(), threshold)
	if err != nil {
		return err
	}
	if reason != types.ReasonOK {
		return k.cancelTimer(ctx, timer, reason)
	}
	due, err := k.anyProposalDue(ctx)
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

// anyProposalDue reports whether gov's EndBlocker ends a voting period in this
// block. It uses gov's own range over the queue. The caller's own proposal is
// not due yet, so a due proposal is always another one, and it postpones ours.
func (k Keeper) anyProposalDue(ctx sdk.Context) (bool, error) {
	iter, err := k.gov.ActiveProposalsQueue.Iterate(ctx, collections.NewPrefixUntilPairRange[time.Time, uint64](ctx.BlockTime()))
	if err != nil {
		return false, err
	}
	defer iter.Close()
	// Valid is true when the range has at least one proposal due.
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
