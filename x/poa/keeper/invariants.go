package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/xrplevm/node/v3/x/poa/types"
)

// RegisterInvariants registers all module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "staking-power", StakingPowerInvariant(k))
	ir.RegisterRoute(types.ModuleName, "self-delegation", SelfDelegationInvariant(k))
}

// StakingPowerInvariant checks that all validators have the same
// staking power as the default power reduction. If not, it returns an invariant error.
// TODO: Check validation
func StakingPowerInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg    string
			broken bool
		)

		validators := k.sk.GetAllValidators(ctx)

		for _, validator := range validators {
			if !validator.Tokens.Equal(sdk.DefaultPowerReduction) && !validator.Tokens.IsZero() {
				msg = fmt.Sprintf("excessive staking power for account %s: %s", validator.GetOperator().String(), validator.Tokens.String())
				broken = true
				break
			}
		}

		return sdk.FormatInvariant(
			types.ModuleName,
			"staking-power-invariant",
			fmt.Sprintf("excessive staking power for account %s", msg),
		), broken
	}
}

// SelfDelegationInvariant checks that all validators have only one self-delegation.
// Each delegation has to match the delegator and validator address.
func SelfDelegationInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg    string
			broken bool
		)

		delegations := k.sk.GetAllDelegations(ctx)
		for _, delegation := range delegations {
			// TODO: Check delegation address conversion
			if !delegation.GetValidatorAddr().Equals(sdk.ValAddress(delegation.GetDelegatorAddr())) {
				msg = fmt.Sprintf("validator address %s and delegation address do not match %s", sdk.ValAddress(delegation.GetDelegatorAddr()), delegation.GetValidatorAddr())
				broken = true
				break
			}
		}

		return sdk.FormatInvariant(
			types.ModuleName,
			"self-delegation-invariant",
			fmt.Sprintf("invalid validator self-delegation %s", msg),
		), broken
	}
}

func CheckSlashingParamsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg    string
			broken bool
		)
		
		params := k.ck.GetParams(ctx)

		if !(params.SlashFractionDoubleSign.IsZero() && params.SlashFractionDowntime.IsZero()) {
			msg = fmt.Sprintf(
				"slashing params are not zero: slash_fraction_double_sign %s, slash_fraction_downtime %s",
				params.SlashFractionDoubleSign.String(),
				params.SlashFractionDowntime.String(),
			)
			broken = true
		}

		return sdk.FormatInvariant(
			types.ModuleName,
			"slashing-params-invariant",
			fmt.Sprintf("slashing params are not zero: %s", msg),
		), broken
	}
}