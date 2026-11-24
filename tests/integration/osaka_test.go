//go:build test

package integration

import (
	"encoding/json"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/evm/testutil/integration/base/factory"
	"github.com/cosmos/evm/testutil/keyring"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethparams "github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v10/app"
	exrpcommon "github.com/xrplevm/node/v10/testutil/integration/exrp/common"
	exrpintegration "github.com/xrplevm/node/v10/testutil/integration/exrp/integration"
)

// clzInitCode is PUSH1 1, CLZ, STOP. CLZ (EIP-7939) is an Osaka opcode: before
// Osaka the execution fails with "invalid opcode".
var clzInitCode = common.FromHex("60011e00")

// newUnitNetwork creates the SetupTest network with access to the app.
func newUnitNetwork(kr keyring.Keyring) *exrpintegration.UnitTestIntegrationNetwork {
	evmGen := evmtypes.DefaultGenesisState()
	evmGen.Params.EvmDenom = app.BaseDenom
	return exrpintegration.NewUnitTestNetwork(
		exrpcommon.WithPreFundedAccounts(kr.GetAllAccAddrs()...),
		exrpcommon.WithAmountOfValidators(5),
		exrpcommon.WithCustomGenesis(exrpcommon.CustomGenesisState{evmtypes.ModuleName: evmGen}),
		exrpcommon.WithBondDenom("apoa"),
		exrpcommon.WithMaxValidators(7),
		exrpcommon.WithMinDepositAmt(sdkmath.NewInt(1)),
		exrpcommon.WithValidatorOperators(kr.GetAllAccAddrs()),
	)
}

// A chain without a scheduled activation time runs Osaka from genesis.
func (s *TestSuite) TestOsaka_Opcode() {
	t := s.T()

	res, err := s.factory.ExecuteEthTx(s.keyring.GetPrivKey(0), evmtypes.EvmTxArgs{Input: clzInitCode})
	require.NoError(t, err)
	require.Zero(t, res.Code, res.Log)
	txRes, err := s.factory.GetEvmTransactionResponseFromTxResult(res)
	require.NoError(t, err)
	require.Empty(t, txRes.VmError)
}

// The precompiles have the Osaka prices.
func (s *TestSuite) TestOsaka_PrecompilePrices() {
	tt := []struct {
		name    string
		address common.Address
		price   uint64
	}{
		{
			name:    "MODEXP minimum (EIP-7883)",
			address: common.BytesToAddress([]byte{0x05}),
			price:   500,
		},
		{
			name:    "P256VERIFY (EIP-7951)",
			address: common.BytesToAddress([]byte{0x01, 0x00}),
			price:   6900,
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			t := s.T()
			callWithGas := func(gas uint64) string {
				args, err := json.Marshal(evmtypes.TransactionArgs{To: &tc.address, Gas: (*hexutil.Uint64)(&gas)})
				require.NoError(t, err)
				res, err := s.network.GetEvmClient().EthCall(
					s.network.GetContext(),
					&evmtypes.EthCallRequest{Args: args, GasCap: 25_000_000},
				)
				require.NoError(t, err)
				return res.VmError
			}

			// A call without input needs exactly the transaction gas and the price.
			require.Empty(t, callWithGas(ethparams.TxGas+tc.price))
			require.Contains(t, callWithGas(ethparams.TxGas+tc.price-1), "out of gas")
		})
	}
}

// P256VERIFY is an Ethereum precompile since Osaka and no longer an EVM
// extension. The app blocks the Osaka precompile addresses, so its address
// still cannot receive funds.
func (s *TestSuite) TestOsaka_P256BlockedRecipient() {
	t := s.T()
	sender := s.keyring.GetKey(0)
	p256 := sdk.AccAddress(common.BytesToAddress([]byte{0x01, 0x00}).Bytes())
	coins := sdk.NewCoins(sdk.NewCoin(s.network.GetBaseDenom(), sdkmath.OneInt()))
	gas := uint64(200_000)

	res, err := s.factory.CommitCosmosTx(sender.Priv, factory.CosmosTxArgs{
		Msgs: []sdk.Msg{banktypes.NewMsgSend(sender.AccAddr, p256, coins)},
		Gas:  &gas,
	})
	require.NoError(t, err)
	require.Equal(t, sdkerrors.ErrUnauthorized.Codespace(), res.Codespace, res.Log)
	require.Equal(t, sdkerrors.ErrUnauthorized.ABCICode(), res.Code, res.Log)
	require.Contains(t, res.Log, "not allowed to receive funds")
}

// The live networks have blocks from before Osaka, so they activate it at a
// scheduled time and not from genesis.
func (s *TestSuite) TestOsaka_ActivationTime() {
	nw := newUnitNetwork(keyring.New(5))
	args, err := json.Marshal(evmtypes.TransactionArgs{Input: (*hexutil.Bytes)(&clzInitCode)})
	s.Require().NoError(err)

	tt := []struct {
		name    string
		chainID string
	}{
		{
			name:    "testnet",
			chainID: "xrplevm_1449000-1",
		},
		{
			name:    "mainnet",
			chainID: "xrplevm_1440000-1",
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			t := s.T()
			clzAt := func(blockTime uint64) string {
				res, err := nw.App.EvmKeeper.EthCall(
					nw.GetContext().WithBlockTime(time.Unix(int64(blockTime), 0)), //nolint:gosec // G115
					&evmtypes.EthCallRequest{Args: args, GasCap: 25_000_000},
				)
				require.NoError(t, err)
				return res.VmError
			}

			// Create the app again with the chain ID of the network, so that
			// app.New sets its Osaka time.
			evmtypes.NewEVMConfigurator().ResetTestConfig()
			exrpcommon.CreateExrpApp(tc.chainID)

			activation := evmtypes.GetEthChainConfig().OsakaTime
			require.NotNil(t, activation)
			require.NotZero(t, *activation)
			require.Contains(t, clzAt(*activation-1), "invalid opcode")
			require.Empty(t, clzAt(*activation))
		})
	}
}
