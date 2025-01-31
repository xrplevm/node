package integration

import (
	"time"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *UpgradeTestSuite) TestUpgrade_Staking_Params() {
	prevParams, err := s.network.GetStakingClient().Params(
		s.network.GetContext(),
		&stakingtypes.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	postParams, err := s.network.GetStakingClient().Params(
		s.network.GetContext(),
		&stakingtypes.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	// Check that not modified params are the same
	s.Require().Equal(prevParams.Params.BondDenom, postParams.Params.BondDenom)
	s.Require().Equal(prevParams.Params.MaxValidators, postParams.Params.MaxValidators)
	s.Require().Equal(prevParams.Params.MinCommissionRate, postParams.Params.MinCommissionRate)
	s.Require().Equal(prevParams.Params.MaxEntries, postParams.Params.MaxEntries)
	s.Require().Equal(prevParams.Params.HistoricalEntries, postParams.Params.HistoricalEntries)

	// Check that unbonding time was modified
	s.Require().Equal(postParams.Params.UnbondingTime, 100*time.Second)
}

func (s *UpgradeTestSuite) TestUpgrade_Staking_Validators() {
	res, err := s.network.GetStakingClient().Validators(
		s.network.GetContext(),
		&stakingtypes.QueryValidatorsRequest{},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	postRes, err := s.network.GetStakingClient().Validators(
		s.network.GetContext(),
		&stakingtypes.QueryValidatorsRequest{},
	)
	s.Require().NoError(err)

	// Check that not modified validators are the same
	s.Require().Equal(res.Validators, postRes.Validators)
}

func (s *UpgradeTestSuite) TestUpgrade_Staking_Delegations() {
	prevDelegations, err := s.network.StakingKeeper().GetAllDelegations(
		s.network.GetContext(),
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	postDelegations, err := s.network.StakingKeeper().GetAllDelegations(
		s.network.GetContext(),
	)
	s.Require().NoError(err)

	// Check that not modified delegations are the same
	s.Require().Equal(prevDelegations, postDelegations)
}

func (s *UpgradeTestSuite) TestUpgrade_Staking_UnbondingDelegations() {
	prevValidators, err := s.network.StakingKeeper().GetAllValidators(
		s.network.GetContext(),
	)
	s.Require().NoError(err)

	unbondingDelegations := make(map[string][]stakingtypes.UnbondingDelegation, len(prevValidators))
	for _, validator := range prevValidators {
		res, err := s.network.GetStakingClient().ValidatorUnbondingDelegations(
			s.network.GetContext(),
			&stakingtypes.QueryValidatorUnbondingDelegationsRequest{
				ValidatorAddr: validator.OperatorAddress,
			},
		)
		s.Require().NoError(err)
		unbondingDelegations[validator.OperatorAddress] = res.UnbondingResponses
	}

	s.RunUpgrade(upgradeName)

	postUnbondingDelegations := make(map[string][]stakingtypes.UnbondingDelegation, len(prevValidators))
	for _, validator := range prevValidators {
		res, err := s.network.GetStakingClient().ValidatorUnbondingDelegations(
			s.network.GetContext(),
			&stakingtypes.QueryValidatorUnbondingDelegationsRequest{
				ValidatorAddr: validator.OperatorAddress,
			},
		)
		s.Require().NoError(err)
		postUnbondingDelegations[validator.OperatorAddress] = res.UnbondingResponses
	}

	// Check that not modified unbonding delegations are the same
	s.Require().Equal(unbondingDelegations, postUnbondingDelegations)
}
