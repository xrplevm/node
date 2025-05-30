// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package network

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	feemarkettypes "github.com/evmos/evmos/v20/x/feemarket/types"
)

// Network is the interface that wraps the common methods to interact with integration test network.
//
// It was designed to avoid users to access module's keepers directly and force integration tests
// to be closer to the real user's behavior.
type Network interface {
	GetContext() sdktypes.Context
	GetChainID() string
	GetDenom() string
	GetOtherDenoms() []string
	GetValidators() []stakingtypes.Validator
	GetMinDepositAmt() sdkmath.Int
	NextBlock() error
	NextBlockAfter(duration time.Duration) error
	NextBlockWithTxs(txBytes ...[]byte) (*abcitypes.ResponseFinalizeBlock, error)

	// Clients
	GetAuthClient() authtypes.QueryClient
	GetAuthzClient() authz.QueryClient
	GetBankClient() banktypes.QueryClient
	GetStakingClient() stakingtypes.QueryClient
	GetDistrClient() distrtypes.QueryClient
	GetFeeMarketClient() feemarkettypes.QueryClient
	GetGovClient() govtypes.QueryClient

	BroadcastTxSync(txBytes []byte) (abcitypes.ExecTxResult, error)
	Simulate(txBytes []byte) (*txtypes.SimulateResponse, error)
	CheckTx(txBytes []byte) (*abcitypes.ResponseCheckTx, error)

	// GetIBCChain returns the IBC test chain.
	// NOTE: this is only used for testing IBC related functionality.
	// The idea is to deprecate this eventually.
	GetIBCChain(t *testing.T, coord *ibctesting.Coordinator) *ibctesting.TestChain
	GetEncodingConfig() sdktestutil.TestEncodingConfig
}
