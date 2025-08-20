package exrpupgrade

import (
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	erc20keeper "github.com/cosmos/evm/x/erc20/keeper"
	feemarketkeeper "github.com/cosmos/evm/x/feemarket/keeper"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
	poakeeper "github.com/xrplevm/node/v9/x/poa/keeper"
)

func (n *UpgradeIntegrationNetwork) BankKeeper() bankkeeper.Keeper {
	return n.app.BankKeeper
}

func (n *UpgradeIntegrationNetwork) ERC20Keeper() erc20keeper.Keeper {
	return n.app.Erc20Keeper
}

func (n *UpgradeIntegrationNetwork) EvmKeeper() evmkeeper.Keeper {
	return *n.app.EvmKeeper
}

func (n *UpgradeIntegrationNetwork) GovKeeper() *govkeeper.Keeper {
	return &n.app.GovKeeper
}

func (n *UpgradeIntegrationNetwork) StakingKeeper() *stakingkeeper.Keeper {
	return n.app.StakingKeeper
}

func (n *UpgradeIntegrationNetwork) SlashingKeeper() slashingkeeper.Keeper {
	return n.app.SlashingKeeper
}

func (n *UpgradeIntegrationNetwork) DistrKeeper() distrkeeper.Keeper {
	return n.app.DistrKeeper
}

func (n *UpgradeIntegrationNetwork) AccountKeeper() authkeeper.AccountKeeper {
	return n.app.AccountKeeper
}

func (n *UpgradeIntegrationNetwork) AuthzKeeper() authzkeeper.Keeper {
	return n.app.AuthzKeeper
}

func (n *UpgradeIntegrationNetwork) FeeMarketKeeper() feemarketkeeper.Keeper {
	return n.app.FeeMarketKeeper
}

func (n *UpgradeIntegrationNetwork) PoaKeeper() poakeeper.Keeper {
	return n.app.PoaKeeper
}
