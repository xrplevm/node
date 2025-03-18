//nolint:dupl
package integration

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (s *UpgradeTestSuite) TestUpgrade_Auth_Params() {
	prevParams, err := s.network.GetAuthClient().Params(
		s.network.GetContext(),
		&authtypes.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	postParams, err := s.network.GetAuthClient().Params(
		s.network.GetContext(),
		&authtypes.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	// Check that not modified params are the same
	s.Require().Equal(prevParams.Params, postParams.Params)
}

func (s *UpgradeTestSuite) TestUpgrade_Auth_Accounts() {
	res, err := s.network.GetAuthClient().Accounts(
		s.network.GetContext(),
		&authtypes.QueryAccountsRequest{},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	postRes, err := s.network.GetAuthClient().Accounts(
		s.network.GetContext(),
		&authtypes.QueryAccountsRequest{},
	)
	s.Require().NoError(err)

	// Check that not modified accounts are the same
	s.Require().Equal(res.Accounts, postRes.Accounts)
}
