package integration

import (
	"math/rand"
	"time"

	sdkmath "cosmossdk.io/math"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	types1 "github.com/cosmos/cosmos-sdk/codec/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	poatypes "github.com/xrplevm/node/v5/x/poa/types"
)

func (s *TestSuite) TestAddValidator_UnexistingValidator() {
	validator := s.Network().GetValidators()[0]
	valAddr, err := sdktypes.ValAddressFromBech32(validator.OperatorAddress)
	require.NoError(s.T(), err)

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
			name:       "add unexisting validator - random address - no balance",
			valAddress: randomAcc.Address.String(),
			afterRun: func() {

				require.NoError(s.T(), s.Network().NextBlock())

				resVal, err := s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: valAddr.String(),
					},
				)
				require.NoError(s.T(), err)

				// Check if the validator is unbonding
				require.Equal(s.T(), resVal.Validator.Status, stakingtypes.Bonded)
			},
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			if tc.beforeRun != nil {
				tc.beforeRun()
			}

			msgPubKey, _ := types1.NewAnyWithValue(randomAcc.ConsKey.PubKey())

			err := s.Network().PoaKeeper().ExecuteAddValidator(
				s.Network().GetContext(),
				&poatypes.MsgAddValidator{
					ValidatorAddress: randomAcc.Address.String(),
					Authority:        authtypes.NewModuleAddress(govtypes.ModuleName).String(),
					Pubkey:           msgPubKey,
					Description: stakingtypes.Description{
						Moniker: "test",
					},
				},
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

func (s *TestSuite) TestAddValidator_ExistingValidator_StatusBonded() {
	validator := s.Network().GetValidators()[0]
	valAddr, err := sdktypes.ValAddressFromBech32(validator.OperatorAddress)
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
			name:          "add existing validator - status bonded",
			valAddress:    valAddr.String(),
			expectedError: poatypes.ErrAddressHasBondedTokens,
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
			},
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			if tc.beforeRun != nil {
				tc.beforeRun()
			}

			err := s.Network().PoaKeeper().ExecuteAddValidator(
				s.Network().GetContext(),
				&poatypes.MsgAddValidator{
					ValidatorAddress: valAccAddr.String(),
					Authority:        authtypes.NewModuleAddress(govtypes.ModuleName).String(),
					Pubkey:           validator.ConsensusPubkey,
					Description: stakingtypes.Description{
						Moniker: "test",
					},
				},
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
			name:       "remove existing validator - status bonded",
			valAddress: valAccAddr.String(),
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

	tt := []struct {
		name          string
		valAddress    string
		expectedError error
		beforeRun     func()
		afterRun      func()
	}{
		{
			name:       "remove existing validator - jailed",
			valAddress: valAccAddr.String(),
			beforeRun: func() {
				// Force jail validator
				valSet := s.Network().GetValidatorSet()
				// Exclude validator at index 1 from validator set
				require.Equal(s.T(), sdktypes.ValAddress(valSet.Validators[1].Address).String(), valAddr.String())
				vf := make([]cmtproto.BlockIDFlag, len(valSet.Validators))
				for i := range valSet.Validators {
					vf[i] = cmtproto.BlockIDFlagCommit
				}
				vf[1] = cmtproto.BlockIDFlagAbsent

				require.NoError(s.T(), s.Network().NextNBlocksWithValidatorFlags(slashingtypes.DefaultSignedBlocksWindow+10, vf))

				resVal, err := s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: valAddr.String(),
					},
				)
				require.NoError(s.T(), err)

				// Check if the validator is jailed
				require.Equal(s.T(), resVal.Validator.Jailed, true)

				// Check if the validator is unbonding
				require.Equal(s.T(), resVal.Validator.Status, stakingtypes.Unbonding)
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
			name:       "remove existing validator - tombstoned",
			valAddress: valAccAddr.String(),
			beforeRun: func() {
				// Force validator to be tombstoned
				require.NoError(s.T(), s.Network().NextBlockWithMisBehaviors(
					[]abcitypes.Misbehavior{
						{
							Type: abcitypes.MisbehaviorType_DUPLICATE_VOTE,
							Validator: abcitypes.Validator{
								Address: valAddr.Bytes(),
							},
							Height:           s.Network().GetContext().BlockHeight(),
							TotalVotingPower: s.Network().GetValidatorSet().TotalVotingPower(),
						},
					},
				))

				resVal, err := s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: valAddr.String(),
					},
				)
				require.NoError(s.T(), err)

				// Check if the validator is jailed
				require.Equal(s.T(), resVal.Validator.Jailed, true)

				// Check if the validator is unbonding
				require.Equal(s.T(), resVal.Validator.Status, stakingtypes.Unbonding)

				info, err := s.Network().GetSlashingClient().SigningInfo(
					s.Network().GetContext(),
					&slashingtypes.QuerySigningInfoRequest{
						ConsAddress: valConsAddr.String(),
					},
				)
				require.NoError(s.T(), err)

				// Check if the validator is tombstoned
				require.Equal(s.T(), info.ValSigningInfo.Tombstoned, true)
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

func (s *TestSuite) TestRemoveValidator_ExistingValidator_StatusUnbonded() {

	// Validators
	validators := s.Network().GetValidators()
	require.NotZero(s.T(), len(validators))

	valAddr, err := sdktypes.ValAddressFromBech32(validators[1].OperatorAddress)
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
			name:          "remove existing validator - status unbonded",
			valAddress:    valAccAddr.String(),
			expectedError: poatypes.ErrAddressHasNoTokens,
			beforeRun: func() {
				// Force jail validator
				valSet := s.Network().GetValidatorSet()
				// Exclude validator at index 1 from validator set
				require.Equal(s.T(), sdktypes.ValAddress(valSet.Validators[1].Address).String(), valAddr.String())
				vf := make([]cmtproto.BlockIDFlag, len(valSet.Validators))
				for i := range valSet.Validators {
					vf[i] = cmtproto.BlockIDFlagCommit
				}
				vf[1] = cmtproto.BlockIDFlagAbsent

				require.NoError(s.T(), s.Network().NextNBlocksWithValidatorFlags(slashingtypes.DefaultSignedBlocksWindow+10, vf))

				resVal, err := s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: valAddr.String(),
					},
				)
				require.NoError(s.T(), err)

				// Check if the validator is jailed
				require.Equal(s.T(), resVal.Validator.Jailed, true)

				// Check if the validator is unbonding
				require.Equal(s.T(), resVal.Validator.Status, stakingtypes.Unbonding)

				require.NoError(s.T(), s.Network().NextBlockAfter(stakingtypes.DefaultUnbondingTime))
				require.NoError(s.T(), s.Network().NextBlock())

				resVal, err = s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: valAddr.String(),
					},
				)
				require.NoError(s.T(), err)

				// Check if the validator is unbonded
				require.Equal(s.T(), resVal.Validator.Status, stakingtypes.Unbonded)
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

