// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package exrpupgrade

import (
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/evm/x/vm/statedb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/xrplevm/node/v9/app"
	exrpcommon "github.com/xrplevm/node/v9/testutil/integration/exrp/common"
)

// UnitTestUpgradeNetwork is the implementation of the Network interface for unit tests.
// It embeds the IntegrationNetwork struct to reuse its methods and
// makes the App public for easier testing.
type UnitTestUpgradeNetwork struct {
	UpgradeIntegrationNetwork
	App *app.App
}

var _ Network = (*UnitTestUpgradeNetwork)(nil)

// NewUnitTestNetwork configures and initializes a new Evmos Network instance with
// the given configuration options. If no configuration options are provided
// it uses the default configuration.
//
// It panics if an error occurs.
// Note: Only uses for Unit Tests
func NewUnitTestUpgradeNetwork(opts ...exrpcommon.ConfigOption) *UnitTestUpgradeNetwork {
	network := New(opts...)
	return &UnitTestUpgradeNetwork{
		UpgradeIntegrationNetwork: *network,
		App:                       network.app,
	}
}

// GetStateDB returns the state database for the current block.
func (n *UnitTestUpgradeNetwork) GetStateDB() *statedb.StateDB {
	headerHash := n.GetContext().HeaderHash()
	return statedb.New(
		n.GetContext(),
		n.App.EvmKeeper,
		statedb.NewEmptyTxConfig(common.BytesToHash(headerHash)),
	)
}

// FundAccount funds the given account with the given amount of coins.
func (n *UnitTestUpgradeNetwork) FundAccount(addr sdktypes.AccAddress, coins sdktypes.Coins) error {
	ctx := n.GetContext()

	if err := n.app.BankKeeper.MintCoins(ctx, banktypes.ModuleName, coins); err != nil {
		return err
	}

	return n.app.BankKeeper.SendCoinsFromModuleToAccount(ctx, banktypes.ModuleName, addr, coins)
}
