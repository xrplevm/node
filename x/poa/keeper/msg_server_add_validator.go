package keeper

import (
	"context"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/xrplevm/node/v6/x/poa/types"
)

func (k msgServer) AddValidator(goCtx context.Context, msg *types.MsgAddValidator) (*types.MsgAddValidatorResponse, error) {
	if k.authority != msg.Authority {
		return nil, errors.Wrapf(gov.ErrInvalidSigner, "expected %s got %s", k.authority, msg.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	err := k.ExecuteAddValidator(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgAddValidatorResponse{}, nil
}
