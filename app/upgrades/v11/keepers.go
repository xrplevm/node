package v11

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
)

// ICAHostKeeper is the narrow interface required by the v11 upgrade
// handler. It matches a subset of icahostkeeper.Keeper.
type ICAHostKeeper interface {
	SetParams(ctx sdk.Context, params icahosttypes.Params)
}

// BankKeeper is the narrow interface required by the v11 upgrade
// handler. It matches a subset of bankkeeper.Keeper.
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
}

// TransferKeeper is the narrow interface required by the v11 upgrade
// handler. It matches a subset of transferkeeper.Keeper.
type TransferKeeper interface {
	GetTotalEscrowForDenom(ctx sdk.Context, denom string) sdk.Coin
	SetTotalEscrowForDenom(ctx sdk.Context, coin sdk.Coin)
}
