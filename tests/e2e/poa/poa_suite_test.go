package poa_test

import (
	"github.com/Peersyst/exrp/tests/e2e"
	"time"
)

type TestSuite struct {
	e2e.IntegrationTestSuite
}

func (s *TestSuite) SetupTest() {
	s.SetupNetwork(3, 2, time.Second, 3)
}
