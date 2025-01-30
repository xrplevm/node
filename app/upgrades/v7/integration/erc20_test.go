//nolint:dupl
package integration

import (
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
)

func (s *UpgradeTestSuite) TestUpgrade_ERC20Params() {
	prevParams, err := s.network.GetERC20Client().Params(
		s.network.GetContext(),
		&erc20types.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	postParams, err := s.network.GetERC20Client().Params(
		s.network.GetContext(),
		&erc20types.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	// Check that not modified params are the same
	s.Require().Equal(prevParams.Params, postParams.Params)
}

func (s *UpgradeTestSuite) TestUpgrade_ERC20_TokenPairs() {
	prevTokenPairs, err := s.network.GetERC20Client().TokenPairs(
		s.network.GetContext(),
		&erc20types.QueryTokenPairsRequest{},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	postTokenPairs, err := s.network.GetERC20Client().TokenPairs(
		s.network.GetContext(),
		&erc20types.QueryTokenPairsRequest{},
	)
	s.Require().NoError(err)

	// Check that not modified token pairs are the same
	s.Require().Equal(prevTokenPairs.TokenPairs, postTokenPairs.TokenPairs)
}
