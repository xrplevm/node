package app_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	legacytypes "github.com/cosmos/evm/rpc/types/legacy"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v12/app"
	"github.com/xrplevm/node/v12/testutil/constants"
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

// x/authz was removed in v12. Blocks from before v12 can contain authz txs, and
// RPC queries must still decode them.
func TestMakeEncodingConfig_DecodesRemovedAuthzMsgs(t *testing.T) {
	granter := sdk.AccAddress("granter_____________")
	grantee := sdk.AccAddress("grantee_____________")
	expiration := time.Unix(2_000_000_000, 0).UTC()
	grant, err := authz.NewMsgGrant(granter, grantee, authz.NewGenericAuthorization("/cosmos.bank.v1beta1.MsgSend"), &expiration)
	require.NoError(t, err)
	revoke := authz.NewMsgRevoke(granter, grantee, "/cosmos.bank.v1beta1.MsgSend")
	exec := authz.NewMsgExec(grantee, []sdk.Msg{&revoke})

	encodingConfig := app.MakeEncodingConfig(constants.LocalnetChainID.EVMChainID)
	for _, msg := range []sdk.Msg{grant, &revoke, &exec} {
		txBuilder := encodingConfig.TxConfig.NewTxBuilder()
		require.NoError(t, txBuilder.SetMsgs(msg))
		bz, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
		require.NoError(t, err)

		tx, err := encodingConfig.TxConfig.TxDecoder()(bz)
		require.NoError(t, err, sdk.MsgTypeURL(msg))
		require.Equal(t, []sdk.Msg{msg}, tx.GetMsgs())
	}
}

// ICA host was removed in v12. Blocks from before v12 can contain ICA host txs, and
// interchain accounts it created stay in the auth store, so both must still decode.
func TestMakeEncodingConfig_DecodesRemovedICAHostTypes(t *testing.T) {
	signer := sdk.AccAddress("signer______________").String()
	updateParams := icahosttypes.NewMsgUpdateParams(signer, icahosttypes.NewParams(false, nil))
	querySafe := icahosttypes.NewMsgModuleQuerySafe(signer, []icahosttypes.QueryRequest{{
		Path: "/cosmos.bank.v1beta1.Query/AllBalances",
		Data: []byte{0x0a, 0x01, 0x01},
	}})

	encodingConfig := app.MakeEncodingConfig(constants.LocalnetChainID.EVMChainID)
	for _, msg := range []sdk.Msg{updateParams, querySafe} {
		txBuilder := encodingConfig.TxConfig.NewTxBuilder()
		require.NoError(t, txBuilder.SetMsgs(msg))
		bz, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
		require.NoError(t, err)

		tx, err := encodingConfig.TxConfig.TxDecoder()(bz)
		require.NoError(t, err, sdk.MsgTypeURL(msg))
		require.Equal(t, []sdk.Msg{msg}, tx.GetMsgs())
	}

	var account sdk.AccountI = icatypes.NewInterchainAccount(
		authtypes.NewBaseAccountWithAddress(sdk.AccAddress("interchain_account__")),
		"ethm1owner",
	)
	accountBz, err := encodingConfig.Codec.MarshalInterface(account)
	require.NoError(t, err)
	var decoded sdk.AccountI
	require.NoError(t, encodingConfig.Codec.UnmarshalInterface(accountBz, &decoded))
	require.Equal(t, account, decoded)
}
