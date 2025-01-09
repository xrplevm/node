package integration

import (
	"math/rand"
	"time"

	sdkmath "cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	poatypes "github.com/xrplevm/node/v5/x/poa/types"
)

func (s *TestSuite) TestRemoveValidator_UnexistingValidator() {
	// Generate a random account
	randomAccs := simtypes.RandomAccounts(rand.New(rand.NewSource(time.Now().UnixNano())), 1)
	randomAcc := randomAccs[0]

	tt := []struct {
		name          string
		valAddress    string
		expectedError error
		beforeRun     func()
		afterRun      func()
	}{
		{
			name:          "remove unexisting validator - random address - no balance",
			valAddress:    randomAcc.Address.String(),
			expectedError: poatypes.ErrAddressHasNoTokens,
			beforeRun: func() {
				balance, err := s.Network().GetBankClient().Balance(
					s.Network().GetContext(),
					&banktypes.QueryBalanceRequest{
						Address: randomAcc.Address.String(),
						Denom:   s.Network().GetBondDenom(),
					},
				)
				require.NoError(s.T(), err)

				// Check account has no balance
				require.True(s.T(), balance.Balance.Amount.IsZero())
			},
		},
		{
			name:          "remove unexisting validator - random address - with balance",
			valAddress:    randomAcc.Address.String(),
			expectedError: poatypes.ErrAddressHasNoTokens,
			beforeRun: func() {
				coins := sdktypes.NewCoin(
					s.Network().GetDenom(),
					sdkmath.NewInt(100000000000),
				)

				err := s.factory.FundAccount(
					s.keyring.GetKey(0),
					randomAcc.Address,
					sdktypes.NewCoins(
						coins,
					),
				)
				require.NoError(s.T(), err)
				require.NoError(s.T(), s.Network().NextBlock())

				balance, err := s.Network().GetBankClient().Balance(
					s.Network().GetContext(),
					&banktypes.QueryBalanceRequest{
						Address: randomAcc.Address.String(),
						Denom:   s.Network().GetBondDenom(),
					},
				)
				require.NoError(s.T(), err)

				// Check account has been funded
				require.True(s.T(), balance.Balance.Amount.IsZero())
			},
			afterRun: func() {
				balance, err := s.Network().GetBankClient().Balance(
					s.Network().GetContext(),
					&banktypes.QueryBalanceRequest{
						Address: randomAcc.Address.String(),
						Denom:   s.Network().GetBondDenom(),
					},
				)
				require.NoError(s.T(), err)

				// Check account has been emptied
				require.True(s.T(), balance.Balance.Amount.IsZero())
			},
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			if tc.beforeRun != nil {
				tc.beforeRun()
			}

			err := s.Network().PoaKeeper().ExecuteRemoveValidator(
				s.Network().GetContext(),
				tc.valAddress,
			)

			if tc.expectedError != nil && err != nil {
				require.Error(s.T(), err)
				require.ErrorIs(s.T(), err, tc.expectedError)
			} else {
				require.NoError(s.T(), err)
			}

			if tc.afterRun != nil {
				tc.afterRun()
			}
		})
	}
}

func (s *TestSuite) TestRemoveValidator_ExistingValidator_StatusBonded() {
	// Validators
	validators := s.Network().GetValidators()
	require.NotZero(s.T(), len(validators))

	valAddr, err := sdktypes.ValAddressFromBech32(validators[0].OperatorAddress)
	require.NoError(s.T(), err)
	valAccAddr := sdktypes.AccAddress(valAddr)

	tt := []struct {
		name          string
		valAddress    string
		expectedError error
		beforeRun     func()
		afterRun      func()
	}{
		{
			name:          "remove existing validator - status bonded",
			valAddress:    valAccAddr.String(),
			beforeRun: func() {
				resVal, err := s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: valAddr.String(),
					},
				)
				require.NoError(s.T(), err)

				// Check if the validator is bonded
				require.Equal(s.T(), resVal.Validator.Status, stakingtypes.Bonded)

				// Check if the validator has delegator shares
				require.False(s.T(), resVal.Validator.DelegatorShares.IsZero())

				// Check if the validator has tokens
				require.NotZero(s.T(), resVal.Validator.Tokens)
			},
			afterRun: func() {
				balance, err := s.Network().GetBankClient().Balance(
					s.Network().GetContext(),
					&banktypes.QueryBalanceRequest{
						Address: valAccAddr.String(),
						Denom:   s.Network().GetDenom(),
					},
				)
				require.NoError(s.T(), err)

				// Check account has been emptied
				require.True(s.T(), balance.Balance.Amount.IsZero())

				resVal, err := s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: valAddr.String(),
					},
				)
				require.NoError(s.T(), err)

				// Check if the validator has delegator shares
				require.True(s.T(), resVal.Validator.DelegatorShares.IsZero())

				// Check if the validator has no tokens
				require.True(s.T(), resVal.Validator.Tokens.IsZero())
			},
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			if tc.beforeRun != nil {
				tc.beforeRun()
			}

			err = s.Network().PoaKeeper().ExecuteRemoveValidator(
				s.Network().GetContext(),
				tc.valAddress,
			)

			if tc.expectedError != nil && err != nil {
				require.Error(s.T(), err)
				require.ErrorIs(s.T(), err, tc.expectedError)
			} else {
				require.NoError(s.T(), err)
			}

			if tc.afterRun != nil {
				tc.afterRun()
			}
		})
	}
}

