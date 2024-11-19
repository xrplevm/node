package e2e

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/suite"
	"github.com/xrplevm/node/v3/e2e/manager"
)

type End2EndSuite struct {
	suite.Suite

	manager manager.Manager
}


func (s *End2EndSuite) SetManager(manager manager.Manager) {
	s.T().Log("setting up test suite manager")

	s.manager = manager
	s.Require().NotNil(s.manager)

	// Load state
	if s.manager.JSONRPCClient() == nil {
		address := fmt.Sprintf("http://%s", s.manager.AppConfig().JSONRPC.Address)
		client, err := ethclient.Dial(address)
		s.Require().NoError(err)
		s.manager.SetJSONRPCClient(client)
	}

	s.T().Log("test suite manager setup successfully")
}

func (s *End2EndSuite) TearDownTest() {
	s.T().Log("tearing down test suite")
	s.manager.Cleanup()
	time.Sleep(5 * time.Second)
}
