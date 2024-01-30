package poa

import (
	"github.com/Peersyst/exrp/x/poa/keeper"
	"github.com/Peersyst/exrp/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// NewValidatorProposalHandler creates a new governance Handler for poa proposals
func NewValidatorProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.AddValidatorProposal:
			return handleAddValidatorProposal(ctx, k, c)
		case *types.RemoveValidatorProposal:
			return handleRemoveValidatorProposal(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized param proposal content type: %T", c)
		}
	}
}

func handleAddValidatorProposal(ctx sdk.Context, k keeper.Keeper, p *types.AddValidatorProposal) error {
	return nil
	// TODO: What this?
	// return k.ExecuteAddValidator(ctx, p)
}

func handleRemoveValidatorProposal(ctx sdk.Context, k keeper.Keeper, p *types.RemoveValidatorProposal) error {
	return k.ExecuteRemoveValidator(ctx, p.ValidatorAddress)
}
