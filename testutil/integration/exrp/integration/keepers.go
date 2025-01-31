package exrpintegration

import (
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	erc20keeper "github.com/evmos/evmos/v20/x/erc20/keeper"
	evmkeeper "github.com/evmos/evmos/v20/x/evm/keeper"
	feemarketkeeper "github.com/evmos/evmos/v20/x/feemarket/keeper"
	poakeeper "github.com/xrplevm/node/v6/x/poa/keeper"
)

func (n *IntegrationNetwork) BankKeeper() bankkeeper.Keeper {
	return n.app.BankKeeper
}

func (n *IntegrationNetwork) ERC20Keeper() erc20keeper.Keeper {
	return n.app.Erc20Keeper
}

func (n *IntegrationNetwork) EvmKeeper() evmkeeper.Keeper {
	return *n.app.EvmKeeper
}

func (n *IntegrationNetwork) GovKeeper() *govkeeper.Keeper {
	return &n.app.GovKeeper
}

func (n *IntegrationNetwork) StakingKeeper() *stakingkeeper.Keeper {
	return n.app.StakingKeeper.Keeper
}

func (n *IntegrationNetwork) SlashingKeeper() slashingkeeper.Keeper {
	return n.app.SlashingKeeper
}

func (n *IntegrationNetwork) DistrKeeper() distrkeeper.Keeper {
	return n.app.DistrKeeper
}

func (n *IntegrationNetwork) AccountKeeper() authkeeper.AccountKeeper {
	return n.app.AccountKeeper
}

func (n *IntegrationNetwork) AuthzKeeper() authzkeeper.Keeper {
	return n.app.AuthzKeeper
}

func (n *IntegrationNetwork) FeeMarketKeeper() feemarketkeeper.Keeper {
	return n.app.FeeMarketKeeper
}

func (n *IntegrationNetwork) PoaKeeper() poakeeper.Keeper {
	return n.app.PoaKeeper
}

func (n *IntegrationNetwork) UpgradeKeeper() upgradekeeper.Keeper {
	return *n.app.UpgradeKeeper
}
