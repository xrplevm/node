// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package exrpupgrade

import (
	"fmt"
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/xrplevm/node/v5/app"
	exrpcommon "github.com/xrplevm/node/v5/testutil/integration/exrp/common"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	ed25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cometbft/cometbft/version"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Network is the interface that wraps the methods to interact with integration test network.
//
// It was designed to avoid users to access module's keepers directly and force integration tests
// to be closer to the real user's behavior.
type Network interface {
	exrpcommon.Network

	GetEIP155ChainID() *big.Int
}

var _ Network = (*UpgradeIntegrationNetwork)(nil)

// UpgradeIntegrationNetwork is the implementation of the Network interface for integration tests.
type UpgradeIntegrationNetwork struct {
	cfg        exrpcommon.Config
	ctx        sdktypes.Context
	validators []stakingtypes.Validator
	app        *app.App

	// This is only needed for IBC chain testing setup
	valSet     *cmttypes.ValidatorSet
	valSigners map[string]cmttypes.PrivValidator
}

// New configures and initializes a new integration Network instance with
// the given configuration options. If no configuration options are provided
// it uses the default configuration.
//
// It panics if an error occurs.
func New(opts ...exrpcommon.ConfigOption) *UpgradeIntegrationNetwork {
	cfg := DefaultUpgradeConfig()
	// Modify the default config with the given options
	for _, opt := range opts {
		opt(&cfg)
	}

	ctx := sdktypes.Context{}
	network := &UpgradeIntegrationNetwork{
		cfg:        cfg,
		ctx:        ctx,
		validators: []stakingtypes.Validator{},
	}

	err := network.configureAndInitChain()
	if err != nil {
		panic(err)
	}
	return network
}

// configureAndInitChain initializes the network with the given configuration.
// It creates the genesis state and starts the network.
func (n *UpgradeIntegrationNetwork) configureAndInitChain() error {
	// Create a new EvmosApp with the following params
	exrpApp := CreateExrpApp(n.cfg.ChainID, n.cfg.DataDir, n.cfg.NodeDBName, n.cfg.CustomBaseAppOpts...)

	validators, err := exrpApp.StakingKeeper.GetBondedValidatorsByPower(exrpApp.NewContext(true))
	if err != nil {
		return err
	}

	valSet, err := createStakingValidators(validators)
	if err != nil {
		return err
	}

	header := cmtproto.Header{
		ChainID:            n.cfg.ChainID,
		Height:             exrpApp.LastBlockHeight() + 1,
		AppHash:            exrpApp.LastCommitID().Hash,
		Time:               time.Now().UTC(),
		ValidatorsHash:     valSet.Hash(),
		NextValidatorsHash: valSet.Hash(),
		ProposerAddress:    valSet.Proposer.Address,
		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
	}

	req := BuildFinalizeBlockReq(header, valSet.Validators, nil, nil)
	if _, err := exrpApp.FinalizeBlock(req); err != nil {
		return err
	}

	// Store upgrade plan on finalizeblock state.
	upgradePlan := upgradetypes.Plan{
		Name:   n.cfg.UpgradePlanName,
		Height: exrpApp.LastBlockHeight() + 1,
	}

	bz, err := exrpApp.AppCodec().Marshal(&upgradePlan)
	if err != nil {
		return err
	}

	// TODO - this might not be the best way to initilize the context
	n.ctx = exrpApp.BaseApp.NewContextLegacy(false, header)

	upgradeKey := exrpApp.GetKey(upgradetypes.StoreKey)
	upgradeStoreService := runtime.NewKVStoreService(upgradeKey)
	upgradeStore := upgradeStoreService.OpenKVStore(n.ctx)
	err = upgradeStore.Set(upgradetypes.PlanKey(), bz)
	if err != nil {
		return err
	}

	// Commit genesis changes
	if _, err := exrpApp.Commit(); err != nil {
		return err
	}

	n.app = exrpApp

	n.valSet = valSet

	return nil
}

// GetContext returns the network's context
func (n *UpgradeIntegrationNetwork) GetContext() sdktypes.Context {
	return n.ctx
}

// WithIsCheckTxCtx switches the network's checkTx property
func (n *UpgradeIntegrationNetwork) WithIsCheckTxCtx(isCheckTx bool) sdktypes.Context {
	n.ctx = n.ctx.WithIsCheckTx(isCheckTx)
	return n.ctx
}

// GetChainID returns the network's chainID
func (n *UpgradeIntegrationNetwork) GetChainID() string {
	return n.cfg.ChainID
}

// GetEIP155ChainID returns the network EIp-155 chainID number
func (n *UpgradeIntegrationNetwork) GetEIP155ChainID() *big.Int {
	return n.cfg.EIP155ChainID
}

// GetDenom returns the network's denom
func (n *UpgradeIntegrationNetwork) GetDenom() string {
	return n.cfg.Denom
}

// GetBondDenom returns the network's bond denom
func (n *UpgradeIntegrationNetwork) GetBondDenom() string {
	return n.cfg.BondDenom
}

// GetOtherDenoms returns network's other supported denoms
func (n *UpgradeIntegrationNetwork) GetOtherDenoms() []string {
	return n.cfg.OtherCoinDenom
}

// GetValidators returns the network's validators
func (n *UpgradeIntegrationNetwork) GetValidators() []stakingtypes.Validator {
	return n.validators
}

// GetMinDepositAmt returns the network's min deposit amount
func (n *UpgradeIntegrationNetwork) GetMinDepositAmt() sdkmath.Int {
	return n.cfg.MinDepositAmt
}

// GetOtherDenoms returns network's other supported denoms
func (n *UpgradeIntegrationNetwork) GetEncodingConfig() sdktestutil.TestEncodingConfig {
	return sdktestutil.TestEncodingConfig{
		InterfaceRegistry: n.app.InterfaceRegistry(),
		Codec:             n.app.AppCodec(),
		TxConfig:          n.app.GetTxConfig(),
		Amino:             n.app.LegacyAmino(),
	}
}

// BroadcastTxSync broadcasts the given txBytes to the network and returns the response.
// TODO - this should be change to gRPC
func (n *UpgradeIntegrationNetwork) BroadcastTxSync(txBytes []byte) (abcitypes.ExecTxResult, error) {
	header := n.ctx.BlockHeader()
	// Update block header and BeginBlock
	header.Height++
	header.AppHash = n.app.LastCommitID().Hash
	// Calculate new block time after duration
	newBlockTime := header.Time.Add(time.Second)
	header.Time = newBlockTime

	req := BuildFinalizeBlockReq(header, n.valSet.Validators, txBytes)

	// dont include the DecidedLastCommit because we're not committing the changes
	// here, is just for broadcasting the tx. To persist the changes, use the
	// NextBlock or NextBlockAfter functions
	req.DecidedLastCommit = abcitypes.CommitInfo{}

	blockRes, err := n.app.BaseApp.FinalizeBlock(req)
	if err != nil {
		return abcitypes.ExecTxResult{}, err
	}
	if len(blockRes.TxResults) != 1 {
		return abcitypes.ExecTxResult{}, fmt.Errorf("unexpected number of tx results. Expected 1, got: %d", len(blockRes.TxResults))
	}
	return *blockRes.TxResults[0], nil
}

// Simulate simulates the given txBytes to the network and returns the simulated response.
// TODO - this should be change to gRPC
func (n *UpgradeIntegrationNetwork) Simulate(txBytes []byte) (*txtypes.SimulateResponse, error) {
	gas, result, err := n.app.BaseApp.Simulate(txBytes)
	if err != nil {
		return nil, err
	}
	return &txtypes.SimulateResponse{
		GasInfo: &gas,
		Result:  result,
	}, nil
}

// CheckTx calls the BaseApp's CheckTx method with the given txBytes to the network and returns the response.
func (n *UpgradeIntegrationNetwork) CheckTx(txBytes []byte) (*abcitypes.ResponseCheckTx, error) {
	req := &abcitypes.RequestCheckTx{Tx: txBytes}
	res, err := n.app.BaseApp.CheckTx(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func createStakingValidators(validators []stakingtypes.Validator) (*cmttypes.ValidatorSet, error) {
	tmValidators := make([]*cmttypes.Validator, 0, len(validators))
	for _, val := range validators {
		pb, err := val.CmtConsPublicKey()
		if err != nil {
			return nil, err
		}
		pubKey := ed25519.PubKey(pb.GetEd25519())
		validator := cmttypes.NewValidator(pubKey, val.GetConsensusPower(val.Tokens))
		tmValidators = append(tmValidators, validator)
	}
	return cmttypes.NewValidatorSet(tmValidators), nil
}
