package poa

import (
	"github.com/Peersyst/exrp/testutil/network"
	"github.com/stretchr/testify/suite"
	"testing"
)

func NewPoaTest(cfg network.Config, net *network.Network) *TestSuite {
	return &TestSuite{cfg: cfg, network: net}
}

func Test(t *testing.T) {
	net := network.NewTestNetwork(t, 3, 2)
	suite.Run(t, NewPoaTest(net.Config, net))
}

func (s *TestSuite) SetupTest() {
	s.network.Cleanup()
	s.network = network.NewTestNetwork(s.T(), 3, 2)
	s.proposalCount = 0
}
