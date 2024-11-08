package e2e

import (
	grpchandler "github.com/evmos/evmos/v20/testutil/integration/common/grpc"
	testkeyring "github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	"github.com/stretchr/testify/suite"
	"time"
)

type IntegrationTestSuite struct {
	suite.Suite

	ProposalCount int
	Cfg           network.Config
	Network       network.Network
	Handler       grpchandler.Handler
	Keyring       testkeyring.Keyring
}

func (s *IntegrationTestSuite) SetupNetwork(numValidators int, numBondedValidators int, blockTime time.Duration, unbondingBlocks int64) {
	nw := network.New(
		network.WithBalances(),
	)
	handler := grpchandler.NewIntegrationHandler(nw)
	s.Network = nw
	s.Handler = handler
}

func (s *IntegrationTestSuite) TearDownTest() {
	s.T().Log("tearing down network test suite")
	s.Network.Cleanup()
	time.Sleep(5 * time.Second)
}
