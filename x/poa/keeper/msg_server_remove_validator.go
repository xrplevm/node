package keeper

import (
	"context"

	"cosmossdk.io/errors"

	gov "github.com/cosmos/cosmos-sdk/x/gov/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/xrplevm/node/v7/x/poa/types"
)

func (k msgServer) RemoveValidator(goCtx context.Context, msg *types.MsgRemoveValidator) (*types.MsgRemoveValidatorResponse, error) {
	if k.authority != msg.Authority {
		return nil, errors.Wrapf(gov.ErrInvalidSigner, "expected %s got %s", k.authority, msg.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	err := k.ExecuteRemoveValidator(ctx, msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	return &types.MsgRemoveValidatorResponse{}, nil
}