func (s *TestSuite) TestRemoveValidator_ExistingValidator_StatusUnbonding() {
	// Validators
	validators := s.Network().GetValidators()
	require.NotZero(s.T(), len(validators))

	valAddr, err := sdktypes.ValAddressFromBech32(validators[1].OperatorAddress)
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
			name:          "remove existing validator - status unbonded",
			valAddress:    valAccAddr.String(),
			expectedError: poatypes.ErrAddressHasNoTokens,
			beforeRun: func() {
				// Force jail validator
				valSet := s.Network().GetValidatorSet()
				// Exclude validator at index 1 from validator set
				require.Equal(s.T(), sdktypes.ValAddress(valSet.Validators[1].Address).String(), valAddr.String())
				vf := make([]cmtproto.BlockIDFlag, len(valSet.Validators))
				for i := range valSet.Validators {
					vf[i] = cmtproto.BlockIDFlagCommit
				}
				vf[1] = cmtproto.BlockIDFlagAbsent

				require.NoError(s.T(), s.Network().NextNBlocksWithValidatorFlags(slashingtypes.DefaultSignedBlocksWindow+10, vf))

				resVal, err := s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: valAddr.String(),
					},
				)
				require.NoError(s.T(), err)

				// Check if the validator is jailed
				require.Equal(s.T(), resVal.Validator.Jailed, true)

				// Check if the validator is unbonding
				require.Equal(s.T(), resVal.Validator.Status, stakingtypes.Unbonding)
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
