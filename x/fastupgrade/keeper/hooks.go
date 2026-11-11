package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Hooks wraps the keeper to implement gov hooks.
type Hooks struct {
	k Keeper
}

var _ govtypes.GovHooks = Hooks{}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// AfterProposalVote notes the proposal for this block's EndBlock check. It
// never rejects the vote: errors are only logged.
//
// The EVM gov precompile calls it on the EVM's snapshot store, which holds
// only KV stores and whose CacheContext is not a separate copy. So nominate
// writes only to the module's KV store and runs no scratch work.
func (h Hooks) AfterProposalVote(goCtx context.Context, proposalID uint64, voter sdk.AccAddress) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := h.k.nominate(ctx, proposalID, voter); err != nil {
		h.k.Logger(ctx).Error("failed to nominate proposal", "proposal", proposalID, "voter", voter.String(), "error", err)
	}
	return nil
}

func (h Hooks) AfterProposalSubmission(context.Context, uint64) error { return nil }

func (h Hooks) AfterProposalDeposit(context.Context, uint64, sdk.AccAddress) error { return nil }

func (h Hooks) AfterProposalFailedMinDeposit(context.Context, uint64) error { return nil }

func (h Hooks) AfterProposalVotingPeriodEnded(context.Context, uint64) error { return nil }

// nominate adds the proposal to this block's nominations when a bonded
// validator voted Yes on an eligible minor upgrade and no timer is running.
func (k Keeper) nominate(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) error {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}
	if !params.Enabled {
		return nil
	}

	running, err := k.Timer.Has(ctx)
	if err != nil || running {
		return err
	}

	validator, err := k.staking.GetValidator(ctx, sdk.ValAddress(voter))
	if errors.Is(err, stakingtypes.ErrNoValidatorFound) {
		return nil
	}
	if err != nil || !validator.IsBonded() {
		return err
	}

	vote, err := k.gov.Votes.Get(ctx, collections.Join(proposalID, voter))
	if err != nil {
		return err
	}
	weight, err := yesWeight(vote)
	if err != nil || !weight.IsPositive() {
		return err
	}

	p, voting, err := k.votingProposal(ctx, proposalID)
	if err != nil || !voting {
		return err
	}
	msg, err := k.eligibleMinorMessage(ctx, p, ctx.BlockHeight())
	if err != nil || msg == nil {
		return err
	}

	return k.Nominations.Set(ctx, proposalID)
}
