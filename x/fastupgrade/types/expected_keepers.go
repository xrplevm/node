package types

import (
	"context"

	"cosmossdk.io/core/address"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// StakingKeeper is the staking state the module reads.
type StakingKeeper interface {
	TotalBondedTokens(ctx context.Context) (math.Int, error)
	IterateBondedValidatorsByPower(ctx context.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool)) error
	GetValidator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error)
	ValidatorAddressCodec() address.Codec
}

// UpgradeKeeper is the x/upgrade state the module reads.
type UpgradeKeeper interface {
	GetLastCompletedUpgrade(ctx context.Context) (string, int64, error)
	GetUpgradePlan(ctx context.Context) (upgradetypes.Plan, error)
	GetDoneHeight(ctx context.Context, name string) (int64, error)
}
