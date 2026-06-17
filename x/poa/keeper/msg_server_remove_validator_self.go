package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/xrplevm/node/v10/x/poa/types"
)

func (k msgServer) RemoveValidatorSelf(goCtx context.Context, msg *types.MsgRemoveValidatorSelf) (*types.MsgRemoveValidatorSelfResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := k.ExecuteRemoveValidatorSelf(ctx, msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	return &types.MsgRemoveValidatorSelfResponse{}, nil
}
