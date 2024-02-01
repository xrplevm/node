package e2e

import (
	"fmt"
	"github.com/Peersyst/exrp/testutil/network"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/suite"
	"time"
)

type IntegrationTestSuite struct {
	suite.Suite

	ProposalCount int
	Cfg           network.Config
	Network       *network.Network
}

func (s *IntegrationTestSuite) SetupNetwork(numValidators int, numBondedValidators int, blockTime time.Duration, unbondingBlocks int64) {
	s.T().Log("setting up network test suite")

	var err error
	cfg := network.DefaultConfig(numValidators, numBondedValidators, blockTime, unbondingBlocks)

	s.Network, err = network.New(s.T(), s.T().TempDir(), cfg)
	s.Cfg = cfg
	s.ProposalCount = 0

	s.Require().NoError(err)
	s.Require().NotNil(s.Network)

	_, err = s.Network.WaitForHeight(2)
	s.Require().NoError(err)

	if s.Network.Validators[0].JSONRPCClient == nil {
		address := fmt.Sprintf("http://%s", s.Network.Validators[0].AppConfig.JSONRPC.Address)
		s.Network.Validators[0].JSONRPCClient, err = ethclient.Dial(address)
		s.Require().NoError(err)
	}
}

func (s *IntegrationTestSuite) TearDownTest() {
	s.T().Log("tearing down network test suite")
	s.Network.Cleanup()
	time.Sleep(5 * time.Second)
}
