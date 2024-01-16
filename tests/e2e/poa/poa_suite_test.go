package poa_test

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"testing"

	"github.com/Peersyst/exrp/testutil/network"
	"github.com/ethereum/go-ethereum/ethclient"
)

type IntegrationTestSuite struct {
	suite.Suite

	proposalCount int

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupTest() {
	s.T().Log("setting up network test suite")

	var err error
	cfg := network.DefaultConfig(3, 2)

	s.network, err = network.New(s.T(), s.T().TempDir(), cfg)
	s.cfg = cfg
	s.proposalCount = 0

	s.Require().NoError(err)
	s.Require().NotNil(s.network)

	_, err = s.network.WaitForHeight(2)
	s.Require().NoError(err)

	if s.network.Validators[0].JSONRPCClient == nil {
		address := fmt.Sprintf("http://%s", s.network.Validators[0].AppConfig.JSONRPC.Address)
		s.network.Validators[0].JSONRPCClient, err = ethclient.Dial(address)
		s.Require().NoError(err)
	}
}

func (s *IntegrationTestSuite) TearDownTest() {
	s.T().Log("tearing down network test suite")
	s.network.Cleanup()
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
