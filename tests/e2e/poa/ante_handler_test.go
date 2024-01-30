package poa_test

import (
	"fmt"
	"strings"
)

func (s *IntegrationTestSuite) Test_AnteHandlerForbiddenTransactions() {
	fmt.Println("==== Test_AnteHandlerForbiddenTransactions")
	validator := s.network.Validators[0]
	dst := s.network.Validators[1]

	res := UnBondTokens(s, validator, DefaultBondedTokens, false)
	s.Require().True(strings.Contains(res, "tx type not allowed"))
	res = Redelegate(s, validator, dst, DefaultBondedTokens)
	s.Require().True(strings.Contains(res, "tx type not allowed"))

	fmt.Println("==== [V] Test_AnteHandlerForbiddenTransactions")
}
