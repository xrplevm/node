package testupgrade

import (
	"testing"
	"time"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

	res, err := s.network.GetUpgradeClient().CurrentPlan(
		s.network.GetContext(),
		&upgradetypes.QueryCurrentPlanRequest{},
	)
	s.Require().NoError(err)
	s.Require().Equal("v6.0.0", res.Plan.Name)
	
	s.Require().True(s.Network().UpgradeKeeper().HasHandler("v6.0.0"))

	err = s.network.UpgradeKeeper().ApplyUpgrade(
		s.Network().GetContext(),
		*res.Plan,
	)

	s.Require().NoError(err)

	resParams, err := s.network.GetStakingClient().Params(
		s.network.GetContext(),
		&stakingtypes.QueryParamsRequest{},
	)

	s.Require().Equal(100 * time.Second, resParams.Params.UnbondingTime)
}
