package v11

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
)

// ICAHostKeeper is the narrow interface required by the v11 upgrade
// handler. It matches a subset of icahostkeeper.Keeper.
type ICAHostKeeper interface {
	SetParams(ctx sdk.Context, params icahosttypes.Params)
}

// StakingKeeper is the narrow interface required by the v11 upgrade
// handler. It matches a subset of staking.Keeper.
type StakingKeeper interface {
	GetParams(ctx context.Context) (params stakingtypes.Params, err error)
	SetParams(ctx context.Context, params stakingtypes.Params) error
}

// BankKeeper is the narrow interface required by the v11 upgrade
// handler. It matches a subset of bankkeeper.Keeper.
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// TransferKeeper is the narrow interface required by the v11 upgrade
// handler. It matches a subset of transferkeeper.Keeper.
type TransferKeeper interface {
	IterateTokensInEscrow(ctx sdk.Context, storeprefix []byte, cb func(denomEscrow sdk.Coin) bool)
	UnescrowCoin(ctx sdk.Context, escrowAddress, receiver sdk.AccAddress, coin sdk.Coin) error
}
