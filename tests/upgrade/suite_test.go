package testupgrade

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {

	denom := s.network.GetDenom()
	s.Require().NotEmpty(denom)
	s.Require().Equal(denom, "uxrp")
}