package testupgrade

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	exrpupgrade "github.com/xrplevm/node/v5/testutil/integration/exrp/upgrade"
)

const (
	DefaultNodeDBName = ".exrp-upgrade"
	DefaultNodeDBDir  = "exrp-upgrade"
)

type UpgradeTestSuite struct {
	suite.Suite

	network *UpgradeTestNetwork
}

func (s *UpgradeTestSuite) Network() *UpgradeTestNetwork {
	return s.network
}

func (s *UpgradeTestSuite) SetupTest() {
	// Setup the SDK config
	s.network.SetupSdkConfig()

	s.Require().Equal(sdk.GetConfig().GetBech32AccountAddrPrefix(), "ethm")


	// Create the network
	s.network = NewUpgradeTestNetwork(
		exrpupgrade.WithUpgradePlanName("v6.0.0"),
	)

	// Check that the network was created successfully
	s.Require().NotNil(s.network)
}
