package app_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	legacytypes "github.com/cosmos/evm/rpc/types/legacy"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v10/app"
	"github.com/xrplevm/node/v10/testutil/constants"
)

// Pre-v9 blocks store Ethereum txs as ethermint.evm.v1 messages. Decoding one
// with a non-empty access list resolves ethermint.evm.v1.AccessTuple by name.
func TestMakeEncodingConfig_DecodesLegacyTxWithAccessList(t *testing.T) {
	evmChainID := constants.LocalnetChainID.EVMChainID
	chainID := sdkmath.NewIntFromUint64(evmChainID)
	zero := sdkmath.ZeroInt()
	accessList := legacytypes.AccessList{{
		Address:     "0x0000000000000000000000000000000000000001",
		StorageKeys: []string{"0x0000000000000000000000000000000000000000000000000000000000000000"},
	}}

	data, err := legacytypes.PackTxData(&legacytypes.AccessListTx{
		ChainID:  &chainID,
		Amount:   &zero,
		GasPrice: &zero,
		Accesses: accessList,
	})
	require.NoError(t, err)

	encodingConfig := app.MakeEncodingConfig(evmChainID)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	err = txBuilder.SetMsgs(&legacytypes.MsgEthereumTx{Data: data})
	require.NoError(t, err)

	bz, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	require.NoError(t, err)

	tx, err := encodingConfig.TxConfig.TxDecoder()(bz)
	require.NoError(t, err)

	msgs := tx.GetMsgs()
	require.Len(t, msgs, 1)
	msg, ok := msgs[0].(*legacytypes.MsgEthereumTx)
	require.True(t, ok)
	txData, err := legacytypes.UnpackTxData(msg.Data)
	require.NoError(t, err)
	require.Equal(t, *accessList.ToEthAccessList(), txData.GetAccessList())
}
