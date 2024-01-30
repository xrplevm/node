package unbonding_test

import (
	"github.com/Peersyst/exrp/tests/e2e"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type TestSuite struct {
	e2e.IntegrationTestSuite
}

func (s *TestSuite) SetupTest() {
	s.SetupNetwork(5, 4, 3*time.Second)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TearDownTest() {
	s.IntegrationTestSuite.TearDownTest()
	time.Sleep(10 * time.Second)
}

func Test_TestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
