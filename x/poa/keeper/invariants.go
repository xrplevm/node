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

// AllInvariants runs all invariants of the module
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Add here module invariants
		return StakingPowerInvariant(k)(ctx)
		
	}
}

// StakingPowerInvariant checks that all validators have the same
// staking power as the default power reduction. If not, it returns an invariant error.
func StakingPowerInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg string
			broken bool
		)

		k.bk.IterateAllBalances(ctx, 
			func(address sdk.AccAddress, coin sdk.Coin) (stop bool) {
				if coin.Denom == k.sk.GetParams(ctx).BondDenom {
					validator, found := k.sk.GetValidator(ctx, sdk.ValAddress(address))
					if !found {
						return false
					}

					if !validator.Tokens.Equal(sdk.DefaultPowerReduction) {
						msg = fmt.Sprintf("validator %s has %s tokens, not %s", 
							validator.GetOperator().String(),
							validator.Tokens.String(),
							sdk.DefaultPowerReduction.String(),
						)
						broken = true
						return true
					}
				}

				return false
			},
		)

		return sdk.FormatInvariant(
			types.ModuleName,
			"staking-power-invariant",
			fmt.Sprintf("excessive staking power for account %s", msg),
		), broken
	}
}

// SelfDelegationInvariant checks that all validators have only one self-delegation.
// Each delegation has to match the delegator and 
func SelfDelegationInvariant(k Keeper) sdk.Invariant {
	return func (ctx sdk.Context) (string, bool) {
		var (
			msg string
			broken bool
		)

		k.bk.IterateAllBalances(ctx, 
			func(address sdk.AccAddress, coin sdk.Coin) (stop bool) {
				valAddress := sdk.ValAddress(address)
				_, found := k.sk.GetValidator(ctx, valAddress)
				if !found {
					return false
				}

				delegations := k.sk.GetAllDelegatorDelegations(ctx, address)

				nDelegations := len(delegations)

				if nDelegations != 1 {
					broken = true
					msg = fmt.Sprintf("invalid number of validator %s delegations, got: %d, expected: 1", address.String(), nDelegations)
					return true
				}

				for _, delegation := range delegations {
					_, found := k.sk.GetValidator(ctx, delegation.GetValidatorAddr())
					if !delegation.GetValidatorAddr().Equals(valAddress) {
						broken = true
						msg = fmt.Sprintf("validator %s and delegator validator %s do not match", address, delegation.GetValidatorAddr())
						return true
					} else if delAddress := sdk.ValAddress(delegation.DelegatorAddress); !delAddress.Equals(valAddress) {
						broken = true
						msg = fmt.Sprintf("validator %s and delegator %s addresses do not match", delAddress, valAddress)
						return true
					} else if !found {
						broken = false
						msg = fmt.Sprintf("validator %s has no self delegation", address.String())
						return true
					}
				}
				return false
			},
		)

		return sdk.FormatInvariant(
			types.ModuleName,
			"self-delegation-invariant",
			fmt.Sprintf("invalid validator self-delegation %s", msg),
		), broken
	}
}