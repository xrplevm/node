package testupgrade

import (
	commonnetwork "github.com/xrplevm/node/v4/testutil/integration/common/network"
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
