package unbonding_test

import (
	"github.com/Peersyst/exrp/v2/tests/e2e"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type TestSuite struct {
	e2e.IntegrationTestSuite
}

func (s *TestSuite) SetupTest() {
	s.SetupNetwork(5, 4, 3*time.Second, 5)
}

func Test_TestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
