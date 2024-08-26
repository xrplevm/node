package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/xrplevm/node/v3/testutil/sims"
	"github.com/xrplevm/node/v3/x/poa/types"
)

// FindAccount find a specific address from an account list
func FindAccount(accs []simtypes.Account, address string) (simtypes.Account, bool) {
	creator, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		panic(err)
	}
	return simtypes.FindAccount(accs, creator)
}

func RandomAccount(r *rand.Rand, accs []sdk.Address) sdk.Address {
	acc := accs[simtypes.RandIntBetween(r, 0, len(accs))]
	return acc
}

func DeliverMsg(msg sdk.Msg, sender simtypes.Account, bk types.BankKeeper, ak types.AccountKeeper, app *baseapp.BaseApp, r *rand.Rand, ctx sdk.Context, cdc *codec.ProtoCodec, chainID string) error {
	txCfg := tx.NewTxConfig(cdc, tx.DefaultSignModes)
	senderAcc := ak.GetAccount(ctx, sender.Address)
	spendableCoins := bk.SpendableCoins(ctx, sender.Address)
	fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
	if err != nil {
		return err
	}
	transaction, err := sims.GenSignedMockTx(
		r,
		txCfg,
		[]sdk.Msg{msg},
		fees,
		sims.DefaultGenTxGas,
		chainID,
		[]uint64{senderAcc.GetAccountNumber()},
		[]uint64{senderAcc.GetSequence()},
		sender.PrivKey,
	)
	if err != nil {
		return err
	}
	if _, _, err = app.SimDeliver(txCfg.TxEncoder(), transaction); err != nil {
		return err
	}
	return nil
}
