package testupgrade

import (
	"testing"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"
	"github.com/xrplevm/node/v6/app"
)

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	denom := s.network.GetDenom()
	s.Require().NotEmpty(denom)
	s.Require().Equal(denom, app.BaseDenom)

	balances, err := s.Network().GetBankClient().AllBalances(s.network.GetContext(), &banktypes.QueryAllBalancesRequest{
		Address: "ethm1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3w48d64",
	})

	s.T().Log("balances", balances)
	s.Require().NoError(err)

	err = s.network.NextBlock()
	s.Require().NoError(err)

	res, err := s.Network().GetStakingClient().Validators(s.network.GetContext(), &stakingtypes.QueryValidatorsRequest{})
	s.Require().NoError(err)

	s.T().Log("validators", len(res.Validators))
	s.Require().Equal(len(res.Validators), 1)
}