func (s *TestSuite) TestRemoveValidator_ExistingValidator_Jailed() {
	// Validators
	validators := s.Network().GetValidators()
	require.NotZero(s.T(), len(validators))

	validator := validators[1]
	valAddr, err := sdktypes.ValAddressFromBech32(validator.OperatorAddress)
	require.NoError(s.T(), err)
	valAccAddr := sdktypes.AccAddress(valAddr)
	valConsAddr := sdktypes.ConsAddress(valAddr)

	tt := []struct {
		name          string
		valAddress    string
		expectedError error
		beforeRun     func()
		afterRun      func()
	}{
		{
			name:          "remove existing validator - jailed",
			valAddress:    valAccAddr.String(),
			beforeRun: func() {
				// Jail validator
				err := s.Network().StakingKeeper().Jail(
					s.Network().GetContext(),
					valConsAddr,
				)
				require.NoError(s.T(), err)

				resVal, err := s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: valAddr.String(),
					},
				)
				require.NoError(s.T(), err)

				// Check if the validator is jailed
				require.Equal(s.T(), resVal.Validator.Jailed, true)
			},
			afterRun: func() {
				// Check if the validator is jailed
				balance, err := s.Network().GetBankClient().Balance(
					s.Network().GetContext(),
					&banktypes.QueryBalanceRequest{
						Address: valAccAddr.String(),
						Denom:   s.Network().GetBondDenom(),
					},
				)
				require.NoError(s.T(), err)

				// Check account has been emptied
				require.True(s.T(), balance.Balance.Amount.IsZero())

				resVal, err := s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: valAddr.String(),
					},
				)
				require.NoError(s.T(), err)

				// Check if the validator has delegator shares
				require.True(s.T(), resVal.Validator.DelegatorShares.IsZero())

				// Check if the validator has no tokens
				require.True(s.T(), resVal.Validator.Tokens.IsZero())
			},
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			if tc.beforeRun != nil {
				tc.beforeRun()
			}

			err = s.Network().PoaKeeper().ExecuteRemoveValidator(
				s.Network().GetContext(),
				tc.valAddress,
			)

			if tc.expectedError != nil && err != nil {
				require.Error(s.T(), err)
				require.ErrorIs(s.T(), err, tc.expectedError)
			} else {
				require.NoError(s.T(), err)
			}

			if tc.afterRun != nil {
				tc.afterRun()
			}
		})
	}
}

func (s *TestSuite) TestRemoveValidator_ExistingValidator_Tombstoned() {
	// Validators
	validators := s.Network().GetValidators()
	require.NotZero(s.T(), len(validators))

	validator := validators[1]
	valAddr, err := sdktypes.ValAddressFromBech32(validator.OperatorAddress)
	require.NoError(s.T(), err)
	valAccAddr := sdktypes.AccAddress(valAddr)
	valConsAddr := sdktypes.ConsAddress(valAddr)

	tt := []struct {
		name          string
		valAddress    string
		expectedError error
		beforeRun     func()
		afterRun      func()
	}{
		{
			name:          "remove existing validator - tombstoned",
			valAddress:    valAccAddr.String(),
			beforeRun: func() {
				// Jail validator
				err := s.Network().StakingKeeper().Jail(
					s.Network().GetContext(),
					valConsAddr,
				)
				require.NoError(s.T(), err)

				resVal, err := s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: valAddr.String(),
					},
				)
				require.NoError(s.T(), err)

				// Check if the validator is jailed
				require.Equal(s.T(), resVal.Validator.Jailed, true)

				err = s.Network().SlashingKeeper().Tombstone(
					s.Network().GetContext(),
					valConsAddr,
				)
				require.NoError(s.T(), err)
			},
			afterRun: func() {
				balance, err := s.Network().GetBankClient().Balance(
					s.Network().GetContext(),
					&banktypes.QueryBalanceRequest{
						Address: valAccAddr.String(),
						Denom:   s.Network().GetBondDenom(),
					},
				)
				require.NoError(s.T(), err)

				// Check account has been emptied
				require.True(s.T(), balance.Balance.Amount.IsZero())

				resVal, err := s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: valAddr.String(),
					},
				)
				require.NoError(s.T(), err)

				// Check if the validator has delegator shares
				require.True(s.T(), resVal.Validator.DelegatorShares.IsZero())

				// Check if the validator has no tokens
				require.True(s.T(), resVal.Validator.Tokens.IsZero())
			},
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			if tc.beforeRun != nil {
				tc.beforeRun()
			}

			err = s.Network().PoaKeeper().ExecuteRemoveValidator(
				s.Network().GetContext(),
				tc.valAddress,
			)

			if tc.expectedError != nil && err != nil {
				require.Error(s.T(), err)
				require.ErrorIs(s.T(), err, tc.expectedError)
			} else {
				require.NoError(s.T(), err)
			}

			if tc.afterRun != nil {
				tc.afterRun()
			}
		})
	}
}