package poa_test

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
	s.SetupNetwork(3, 2, time.Second, 3)
}

func Test_TestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
