// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package exrpupgrade

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v20/x/feemarket/types"
)

func getQueryHelper(ctx sdktypes.Context, encCfg testutil.TestEncodingConfig) *baseapp.QueryServiceTestHelper {
	interfaceRegistry := encCfg.InterfaceRegistry
	// This is needed so that state changes are not committed in precompiles
	// simulations.
	cacheCtx, _ := ctx.CacheContext()
	return baseapp.NewQueryServerTestHelper(cacheCtx, interfaceRegistry)
}

func (n *UpgradeIntegrationNetwork) GetERC20Client() erc20types.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	erc20types.RegisterQueryServer(queryHelper, n.app.Erc20Keeper)
	return erc20types.NewQueryClient(queryHelper)
}

func (n *UpgradeIntegrationNetwork) GetEvmClient() evmtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	evmtypes.RegisterQueryServer(queryHelper, n.app.EvmKeeper)
	return evmtypes.NewQueryClient(queryHelper)
}

func (n *UpgradeIntegrationNetwork) GetGovClient() govtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	govtypes.RegisterQueryServer(queryHelper, govkeeper.NewQueryServer(&n.app.GovKeeper))
	return govtypes.NewQueryClient(queryHelper)
}

func (n *UpgradeIntegrationNetwork) GetBankClient() banktypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	banktypes.RegisterQueryServer(queryHelper, n.app.BankKeeper)
	return banktypes.NewQueryClient(queryHelper)
}

func (n *UpgradeIntegrationNetwork) GetFeeMarketClient() feemarkettypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	feemarkettypes.RegisterQueryServer(queryHelper, n.app.FeeMarketKeeper)
	return feemarkettypes.NewQueryClient(queryHelper)
}

func (n *UpgradeIntegrationNetwork) GetAuthClient() authtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	authtypes.RegisterQueryServer(queryHelper, authkeeper.NewQueryServer(n.app.AccountKeeper))
	return authtypes.NewQueryClient(queryHelper)
}

func (n *UpgradeIntegrationNetwork) GetAuthzClient() authz.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	authz.RegisterQueryServer(queryHelper, n.app.AuthzKeeper)
	return authz.NewQueryClient(queryHelper)
}

func (n *UpgradeIntegrationNetwork) GetStakingClient() stakingtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	stakingtypes.RegisterQueryServer(queryHelper, stakingkeeper.Querier{Keeper: n.app.StakingKeeper.Keeper})
	return stakingtypes.NewQueryClient(queryHelper)
}

func (n *UpgradeIntegrationNetwork) GetDistrClient() distrtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	distrtypes.RegisterQueryServer(queryHelper, distrkeeper.Querier{Keeper: n.app.DistrKeeper})
	return distrtypes.NewQueryClient(queryHelper)
}
