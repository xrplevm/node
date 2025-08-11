// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package exrpcommon

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	erc20keeper "github.com/cosmos/evm/x/erc20/keeper"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	feemarketkeeper "github.com/cosmos/evm/x/feemarket/keeper"
	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	poakeeper "github.com/xrplevm/node/v8/x/poa/keeper"
	poatypes "github.com/xrplevm/node/v8/x/poa/types"
)

type NetworkKeepers interface {
	GetContext() sdktypes.Context
	GetEncodingConfig() testutil.TestEncodingConfig

	ERC20Keeper() erc20keeper.Keeper
	EvmKeeper() evmkeeper.Keeper
	GovKeeper() *govkeeper.Keeper
	BankKeeper() bankkeeper.Keeper
	StakingKeeper() *stakingkeeper.Keeper
	SlashingKeeper() slashingkeeper.Keeper
	DistrKeeper() distrkeeper.Keeper
	AccountKeeper() authkeeper.AccountKeeper
	AuthzKeeper() authzkeeper.Keeper
	FeeMarketKeeper() feemarketkeeper.Keeper
	PoaKeeper() poakeeper.Keeper
}

func getQueryHelper(ctx sdktypes.Context, encCfg testutil.TestEncodingConfig) *baseapp.QueryServiceTestHelper {
	interfaceRegistry := encCfg.InterfaceRegistry
	// This is needed so that state changes are not committed in precompiles
	// simulations.
	cacheCtx, _ := ctx.CacheContext()
	return baseapp.NewQueryServerTestHelper(cacheCtx, interfaceRegistry)
}

func GetERC20Client(n NetworkKeepers) erc20types.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	erc20types.RegisterQueryServer(queryHelper, n.ERC20Keeper())
	return erc20types.NewQueryClient(queryHelper)
}

func GetEvmClient(n NetworkKeepers) evmtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	evmtypes.RegisterQueryServer(queryHelper, n.EvmKeeper())
	return evmtypes.NewQueryClient(queryHelper)
}

func GetGovClient(n NetworkKeepers) govtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	govtypes.RegisterQueryServer(queryHelper, govkeeper.NewQueryServer(n.GovKeeper()))
	return govtypes.NewQueryClient(queryHelper)
}

func GetBankClient(n NetworkKeepers) banktypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	banktypes.RegisterQueryServer(queryHelper, n.BankKeeper())
	return banktypes.NewQueryClient(queryHelper)
}

func GetFeeMarketClient(n NetworkKeepers) feemarkettypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	feemarkettypes.RegisterQueryServer(queryHelper, n.FeeMarketKeeper())
	return feemarkettypes.NewQueryClient(queryHelper)
}

func GetAuthClient(n NetworkKeepers) authtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	authtypes.RegisterQueryServer(queryHelper, authkeeper.NewQueryServer(n.AccountKeeper()))
	return authtypes.NewQueryClient(queryHelper)
}

func GetAuthzClient(n NetworkKeepers) authz.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	authz.RegisterQueryServer(queryHelper, n.AuthzKeeper())
	return authz.NewQueryClient(queryHelper)
}

func GetStakingClient(n NetworkKeepers) stakingtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	stakingtypes.RegisterQueryServer(queryHelper, stakingkeeper.Querier{Keeper: n.StakingKeeper()})
	return stakingtypes.NewQueryClient(queryHelper)
}

func GetSlashingClient(n NetworkKeepers) slashingtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	slashingtypes.RegisterQueryServer(queryHelper, slashingkeeper.Querier{Keeper: n.SlashingKeeper()})
	return slashingtypes.NewQueryClient(queryHelper)
}

func GetDistrClient(n NetworkKeepers) distrtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	distrtypes.RegisterQueryServer(queryHelper, distrkeeper.Querier{Keeper: n.DistrKeeper()})
	return distrtypes.NewQueryClient(queryHelper)
}

func GetPoaClient(n NetworkKeepers) poatypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	poatypes.RegisterQueryServer(queryHelper, poakeeper.Querier{Keeper: n.PoaKeeper()})
	return poatypes.NewQueryClient(queryHelper)
}
