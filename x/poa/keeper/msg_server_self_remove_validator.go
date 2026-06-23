package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/xrplevm/node/v10/x/poa/types"
)

func (k msgServer) SelfRemoveValidator(goCtx context.Context, msg *types.MsgSelfRemoveValidator) (*types.MsgSelfRemoveValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := k.ExecuteSelfRemoveValidator(ctx, msg.Address)
	if err != nil {
		return nil, err
	}

	return &types.MsgSelfRemoveValidatorResponse{}, nil
}
