package unbonding_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/xrplevm/node/v3/tests/e2e"
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
