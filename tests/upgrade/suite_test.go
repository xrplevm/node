package testupgrade

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/xrplevm/node/v5/app"
)

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	denom := s.network.GetDenom()
	s.Require().NotEmpty(denom)
	s.Require().Equal(denom, app.BaseDenom)
}
