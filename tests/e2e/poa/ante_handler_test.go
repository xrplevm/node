package poa_test

import (
	"github.com/Peersyst/exrp/v2/tests/e2e"
	"strings"
)

func (s *TestSuite) Test_AnteHandlerForbiddenTransactions() {
	s.T().Logf("==== Test_AnteHandlerForbiddenTransactions")
	validator := s.Network.Validators[0]
	dst := s.Network.Validators[1]

	res := e2e.UnBondTokens(&s.IntegrationTestSuite, validator, e2e.DefaultBondedTokens, false)
	s.Require().True(strings.Contains(res, "tx type not allowed"))
	res = e2e.Redelegate(&s.IntegrationTestSuite, validator, dst, e2e.DefaultBondedTokens)
	s.Require().True(strings.Contains(res, "tx type not allowed"))

	s.T().Logf("==== [V] Test_AnteHandlerForbiddenTransactions")
}
