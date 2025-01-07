package integration

import (
	"testing"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"
	"github.com/xrplevm/node/v5/app"
)

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestIntegration() {
	denom := s.Network().GetDenom()
	s.Require().NotEmpty(denom)
	s.Require().Equal(app.BaseDenom, denom)

	balances, err := s.Network().BankClient().AllBalances(s.network.GetContext(), &banktypes.QueryAllBalancesRequest{
		Address: "ethm1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3w48d64",
	})

	s.T().Log("balances", balances)
	s.Require().NoError(err)

	err = s.network.NextBlock()
	s.Require().NoError(err)

	res, err := s.Network().StakingClient().Validators(s.network.GetContext(), &stakingtypes.QueryValidatorsRequest{})
	s.Require().NoError(err)

	s.T().Log("validators", len(res.Validators))

	s.Require().Equal(len(res.Validators), len(s.network.GetValidators()))
}
