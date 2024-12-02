package testupgrade

import (
	commonnetwork "github.com/xrplevm/node/v4/testutil/integration/common/network"
	exrpcommon "github.com/xrplevm/node/v4/testutil/integration/exrp/common"
	exrpnetwork "github.com/xrplevm/node/v4/testutil/integration/exrp/network"
)

var _ commonnetwork.Network = (*UpgradeTestNetwork)(nil)

type UpgradeTestNetwork struct {
	exrpnetwork.IntegrationNetwork
}

func NewUpgradeTestNetwork(opts ...exrpnetwork.ConfigOption) *UpgradeTestNetwork {
	network := exrpnetwork.New(opts...)
	return &UpgradeTestNetwork{
		IntegrationNetwork: *network,
	}
}

func (n *UpgradeTestNetwork) SetupSdkConfig() {
	exrpcommon.SetupSdkConfig()
}
