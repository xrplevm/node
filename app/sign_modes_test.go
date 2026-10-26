package app_test

import (
	"context"
	"testing"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txsigning "cosmossdk.io/x/tx/signing"
	"github.com/stretchr/testify/require"
)

func TestAppSignModeHandler(t *testing.T) {
	handler := sharedTestApp().TxConfig().SignModeHandler()

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
