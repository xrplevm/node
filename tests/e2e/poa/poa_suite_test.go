package poa_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/xrplevm/node/v2/tests/e2e"
)

type TestSuite struct {
	e2e.IntegrationTestSuite
}

func (s *TestSuite) SetupTest() {
	s.SetupNetwork(3, 2, time.Second, 3)
}

func Test_TestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
