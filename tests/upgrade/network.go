package testupgrade

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v20/x/feemarket/types"
	commonnetwork "github.com/xrplevm/node/v9/testutil/integration/common/network"
	exrpcommon "github.com/xrplevm/node/v9/testutil/integration/exrp/common"
	upgradenetwork "github.com/xrplevm/node/v9/testutil/integration/exrp/upgrade"
	poatypes "github.com/xrplevm/node/v9/x/poa/types"
)

var _ commonnetwork.Network = (*UpgradeTestNetwork)(nil)

type UpgradeTestNetwork struct {
	upgradenetwork.UpgradeIntegrationNetwork
}

func NewUpgradeTestNetwork(opts ...exrpcommon.ConfigOption) *UpgradeTestNetwork {
	network := upgradenetwork.New(opts...)
	return &UpgradeTestNetwork{
		UpgradeIntegrationNetwork: *network,
	}
}

func (n *UpgradeTestNetwork) SetupSdkConfig() {
	exrpcommon.SetupSdkConfig()
}

func (n *UpgradeTestNetwork) GetERC20Client() erc20types.QueryClient {
	return exrpcommon.GetERC20Client(n)
}

func (n *UpgradeTestNetwork) GetEvmClient() evmtypes.QueryClient {
	return exrpcommon.GetEvmClient(n)
}

func (n *UpgradeTestNetwork) GetGovClient() govtypes.QueryClient {
	return exrpcommon.GetGovClient(n)
}

func (n *UpgradeTestNetwork) GetBankClient() banktypes.QueryClient {
	return exrpcommon.GetBankClient(n)
}

func (n *UpgradeTestNetwork) GetFeeMarketClient() feemarkettypes.QueryClient {
	return exrpcommon.GetFeeMarketClient(n)
}

func (n *UpgradeTestNetwork) GetAuthClient() authtypes.QueryClient {
	return exrpcommon.GetAuthClient(n)
}

func (n *UpgradeTestNetwork) GetAuthzClient() authz.QueryClient {
	return exrpcommon.GetAuthzClient(n)
}

func (n *UpgradeTestNetwork) GetStakingClient() stakingtypes.QueryClient {
	return exrpcommon.GetStakingClient(n)
}

func (n *UpgradeTestNetwork) GetDistrClient() distrtypes.QueryClient {
	return exrpcommon.GetDistrClient(n)
}

func (n *UpgradeTestNetwork) GetPoaClient() poatypes.QueryClient {
	return exrpcommon.GetPoaClient(n)
}
