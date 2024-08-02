package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/xrplevm/node/v2/x/poa/types"
)

func SimulateMsgAddValidator() simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgAddValidator{
			ValidatorAddress: simAccount.Address.String(),
		}

		// TODO: Handling the AddValidator simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "AddValidator simulation not implemented"), nil, nil
	}
}
