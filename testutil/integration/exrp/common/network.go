package exrpcommon

import (
	"testing"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type Network interface {
	// Keepers
	NetworkKeepers

	// Clients
	BroadcastTxSync(txBytes []byte) (abcitypes.ExecTxResult, error)
	Simulate(txBytes []byte) (*txtypes.SimulateResponse, error)
	CheckTx(txBytes []byte) (*abcitypes.ResponseCheckTx, error)

	// GetIBCChain returns the IBC test chain.
	// NOTE: this is only used for testing IBC related functionality.
	// The idea is to deprecate this eventually.
	GetIBCChain(t *testing.T, coord *ibctesting.Coordinator) *ibctesting.TestChain
	GetEncodingConfig() sdktestutil.TestEncodingConfig

	// Getters
	GetContext() sdktypes.Context
	GetChainID() string
	GetDenom() string
	GetOtherDenoms() []string
	GetValidators() []stakingtypes.Validator

	// ABCI
	NextBlock() error
	NextBlockAfter(duration time.Duration) error
	NextBlockWithTxs(txBytes ...[]byte) (*abcitypes.ResponseFinalizeBlock, error)
}
