package integration

import (
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func (s *UpgradeTestSuite) TestUpgrade_GovParams() {
	prevParams, err := s.network.GetGovClient().Params(
		s.network.GetContext(),
		&govtypes.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	postParams, err := s.network.GetGovClient().Params(
		s.network.GetContext(),
		&govtypes.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	// Check that not modified params are the same
	s.Require().Equal(prevParams.Params, postParams.Params)
}
