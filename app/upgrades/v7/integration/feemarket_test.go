package integration

import (
	feemarkettypes "github.com/evmos/evmos/v20/x/feemarket/types"
)

func (s *UpgradeTestSuite) TestUpgrade_FeeMarketParams() {
	prevParams, err := s.network.GetFeeMarketClient().Params(
		s.network.GetContext(),
		&feemarkettypes.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	postParams, err := s.network.GetFeeMarketClient().Params(
		s.network.GetContext(),
		&feemarkettypes.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	// Check that not modified params are the same
	s.Require().Equal(prevParams.Params, postParams.Params)
}
