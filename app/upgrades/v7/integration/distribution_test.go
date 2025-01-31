package integration

import (
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func (s *UpgradeTestSuite) TestUpgrade_DistributionParams() {
	prevParams, err := s.network.GetDistrClient().Params(
		s.network.GetContext(),
		&distributiontypes.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	postParams, err := s.network.GetDistrClient().Params(
		s.network.GetContext(),
		&distributiontypes.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	// Check that not modified params are the same
	s.Require().Equal(prevParams.Params, postParams.Params)
}
