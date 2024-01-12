package poa_test

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"testing"

	"github.com/Peersyst/exrp/testutil/network"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/evmos/evmos/v15/server/config"
)

type IntegrationTestSuite struct {
	suite.Suite

	proposalCount int

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	var err error
	cfg := network.DefaultConfig(2, 4)
	cfg.JSONRPCAddress = config.DefaultJSONRPCAddress
	cfg.NumValidators = 1

	s.network, err = network.New(s.T(), s.T().TempDir(), cfg)
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

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
