package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/xrplevm/node/v3/e2e/manager/network"
)

func (s *End2EndSuite) SetupSuite() {
	networkManager := network.NewManager(s.T(), 5, 4, 3*time.Second, 5, s.T().TempDir())
	s.SetManager(networkManager)
}

func Test_Suite(t *testing.T) {
	suite.Run(t, new(End2EndSuite))
}