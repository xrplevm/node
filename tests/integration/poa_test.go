package integration

import (
	"fmt"
	"math/rand"
	"time"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v9/testutil/integration/exrp/utils"
	poatypes "github.com/xrplevm/node/v9/x/poa/types"
)

// AddValidator tests

func (s *TestSuite) TestAddValidator_UnexistingValidator() {
	// Generate a random account
	randomAccs := simtypes.RandomAccounts(rand.New(rand.NewSource(time.Now().UnixNano())), 1) //nolint:gosec
	randomAcc := randomAccs[0]
	randomValAddr := sdktypes.ValAddress(randomAcc.Address.Bytes())

	tt := []struct {
		name          string
		valAddress    string
		valPubKey     cryptotypes.PubKey
		expectedError error
		beforeRun     func()
		afterRun      func()
	}{
		{
			name:       "add unexisting validator - random address",
			valAddress: randomAcc.Address.String(),
			valPubKey:  randomAcc.ConsKey.PubKey(),
			afterRun: func() {
				require.NoError(s.T(), s.Network().NextBlock())

				resVal, err := s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: randomValAddr.String(),
					},
				)
				require.NoError(s.T(), err)

				// Check if the validator is bonded
				require.Equal(s.T(), resVal.Validator.Status, stakingtypes.Bonded)

				// Check if the validator has the default amount of tokens
				require.Equal(s.T(), sdktypes.DefaultPowerReduction, resVal.Validator.Tokens)

				// Check if the validator has the default delegator shares
				require.Equal(s.T(), sdktypes.DefaultPowerReduction.ToLegacyDec(), resVal.Validator.DelegatorShares)
			},
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			if tc.beforeRun != nil {
				tc.beforeRun()
			}

			authority := sdktypes.AccAddress(address.Module("gov"))
			msg, err := poatypes.NewMsgAddValidator(
				authority.String(),
				tc.valAddress,
				tc.valPubKey,
				stakingtypes.Description{
					Moniker: "test",
				},
			)
			require.NoError(s.T(), err)

			proposal, err := utils.SubmitAndAwaitProposalResolution(s.factory, s.Network(), s.keyring.GetKeys(), "test", msg)
			require.NoError(s.T(), err)

			require.Equal(s.T(), govv1.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)

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

func (s *TestSuite) TestAddValidator_InvalidMsgAddValidator() {
	// Generate a random account
	randomAccs := simtypes.RandomAccounts(rand.New(rand.NewSource(time.Now().UnixNano())), 1) //nolint:gosec
	randomAcc := randomAccs[0]

	validator := s.Network().GetValidators()[0]
	valAddr, err := sdktypes.ValAddressFromBech32(validator.OperatorAddress)
	require.NoError(s.T(), err)
	valAccAddr := sdktypes.AccAddress(valAddr)

	tt := []struct {
		name          string
		valAddress    string
		valPubKey     cryptotypes.PubKey
		expectedError error
		beforeRun     func()
		afterRun      func()
	}{
		{
			name:          "add validator - already used pubkey",
			valAddress:    randomAcc.Address.String(),
			valPubKey:     validator.ConsensusPubkey.GetCachedValue().(cryptotypes.PubKey),
			expectedError: stakingtypes.ErrValidatorPubKeyExists,
		},
		{
			name:          "add validator - already used validator address",
			valAddress:    valAccAddr.String(),
			valPubKey:     randomAcc.ConsKey.PubKey(),
			expectedError: poatypes.ErrAddressHasBondedTokens,
			beforeRun: func() {
				// Check if the validator exists
				_, err := s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: valAddr.String(),
					},
				)
				require.NoError(s.T(), err)
			},
		},
	}

	//nolint:dupl
	for _, tc := range tt {
		s.Run(tc.name, func() {
			if tc.beforeRun != nil {
				tc.beforeRun()
			}

			authority := sdktypes.AccAddress(address.Module("gov"))
			msg, err := poatypes.NewMsgAddValidator(
				authority.String(),
				tc.valAddress,
				tc.valPubKey,
				stakingtypes.Description{
					Moniker: "test",
				},
			)
			require.NoError(s.T(), err)

			proposal, err := utils.SubmitAndAwaitProposalResolution(s.factory, s.Network(), s.keyring.GetKeys(), "test", msg)
			require.NoError(s.T(), err)

			require.Equal(s.T(), govv1.ProposalStatus_PROPOSAL_STATUS_FAILED, proposal.Status)
			require.Contains(s.T(), proposal.FailedReason, tc.expectedError.Error())

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
		valPubKey     cryptotypes.PubKey
		expectedError error
		beforeRun     func()
		afterRun      func()
	}{
		{
			name:          "add existing validator - status bonded",
			valAddress:    valAccAddr.String(),
			valPubKey:     validator.ConsensusPubkey.GetCachedValue().(cryptotypes.PubKey),
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
			afterRun: func() {
				resVal, err := s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: valAddr.String(),
					},
				)
				require.NoError(s.T(), err)

				// Check if the validator is still bonded
				require.Equal(s.T(), resVal.Validator.Status, stakingtypes.Bonded)
			},
		},
	}

	//nolint:dupl
	for _, tc := range tt {
		s.Run(tc.name, func() {
			if tc.beforeRun != nil {
				tc.beforeRun()
			}

			authority := sdktypes.AccAddress(address.Module("gov"))
			msg, err := poatypes.NewMsgAddValidator(
				authority.String(),
				tc.valAddress,
				tc.valPubKey,
				stakingtypes.Description{
					Moniker: "test",
				},
			)
			require.NoError(s.T(), err)

			proposal, err := utils.SubmitAndAwaitProposalResolution(s.factory, s.Network(), s.keyring.GetKeys(), "test", msg)
			require.NoError(s.T(), err)

			require.Equal(s.T(), govv1.ProposalStatus_PROPOSAL_STATUS_FAILED, proposal.Status)
			require.Contains(s.T(), proposal.FailedReason, tc.expectedError.Error())

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

func (s *TestSuite) TestAddValidator_ExistingValidator_Jailed() {
	valIndex := 1
	validator := s.Network().GetValidators()[valIndex]
	valAddr, err := sdktypes.ValAddressFromBech32(validator.OperatorAddress)
	require.NoError(s.T(), err)
	valAccAddr := sdktypes.AccAddress(valAddr)

	tt := []struct {
		name          string
		valAddress    string
		valPubKey     cryptotypes.PubKey
		expectedError error
		beforeRun     func()
		afterRun      func()
	}{
		{
			name:          "add existing validator - status jailed",
			valAddress:    valAccAddr.String(),
			valPubKey:     validator.ConsensusPubkey.GetCachedValue().(cryptotypes.PubKey),
			expectedError: poatypes.ErrAddressHasBondedTokens,
			beforeRun: func() {
				// Force jail validator
				valSet := s.Network().GetValidatorSet()

				require.NoError(
					s.T(),
					s.Network().NextNBlocksWithValidatorFlags(
						slashingtypes.DefaultSignedBlocksWindow,
						utils.NewValidatorFlags(
							len(valSet.Validators),
							utils.NewValidatorFlagOverride(valIndex, cmtproto.BlockIDFlagAbsent),
						),
					),
				)

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
				// Check if the validator is still jailed
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

	//nolint:dupl
	for _, tc := range tt {
		s.Run(tc.name, func() {
			if tc.beforeRun != nil {
				tc.beforeRun()
			}

			authority := sdktypes.AccAddress(address.Module("gov"))
			msg, err := poatypes.NewMsgAddValidator(
				authority.String(),
				tc.valAddress,
				tc.valPubKey,
				stakingtypes.Description{
					Moniker: "test",
				},
			)
			require.NoError(s.T(), err)

			proposal, err := utils.SubmitAndAwaitProposalResolution(s.factory, s.Network(), s.keyring.GetKeys(), "test", msg)
			require.NoError(s.T(), err)

			require.Equal(s.T(), govv1.ProposalStatus_PROPOSAL_STATUS_FAILED, proposal.Status)
			require.Contains(s.T(), proposal.FailedReason, tc.expectedError.Error())

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

func (s *TestSuite) TestAddValidator_ExistingValidator_Tombstoned() {
	valIndex := 1

	// CometBFT validators
	valSet := s.Network().GetValidatorSet()
	cmtValAddr := sdktypes.AccAddress(valSet.Validators[valIndex].Address.Bytes())
	cmtValConsAddr := sdktypes.ConsAddress(valSet.Validators[valIndex].Address.Bytes())

	// Cosmos validators
	validators := s.Network().GetValidators()
	require.NotZero(s.T(), len(validators))

	validator := validators[valIndex]
	valAddr, err := sdktypes.ValAddressFromBech32(validator.OperatorAddress)
	require.NoError(s.T(), err)
	valAccAddr := sdktypes.AccAddress(valAddr)

	tt := []struct {
		name          string
		valAddress    string
		valPubKey     cryptotypes.PubKey
		expectedError error
		beforeRun     func()
		afterRun      func()
	}{
		{
			name:          "add existing validator - status tombstoned",
			valAddress:    valAccAddr.String(),
			valPubKey:     validator.ConsensusPubkey.GetCachedValue().(cryptotypes.PubKey),
			expectedError: poatypes.ErrAddressHasBondedTokens,
			beforeRun: func() {
				// Force validator to be tombstoned
				require.NoError(s.T(), s.Network().NextBlockWithMisBehaviors(
					[]abcitypes.Misbehavior{
						{
							Type: abcitypes.MisbehaviorType_DUPLICATE_VOTE,
							Validator: abcitypes.Validator{
								Address: cmtValAddr.Bytes(),
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
						ConsAddress: cmtValConsAddr.String(),
					},
				)
				require.NoError(s.T(), err)

				// Check if the validator is tombstoned
				require.Equal(s.T(), info.ValSigningInfo.Tombstoned, true)
			},
			afterRun: func() {
				// Check if the validator is still jailed
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

	//nolint:dupl
	for _, tc := range tt {
		s.Run(tc.name, func() {
			if tc.beforeRun != nil {
				tc.beforeRun()
			}

			authority := sdktypes.AccAddress(address.Module("gov"))
			msg, err := poatypes.NewMsgAddValidator(
				authority.String(),
				tc.valAddress,
				tc.valPubKey,
				stakingtypes.Description{
					Moniker: "test",
				},
			)
			require.NoError(s.T(), err)

			proposal, err := utils.SubmitAndAwaitProposalResolution(s.factory, s.Network(), s.keyring.GetKeys(), "test", msg)
			require.NoError(s.T(), err)

			require.Equal(s.T(), govv1.ProposalStatus_PROPOSAL_STATUS_FAILED, proposal.Status)
			require.Contains(s.T(), proposal.FailedReason, tc.expectedError.Error())

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

func (s *TestSuite) TestAddValidator_MaximumValidators() {
	// Generate a random account
	randomAccs := simtypes.RandomAccounts(rand.New(rand.NewSource(time.Now().UnixNano())), 1) //nolint:gosec
	randomAcc := randomAccs[0]
	randomValAddr := sdktypes.ValAddress(randomAcc.Address.Bytes())

	tt := []struct {
		name          string
		valAddress    string
		valPubKey     cryptotypes.PubKey
		expectedError error
		beforeRun     func()
		afterRun      func()
	}{
		{
			name:          "add validator - maximum validators reached",
			valAddress:    randomAcc.Address.String(),
			valPubKey:     randomAcc.PubKey,
			expectedError: poatypes.ErrMaxValidatorsReached,
			beforeRun: func() {
				resVal, err := s.Network().GetStakingClient().Params(
					s.Network().GetContext(),
					&stakingtypes.QueryParamsRequest{},
				)
				s.Require().NoError(err)
				amountOfValidators := uint32(5)
				maxValidators := resVal.Params.MaxValidators
				authority := sdktypes.AccAddress(address.Module("gov")).String()

				for i := uint32(0); i < maxValidators-amountOfValidators; i++ {
					randomValidator := simtypes.RandomAccounts(rand.New(rand.NewSource(time.Now().UnixNano())), 1) //nolint:gosec
					randomValidatorAcc := randomValidator[0]
					msg, err := poatypes.NewMsgAddValidator(
						authority,
						randomValidatorAcc.Address.String(),
						randomValidatorAcc.ConsKey.PubKey(),
						stakingtypes.Description{
							Moniker: "test",
						},
					)
					require.NoError(s.T(), err)
					proposal, err := utils.SubmitAndAwaitProposalResolution(s.factory, s.Network(), s.keyring.GetKeys(), "test", msg)
					require.NoError(s.T(), err)
					require.Equal(s.T(), govv1.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)
				}
			},
			afterRun: func() {
				// Check validator not added
				_, err := s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: randomValAddr.String(),
					},
				)
				require.Error(s.T(), err)
			},
		},
	}

	//nolint:dupl
	for _, tc := range tt {
		s.Run(tc.name, func() {
			if tc.beforeRun != nil {
				tc.beforeRun()
			}

			authority := sdktypes.AccAddress(address.Module("gov"))
			msg, err := poatypes.NewMsgAddValidator(
				authority.String(),
				tc.valAddress,
				tc.valPubKey,
				stakingtypes.Description{
					Moniker: "test",
				},
			)
			require.NoError(s.T(), err)

			proposal, err := utils.SubmitAndAwaitProposalResolution(s.factory, s.Network(), s.keyring.GetKeys(), "test", msg)
			require.NoError(s.T(), err)

			require.Equal(s.T(), govv1.ProposalStatus_PROPOSAL_STATUS_FAILED, proposal.Status)
			require.Contains(s.T(), proposal.FailedReason, tc.expectedError.Error())

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

// RemoveValidator tests

func (s *TestSuite) TestRemoveValidator_UnexistingValidator() {
	// Generate a random account
	randomAccs := simtypes.RandomAccounts(rand.New(rand.NewSource(time.Now().UnixNano())), 1) //nolint:gosec

	randomValAddr := sdktypes.ValAddress(randomAccs[0].Address.Bytes())

	tt := []struct {
		name          string
		valAddress    string
		expectedError error
		beforeRun     func()
		afterRun      func()
	}{
		{
			name:          "remove unexisting validator - random address - with balance",
			valAddress:    randomValAddr.String(),
			expectedError: poatypes.ErrAddressIsNotAValidator,
			beforeRun: func() {
				_, err := s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: randomValAddr.String(),
					},
				)

				// Check if the validator does not exist
				require.Contains(s.T(), err.Error(), fmt.Sprintf("validator %s not found", randomValAddr.String()))
			},
			afterRun: func() {
				_, err := s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: randomValAddr.String(),
					},
				)

				// Check if the validator is not found
				require.Contains(s.T(), err.Error(), fmt.Sprintf("validator %s not found", randomValAddr.String()))
			},
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			if tc.beforeRun != nil {
				tc.beforeRun()
			}

			authority := sdktypes.AccAddress(address.Module("gov"))
			msg := poatypes.NewMsgRemoveValidator(
				authority.String(),
				tc.valAddress,
			)

			proposal, err := utils.SubmitAndAwaitProposalResolution(s.factory, s.Network(), s.keyring.GetKeys(), "test", msg)
			require.NoError(s.T(), err)

			require.Equal(s.T(), govv1.ProposalStatus_PROPOSAL_STATUS_FAILED, proposal.Status)
			require.Equal(s.T(), tc.expectedError.Error(), proposal.FailedReason)

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

	validator := validators[0]
	valAddr, err := sdktypes.ValAddressFromBech32(validator.OperatorAddress)
	require.NoError(s.T(), err)

	tt := []struct {
		name          string
		valAddress    string
		expectedError error
		beforeRun     func()
		afterRun      func()
	}{
		{
			name:       "remove existing validator - status bonded",
			valAddress: valAddr.String(),
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
				require.Equal(s.T(), sdktypes.DefaultPowerReduction.ToLegacyDec(), resVal.Validator.DelegatorShares)

				// Check if the validator has tokens
				require.NotZero(s.T(), resVal.Validator.Tokens)
			},
			afterRun: func() {
				resVal, err := s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: valAddr.String(),
					},
				)
				require.NoError(s.T(), err)

				// Check if the validator has delegator shares
				require.True(s.T(),
					resVal.Validator.DelegatorShares.IsZero(),
					"delegator shares should be zero, got %s",
					resVal.Validator.DelegatorShares,
				)

				// Check if the validator has no tokens
				require.True(s.T(), resVal.Validator.Tokens.IsZero())

				resVal, err = s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: valAddr.String(),
					},
				)
				require.NoError(s.T(), err)

				// Check if the validator is unbonded
				require.Equal(s.T(), resVal.Validator.Status, stakingtypes.Unbonding)

				require.NoError(s.T(), s.Network().NextBlockAfter(stakingtypes.DefaultUnbondingTime))

				_, err = s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: valAddr.String(),
					},
				)
				require.Contains(s.T(), err.Error(), fmt.Sprintf("validator %s not found", valAddr.String()))
			},
		},
	}

	//nolint:dupl
	for _, tc := range tt {
		s.Run(tc.name, func() {
			if tc.beforeRun != nil {
				tc.beforeRun()
			}

			authority := sdktypes.AccAddress(address.Module("gov"))
			msg := poatypes.NewMsgRemoveValidator(
				authority.String(),
				valAddr.String(),
			)

			proposal, err := utils.SubmitAndAwaitProposalResolution(s.factory, s.Network(), s.keyring.GetKeys(), "test", msg)
			require.NoError(s.T(), err)

			require.Equal(s.T(), govv1.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)

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

	valIndex := 1
	validator := validators[valIndex]
	valAddr, err := sdktypes.ValAddressFromBech32(validator.OperatorAddress)
	require.NoError(s.T(), err)

	tt := []struct {
		name          string
		valAddress    string
		expectedError error
		beforeRun     func()
		afterRun      func()
	}{
		{
			name:       "remove existing validator - jailed",
			valAddress: valAddr.String(),
			beforeRun: func() {
				// Force jail validator
				valSet := s.Network().GetValidatorSet()

				require.NoError(
					s.T(),
					s.Network().NextNBlocksWithValidatorFlags(
						slashingtypes.DefaultSignedBlocksWindow,
						utils.NewValidatorFlags(
							len(valSet.Validators),
							utils.NewValidatorFlagOverride(valIndex, cmtproto.BlockIDFlagAbsent),
						),
					),
				)

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

				// Check if the validator is unbonding
				require.Equal(s.T(), resVal.Validator.Status, stakingtypes.Unbonding)

				require.NoError(s.T(), s.Network().NextBlockAfter(stakingtypes.DefaultUnbondingTime))

				_, err = s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: valAddr.String(),
					},
				)

				require.Contains(s.T(), err.Error(), fmt.Sprintf("validator %s not found", valAddr.String()))
			},
		},
	}

	//nolint:dupl
	for _, tc := range tt {
		s.Run(tc.name, func() {
			if tc.beforeRun != nil {
				tc.beforeRun()
			}

			authority := sdktypes.AccAddress(address.Module("gov"))
			msg := poatypes.NewMsgRemoveValidator(
				authority.String(),
				valAddr.String(),
			)

			proposal, err := utils.SubmitAndAwaitProposalResolution(s.factory, s.Network(), s.keyring.GetKeys(), "test", msg)
			require.NoError(s.T(), err)

			require.Equal(s.T(), govv1.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)

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
	valIndex := 1

	// CometBFT validators
	valSet := s.Network().GetValidatorSet()
	cmtValAddr := sdktypes.AccAddress(valSet.Validators[valIndex].Address.Bytes())
	cmtValConsAddr := sdktypes.ConsAddress(valSet.Validators[valIndex].Address.Bytes())

	// Cosmos validators
	validators := s.Network().GetValidators()
	require.NotZero(s.T(), len(validators))

	validator := validators[valIndex]
	valAddr, err := sdktypes.ValAddressFromBech32(validator.OperatorAddress)
	require.NoError(s.T(), err)

	tt := []struct {
		name          string
		valAddress    string
		expectedError error
		beforeRun     func()
		afterRun      func()
	}{
		{
			name:       "remove existing validator - tombstoned",
			valAddress: valAddr.String(),
			beforeRun: func() {
				// Force validator to be tombstoned
				require.NoError(s.T(), s.Network().NextBlockWithMisBehaviors(
					[]abcitypes.Misbehavior{
						{
							Type: abcitypes.MisbehaviorType_DUPLICATE_VOTE,
							Validator: abcitypes.Validator{
								Address: cmtValAddr,
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
						ConsAddress: cmtValConsAddr.String(),
					},
				)
				require.NoError(s.T(), err)

				// Check if the validator is tombstoned
				require.Equal(s.T(), info.ValSigningInfo.Tombstoned, true)
			},
			afterRun: func() {
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

				// Check if the validator is unbonding
				require.Equal(s.T(), resVal.Validator.Status, stakingtypes.Unbonding)

				// Await unbonding time to pass
				require.NoError(s.T(), s.Network().NextBlockAfter(stakingtypes.DefaultUnbondingTime))

				_, err = s.Network().GetStakingClient().Validator(
					s.Network().GetContext(),
					&stakingtypes.QueryValidatorRequest{
						ValidatorAddr: valAddr.String(),
					},
				)

				// Check if the validator is not found
				require.Contains(s.T(), err.Error(), fmt.Sprintf("validator %s not found", valAddr.String()))

				info, err := s.Network().GetSlashingClient().SigningInfo(
					s.Network().GetContext(),
					&slashingtypes.QuerySigningInfoRequest{
						ConsAddress: cmtValConsAddr.String(),
					},
				)
				require.NoError(s.T(), err)

				// Check if the validator is tombstoned
				require.True(s.T(), info.ValSigningInfo.Tombstoned)
			},
		},
	}

	//nolint:dupl
	for _, tc := range tt {
		s.Run(tc.name, func() {
			if tc.beforeRun != nil {
				tc.beforeRun()
			}

			authority := sdktypes.AccAddress(address.Module("gov"))
			msg := poatypes.NewMsgRemoveValidator(
				authority.String(),
				valAddr.String(),
			)

			proposal, err := utils.SubmitAndAwaitProposalResolution(s.factory, s.Network(), s.keyring.GetKeys(), "test", msg)
			require.NoError(s.T(), err)

			require.Equal(s.T(), govv1.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)

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
