//nolint:dupl
package integration

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (s *UpgradeTestSuite) TestUpgrade_Bank_Params() {
	prevParams, err := s.network.GetBankClient().Params(
		s.network.GetContext(),
		&banktypes.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	postParams, err := s.network.GetBankClient().Params(
		s.network.GetContext(),
		&banktypes.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	// Check that not modified params are the same
	s.Require().Equal(prevParams.Params, postParams.Params)
}

func (s *UpgradeTestSuite) TestUpgrade_Bank_TotalSupply() {
	res, err := s.network.GetBankClient().TotalSupply(
		s.network.GetContext(),
		&banktypes.QueryTotalSupplyRequest{},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	postRes, err := s.network.GetBankClient().TotalSupply(
		s.network.GetContext(),
		&banktypes.QueryTotalSupplyRequest{},
	)
	s.Require().NoError(err)

	// Check that not modified balances are the same
	s.Require().Equal(res.Supply, postRes.Supply)
}
