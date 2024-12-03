package testupgrade

import (
	commonnetwork "github.com/xrplevm/node/v4/testutil/integration/common/network"
	exrpcommon "github.com/xrplevm/node/v4/testutil/integration/exrp/common"
	upgradenetwork "github.com/xrplevm/node/v4/testutil/integration/exrp/upgrade"
)

var _ commonnetwork.Network = (*UpgradeTestNetwork)(nil)

type UpgradeTestNetwork struct {
	upgradenetwork.UpgradeIntegrationNetwork
}

func NewUpgradeTestNetwork(opts ...upgradenetwork.UpgradeConfigOption) *UpgradeTestNetwork {
	network := upgradenetwork.New(opts...)
	return &UpgradeTestNetwork{
		UpgradeIntegrationNetwork: *network,
	}
}

func (n *UpgradeTestNetwork) SetupSdkConfig() {
	exrpcommon.SetupSdkConfig()
}
