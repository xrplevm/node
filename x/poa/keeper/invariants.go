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
	ir.RegisterRoute(types.ModuleName, "keeper-dependencies-params", CheckKeeperDependenciesParamsInvariant(k))
}

// StakingPowerInvariant checks that all validators have the same
// staking power as the default power reduction. If not, it returns an invariant error.
func StakingPowerInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg    string
			broken bool
		)

		validators, err := k.sk.GetAllValidators(ctx)
		if err != nil {
			panic(err)
		}

		for _, validator := range validators {
			if !validator.Tokens.Equal(sdk.DefaultPowerReduction) && !validator.Tokens.IsZero() {
				msg = fmt.Sprintf("excessive staking power for account %s: %s", validator.GetOperator(), validator.Tokens.String())
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

		delegations, err := k.sk.GetAllDelegations(ctx)
		if err != nil {
			panic(err)
		}
		for _, delegation := range delegations {
			validatorAddress, err := sdk.ValAddressFromBech32(delegation.GetValidatorAddr())
			if err != nil {
				panic(err)
			}
			delegatorAddress, err := sdk.AccAddressFromBech32(delegation.GetDelegatorAddr())
			if err != nil {
				panic(err)
			}
			if !sdk.AccAddress(validatorAddress).Equals(delegatorAddress) {
				msg = fmt.Sprintf("validator address %s and delegation address do not match %s", sdk.AccAddress(validatorAddress), delegatorAddress)
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

// CheckKeeperDependenciesParamsInvariant checks that keeper dependencies params
// are set to the expected values.
// This is to ensure that the keeper dependencies are correctly initialized.
// Slashing params SlashFractionDoubleSign and SlashFractionDowntime should be zero.
func CheckKeeperDependenciesParamsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg    string
			broken bool
		)

		params, err := k.ck.GetParams(ctx)
		if err != nil {
			panic(err)
		}

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
