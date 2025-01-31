package integration

import (
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
)

func (s *UpgradeTestSuite) TestUpgrade_EvmParams() {
	prevParams, err := s.network.GetEvmClient().Params(
		s.network.GetContext(),
		&evmtypes.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	postParams, err := s.network.GetEvmClient().Params(
		s.network.GetContext(),
		&evmtypes.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	// Check that not modified params are the same
	s.Require().Equal(prevParams.Params, postParams.Params)
}
