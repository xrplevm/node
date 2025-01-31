//nolint:dupl
package integration

import (
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func (s *UpgradeTestSuite) TestUpgrade_SlashingParams() {
	prevParams, err := s.network.GetSlashingClient().Params(
		s.network.GetContext(),
		&slashingtypes.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	postParams, err := s.network.GetSlashingClient().Params(
		s.network.GetContext(),
		&slashingtypes.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	// Check that not modified params are the same
	s.Require().Equal(prevParams.Params, postParams.Params)
}

func (s *UpgradeTestSuite) TestUpgrade_Slashing_SigningInfos() {
	prevSigningInfos, err := s.network.GetSlashingClient().SigningInfos(
		s.network.GetContext(),
		&slashingtypes.QuerySigningInfosRequest{},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	postSigningInfos, err := s.network.GetSlashingClient().SigningInfos(
		s.network.GetContext(),
		&slashingtypes.QuerySigningInfosRequest{},
	)
	s.Require().NoError(err)

	// Check that not modified signing infos are the same
	s.Require().Equal(prevSigningInfos.Info, postSigningInfos.Info)
}
