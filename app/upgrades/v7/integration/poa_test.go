package integration

import (
	"math/rand"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	poatypes "github.com/xrplevm/node/v6/x/poa/types"
)

func (s *UpgradeTestSuite) TestUpgrade_Poa_ExecuteRemoveValidator() {
	validators, err := s.network.GetStakingClient().Validators(
		s.network.GetContext(),
		&stakingtypes.QueryValidatorsRequest{},
	)
	s.Require().NoError(err)
	s.Require().Greater(len(validators.Validators), 0)

	validator := validators.Validators[0]
	valAddr, err := sdktypes.ValAddressFromBech32(validator.OperatorAddress)
	s.Require().NoError(err)
	valAccAddr := sdktypes.AccAddress(valAddr)

	_, err = s.network.GetStakingClient().Validator(
		s.network.GetContext(),
		&stakingtypes.QueryValidatorRequest{
			ValidatorAddr: validator.OperatorAddress,
		},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	err = s.network.PoaKeeper().ExecuteRemoveValidator(
		s.network.GetContext(),
		valAccAddr.String(),
	)
	s.Require().NoError(err)

	postValidator, err := s.network.GetStakingClient().Validator(
		s.network.GetContext(),
		&stakingtypes.QueryValidatorRequest{
			ValidatorAddr: validator.OperatorAddress,
		},
	)

	s.Require().NoError(err)
	s.Require().True(postValidator.Validator.Tokens.IsZero(), "validator tokens should be zero")
	s.Require().True(postValidator.Validator.DelegatorShares.RoundInt().IsZero(), "validator delegator shares should be zero")
}

func (s *UpgradeTestSuite) TestUpgrade_Poa_ExecuteAddValidator() {
	randomAccs := simtypes.RandomAccounts(rand.New(rand.NewSource(time.Now().UnixNano())), 1) //nolint:gosec
	randomAcc := randomAccs[0]
	randomValAddr := sdktypes.ValAddress(randomAcc.Address.Bytes())

	authority := sdktypes.AccAddress(address.Module("gov"))
	msg, err := poatypes.NewMsgAddValidator(
		authority.String(),
		randomAcc.Address.String(),
		randomAcc.ConsKey.PubKey(),
		stakingtypes.Description{
			Moniker: "test",
		},
	)
	s.Require().NoError(err)

	_, err = s.network.GetStakingClient().Validator(
		s.network.GetContext(),
		&stakingtypes.QueryValidatorRequest{
			ValidatorAddr: randomValAddr.String(),
		},
	)
	s.Require().Error(err)

	s.RunUpgrade(upgradeName)

	err = s.network.PoaKeeper().ExecuteAddValidator(
		s.network.GetContext(),
		msg,
	)
	s.Require().NoError(err)

	val, err := s.network.GetStakingClient().Validator(
		s.network.GetContext(),
		&stakingtypes.QueryValidatorRequest{
			ValidatorAddr: randomValAddr.String(),
		},
	)
	s.Require().NoError(err)
	s.Require().Equal(val.Validator.Status, stakingtypes.Unbonded)
	s.Require().Equal(val.Validator.Tokens, sdktypes.DefaultPowerReduction)
	s.Require().Equal(val.Validator.DelegatorShares, sdktypes.DefaultPowerReduction.ToLegacyDec())
}
