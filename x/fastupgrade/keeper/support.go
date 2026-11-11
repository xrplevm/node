package keeper

import (
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/xrplevm/node/v12/x/fastupgrade/types"
)

// eligibleMinorMessage returns the proposal's upgrade message when the
// proposal holds exactly one MsgSoftwareUpgrade whose plan is valid and not
// below minPlanHeight, whose plan name is a strict vMAJOR.MINOR.PATCH with
// the same major and a higher minor than the last completed upgrade, and
// whose plan name has not already completed. It returns nil when the proposal
// is not eligible.
func (k Keeper) eligibleMinorMessage(ctx sdk.Context, p govv1.Proposal, minPlanHeight int64) (*upgradetypes.MsgSoftwareUpgrade, error) {
	if len(p.Messages) != 1 {
		return nil, nil
	}
	msg, ok := p.Messages[0].GetCachedValue().(*upgradetypes.MsgSoftwareUpgrade)
	if !ok {
		return nil, nil
	}
	// An invalid plan is not an error here: stock gov decides the proposal.
	//nolint:nilerr
	if msg.Plan.ValidateBasic() != nil {
		return nil, nil
	}
	// x/upgrade refuses a plan below the current height, so stock gov would
	// fail this proposal. minPlanHeight is never below the current height.
	if msg.Plan.Height < minPlanHeight {
		return nil, nil
	}
	target, ok := types.ParseVersion(msg.Plan.Name)
	if !ok {
		return nil, nil
	}
	lastName, _, err := k.upgrade.GetLastCompletedUpgrade(ctx)
	if err != nil {
		return nil, err
	}
	base, ok := types.ParseVersion(lastName)
	if !ok || !target.IsMinorBumpOf(base) {
		return nil, nil
	}
	doneHeight, err := k.upgrade.GetDoneHeight(ctx, msg.Plan.Name)
	if err != nil {
		return nil, err
	}
	if doneHeight != 0 {
		return nil, nil
	}
	return msg, nil
}

// votingProposal returns the proposal when it exists and is in its voting period.
func (k Keeper) votingProposal(ctx sdk.Context, id uint64) (govv1.Proposal, bool, error) {
	p, err := k.gov.Proposals.Get(ctx, id)
	if errors.Is(err, collections.ErrNotFound) {
		return govv1.Proposal{}, false, nil
	}
	if err != nil {
		return govv1.Proposal{}, false, err
	}
	return p, p.Status == govv1.StatusVotingPeriod, nil
}

func (k Keeper) upgradePlanExists(ctx sdk.Context) (bool, error) {
	_, err := k.upgrade.GetUpgradePlan(ctx)
	if errors.Is(err, upgradetypes.ErrNoUpgradePlanFound) {
		return false, nil
	}
	return err == nil, err
}

// checkGovCompatibility checks threshold against gov's current params.
func (k Keeper) checkGovCompatibility(ctx sdk.Context, threshold math.LegacyDec) error {
	govParams, err := k.gov.Params.Get(ctx)
	if err != nil {
		return err
	}
	return types.CheckGovCompatibility(threshold, govParams)
}

// yesWeight sums the Yes weights of a vote. A plain Yes vote weighs 1.
func yesWeight(vote govv1.Vote) (math.LegacyDec, error) {
	weight := math.LegacyZeroDec()
	for _, option := range vote.Options {
		if option.Option != govv1.OptionYes {
			continue
		}
		w, err := math.LegacyNewDecFromStr(option.Weight)
		if err != nil {
			return math.LegacyDec{}, err
		}
		weight = weight.Add(w)
	}
	return weight, nil
}

// meetsThreshold reports yes >= threshold * bonded. yes is a sum of
// weight.MulInt(tokens) terms, and threshold.MulInt(bonded) is exact, so the
// comparison has no rounding. An empty bonded pool never meets the threshold.
func meetsThreshold(yes, threshold math.LegacyDec, bonded math.Int) bool {
	return bonded.IsPositive() && yes.GTE(threshold.MulInt(bonded))
}

// validatorGate reports whether bonded validators voting Yes hold at least
// threshold of the bonded pool. It reads one vote per bonded validator.
func (k Keeper) validatorGate(ctx sdk.Context, proposalID uint64, threshold math.LegacyDec) (bool, error) {
	bonded, err := k.staking.TotalBondedTokens(ctx)
	if err != nil {
		return false, err
	}
	yes := math.LegacyZeroDec()
	var walkErr error
	err = k.staking.IterateBondedValidatorsByPower(ctx, func(_ int64, val stakingtypes.ValidatorI) bool {
		operator, err := k.staking.ValidatorAddressCodec().StringToBytes(val.GetOperator())
		if err != nil {
			walkErr = err
			return true
		}
		vote, err := k.gov.Votes.Get(ctx, collections.Join(proposalID, sdk.AccAddress(operator)))
		if errors.Is(err, collections.ErrNotFound) {
			return false
		}
		if err != nil {
			walkErr = err
			return true
		}
		weight, err := yesWeight(vote)
		if err != nil {
			walkErr = err
			return true
		}
		yes = yes.Add(weight.MulInt(val.GetBondedTokens()))
		return false
	})
	if err != nil {
		return false, err
	}
	if walkErr != nil {
		return false, walkErr
	}
	return meetsThreshold(yes, threshold, bonded), nil
}

// supportCheck is the support check used at timer start and at H-1: the cheap
// validator gate first, the full tally last.
func (k Keeper) supportCheck(ctx sdk.Context, p govv1.Proposal, threshold math.LegacyDec) (types.Reason, error) {
	ready, err := k.validatorGate(ctx, p.Id, threshold)
	if err != nil {
		return "", err
	}
	if !ready {
		return types.ReasonGateBelowThreshold, nil
	}
	err = k.checkGovCompatibility(ctx, threshold)
	if errors.Is(err, types.ErrIncompatibleWithGov) {
		k.Logger(ctx).Info("fast upgrade threshold is incompatible with gov params", "proposal", p.Id, "error", err)
		return types.ReasonIncompatibleWithGov, nil
	}
	if err != nil {
		return "", err
	}
	return k.tallyCheck(ctx, p, threshold)
}

// tallyCheck runs gov's stock Tally on a scratch copy of state that is never
// written back: Tally deletes every vote it counts.
func (k Keeper) tallyCheck(ctx sdk.Context, p govv1.Proposal, threshold math.LegacyDec) (types.Reason, error) {
	scratch, _ := ctx.CacheContext()
	passes, _, result, err := k.gov.Tally(scratch, p)
	if err != nil {
		return "", err
	}
	if !passes {
		return types.ReasonGovTallyRejected, nil
	}
	yes, ok := math.NewIntFromString(result.YesCount)
	if !ok {
		return "", fmt.Errorf("gov yes count %q is not an integer", result.YesCount)
	}
	bonded, err := k.staking.TotalBondedTokens(ctx)
	if err != nil {
		return "", err
	}
	if !meetsThreshold(math.LegacyNewDecFromInt(yes), threshold, bonded) {
		return types.ReasonTallyBelowThreshold, nil
	}
	return types.ReasonOK, nil
}
