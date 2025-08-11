package integration

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	commonnetwork "github.com/xrplevm/node/v8/testutil/integration/common/network"
	exrpcommon "github.com/xrplevm/node/v8/testutil/integration/exrp/common"
	exrpintegration "github.com/xrplevm/node/v8/testutil/integration/exrp/integration"
	poatypes "github.com/xrplevm/node/v8/x/poa/types"
)

var _ commonnetwork.Network = (*Network)(nil)

type Network struct {
	exrpintegration.IntegrationNetwork
}

func NewIntegrationNetwork(opts ...exrpcommon.ConfigOption) *Network {
	network := exrpintegration.New(opts...)
	return &Network{
		IntegrationNetwork: *network,
	}
}

func (n *Network) SetupSdkConfig() {
	exrpcommon.SetupSdkConfig()
}

func (n *Network) GetERC20Client() erc20types.QueryClient {
	return exrpcommon.GetERC20Client(n)
}

func (n *Network) GetEvmClient() evmtypes.QueryClient {
	return exrpcommon.GetEvmClient(n)
}

func (n *Network) GetGovClient() govtypes.QueryClient {
	return exrpcommon.GetGovClient(n)
}

func (n *Network) GetBankClient() banktypes.QueryClient {
	return exrpcommon.GetBankClient(n)
}

func (n *Network) GetFeeMarketClient() feemarkettypes.QueryClient {
	return exrpcommon.GetFeeMarketClient(n)
}

func (n *Network) GetAuthClient() authtypes.QueryClient {
	return exrpcommon.GetAuthClient(n)
}

func (n *Network) GetAuthzClient() authz.QueryClient {
	return exrpcommon.GetAuthzClient(n)
}

func (n *Network) GetStakingClient() stakingtypes.QueryClient {
	return exrpcommon.GetStakingClient(n)
}

func (n *Network) GetSlashingClient() slashingtypes.QueryClient {
	return exrpcommon.GetSlashingClient(n)
}

func (n *Network) GetDistrClient() distrtypes.QueryClient {
	return exrpcommon.GetDistrClient(n)
}

func (n *Network) GetPoaClient() poatypes.QueryClient {
	return exrpcommon.GetPoaClient(n)
}
