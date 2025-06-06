// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package exrpintegration

import (
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/x/evm/statedb"
	inflationtypes "github.com/evmos/evmos/v20/x/inflation/v1/types"
	"github.com/xrplevm/node/v8/app"
	exrpcommon "github.com/xrplevm/node/v8/testutil/integration/exrp/common"
)

// UnitTestIntegrationNetwork is the implementation of the Network interface for unit tests.
// It embeds the IntegrationNetwork struct to reuse its methods and
// makes the App public for easier testing.
type UnitTestIntegrationNetwork struct {
	IntegrationNetwork
	App *app.App
}

var _ Network = (*UnitTestIntegrationNetwork)(nil)

// NewUnitTestNetwork configures and initializes a new Evmos Network instance with
// the given configuration options. If no configuration options are provided
// it uses the default configuration.
//
// It panics if an error occurs.
// Note: Only uses for Unit Tests
func NewUnitTestNetwork(opts ...exrpcommon.ConfigOption) *UnitTestIntegrationNetwork {
	network := New(opts...)
	return &UnitTestIntegrationNetwork{
		IntegrationNetwork: *network,
		App:                network.app,
	}
}

// GetStateDB returns the state database for the current block.
func (n *UnitTestIntegrationNetwork) GetStateDB() *statedb.StateDB {
	headerHash := n.GetContext().HeaderHash()
	return statedb.New(
		n.GetContext(),
		n.App.EvmKeeper,
		statedb.NewEmptyTxConfig(common.BytesToHash(headerHash)),
	)
}

// FundAccount funds the given account with the given amount of coins.
func (n *UnitTestIntegrationNetwork) FundAccount(addr sdktypes.AccAddress, coins sdktypes.Coins) error {
	ctx := n.GetContext()

	if err := n.app.BankKeeper.MintCoins(ctx, inflationtypes.ModuleName, coins); err != nil {
		return err
	}

	return n.app.BankKeeper.SendCoinsFromModuleToAccount(ctx, inflationtypes.ModuleName, addr, coins)
}
