package app_test

import (
	"context"
	"testing"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/log"
	txsigning "cosmossdk.io/x/tx/signing"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/stretchr/testify/require"

	"github.com/xrplevm/node/v10/app"
)

func TestAppSignModeHandler(t *testing.T) {
	// app.New sets a process-wide EVM chain config. Only the default EVM chain ID
	// lets a later test in this package (TestFullAppSimulation) set it again.
	bApp := app.New(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		false,
		map[int64]bool{},
		0,
		evmtypes.DefaultEVMChainID,
		simtestutil.AppOptionsMap{flags.FlagHome: t.TempDir()},
		baseapp.SetChainID("signmode_262144-1"),
	)
	handler := bApp.TxConfig().SignModeHandler()

	require.Equal(t, signingv1beta1.SignMode_SIGN_MODE_DIRECT, handler.DefaultMode())
	require.Equal(t, []signingv1beta1.SignMode{
		signingv1beta1.SignMode_SIGN_MODE_DIRECT,
		signingv1beta1.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
	}, handler.SupportedModes())

	_, err := handler.GetSignBytes(
		context.Background(),
		signingv1beta1.SignMode_SIGN_MODE_DIRECT_AUX,
		txsigning.SignerData{},
		txsigning.TxData{},
	)
	require.ErrorContains(t, err, "unsuppored sign mode SIGN_MODE_DIRECT_AUX")
}
