package testupgrade

import (
	"github.com/stretchr/testify/suite"
	exrpnetwork "github.com/xrplevm/node/v4/testutil/integration/exrp/network"
)

type UpgradeTestSuite struct {
	suite.Suite

	network *UpgradeTestNetwork
}

func (s *UpgradeTestSuite) SetupTest() {
	// READ APP STATE FILE

	s.network = NewUpgradeTestNetwork(
		// LOAD APP STATE FROM FILE
		exrpnetwork.WithGenesisFile("exported-state.json"),
	)

	s.Require().NotNil(s.network)
}