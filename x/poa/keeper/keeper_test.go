package keeper

import (
	"errors"
	"testing"

	"cosmossdk.io/math"
	types1 "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v10/x/poa/testutil"
	"github.com/xrplevm/node/v10/x/poa/types"
	"go.uber.org/mock/gomock"
)

// Define here Keeper methods to be unit tested
func TestKeeper_ExecuteAddValidator(t *testing.T) {
	ctrl := gomock.NewController(t)
	pubKey := testutil.NewMockPubKey(ctrl)
	msgPubKey, _ := types1.NewAnyWithValue(pubKey)

	tt := []struct {
		name             string
		validatorAddress string
		pubKey           *types1.Any
		stakingMocks     func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper)
		bankMocks        func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper)
		expectedError    error
	}{
		{
			name:             "should fail - invalid validator address",
			validatorAddress: "invalidnaddress",
			expectedError:    errors.New("decoding bech32 failed"),
			stakingMocks:     func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
			bankMocks:        func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		},
		{
			name:             "should fail - staking keeper returns error on GetParams",
			validatorAddress: validatorAccAddress,
			expectedError:    errors.New("staking params error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{}, errors.New("staking params error"))
			},
			bankMocks: func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		},
		{
			name:             "should fail - maximum validators reached",
			validatorAddress: validatorAccAddress,
			expectedError:    types.ErrMaxValidatorsReached,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom:     "BND",
					MaxValidators: 1,
				}, nil)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{{}}, nil)
			},
			bankMocks: func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		},
		{
			name:             "should fail - validator has bonded tokens",
			validatorAddress: validatorAccAddress,
			expectedError:    types.ErrAddressHasBankTokens,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom:     "BND",
					MaxValidators: 2,
				}, nil)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{}, nil)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: sdk.DefaultPowerReduction,
				})
			},
		},
		{
			name:             "should fail - staking keeper returns error on GetValidator",
			validatorAddress: validatorAccAddress,
			expectedError:    errors.New("staking validator error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom:     "BND",
					MaxValidators: 2,
				}, nil)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{}, errors.New("staking validator error"))
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
			},
		},
		{
			name:             "should fail - staking keeper returns validator with tokens",
			validatorAddress: validatorAccAddress,
			expectedError:    types.ErrAddressHasBondedTokens,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom:     "BND",
					MaxValidators: 2,
				}, nil)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: sdk.DefaultPowerReduction}, nil)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
			},
		},
		{
			name:             "should fail - staking keeper returns error on GetAllDelegatorDelegations",
			validatorAddress: validatorAccAddress,
			expectedError:    errors.New("staking delegations error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom:     "BND",
					MaxValidators: 2,
				}, nil)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, nil)
				stakingKeeper.EXPECT().GetAllDelegatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{}, errors.New("staking delegations error"))
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
			},
		},
		{
			name:             "should fail - delegations are greater than 0 with invalid delegation validator address",
			validatorAddress: validatorAccAddress,
			expectedError:    errors.New("decoding bech32 failed"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom:     "BND",
					MaxValidators: 2,
				}, nil)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, nil)
				stakingKeeper.EXPECT().GetAllDelegatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{
					{
						ValidatorAddress: "invalidvalidatoraddress",
						Shares:           sdk.DefaultPowerReduction.ToLegacyDec(),
					},
				}, nil)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
			},
		},
		{
			name:             "should fail - delegations are greater than 0 with error on GetValidator call",
			validatorAddress: validatorAccAddress,
			expectedError:    errors.New("staking validator error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom:     "BND",
					MaxValidators: 2,
				}, nil)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, nil).Times(1)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, errors.New("staking validator error")).Times(1)
				stakingKeeper.EXPECT().GetAllDelegatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{
					{
						ValidatorAddress: validatorValAddress,
						DelegatorAddress: validatorAccAddress,
						Shares:           sdk.DefaultPowerReduction.ToLegacyDec(),
					},
				}, nil)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
			},
		},
		{
			name:             "should fail - delegations are greater than 0 with delegated tokens",
			validatorAddress: validatorAccAddress,
			expectedError:    types.ErrAddressHasDelegatedTokens,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom:     "BND",
					MaxValidators: 2,
				}, nil)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, nil).Times(1)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: sdk.DefaultPowerReduction}, nil).Times(1)
				stakingKeeper.EXPECT().GetAllDelegatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{
					{
						ValidatorAddress: validatorValAddress,
						DelegatorAddress: validatorAccAddress,
						Shares:           sdk.DefaultPowerReduction.ToLegacyDec(),
					},
				}, nil)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
			},
		},
		{
			name:             "should fail - GetUnbondingDelegations returns error",
			validatorAddress: validatorAccAddress,
			expectedError:    errors.New("staking unbonding delegations error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom:     "BND",
					MaxValidators: 2,
				}, nil)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, nil)
				stakingKeeper.EXPECT().GetAllDelegatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegations(ctx, gomock.Any(), gomock.Any()).Return([]stakingtypes.UnbondingDelegation{}, errors.New("staking unbonding delegations error"))
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
			},
		},
		{
			name:             "should fail - unbonding delegations balances are greater than 0",
			validatorAddress: validatorAccAddress,
			expectedError:    types.ErrAddressHasUnbondingTokens,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom:     "BND",
					MaxValidators: 2,
				}, nil)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, nil)
				stakingKeeper.EXPECT().GetAllDelegatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegations(ctx, gomock.Any(), gomock.Any()).Return([]stakingtypes.UnbondingDelegation{
					{
						Entries: []stakingtypes.UnbondingDelegationEntry{
							{
								Balance: sdk.DefaultPowerReduction,
							},
						},
					},
				}, nil)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
			},
		},
		{
			name:             "should fail - bank keeper MintCoins returns error",
			validatorAddress: validatorAccAddress,
			expectedError:    errors.New("bank mint coins error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom:     "BND",
					MaxValidators: 2,
				}, nil)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, nil)
				stakingKeeper.EXPECT().GetAllDelegatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegations(ctx, gomock.Any(), gomock.Any()).Return([]stakingtypes.UnbondingDelegation{}, nil)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
				bankKeeper.EXPECT().MintCoins(ctx, gomock.Any(), gomock.Any()).Return(errors.New("bank mint coins error"))
			},
		},
		{
			name:             "should fail - bank keeper SendCoinsFromModuleToAccount returns error",
			validatorAddress: validatorAccAddress,
			expectedError:    errors.New("bank send coins from module to account error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom:     "BND",
					MaxValidators: 2,
				}, nil)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, nil)
				stakingKeeper.EXPECT().GetAllDelegatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegations(ctx, gomock.Any(), gomock.Any()).Return([]stakingtypes.UnbondingDelegation{}, nil)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
				bankKeeper.EXPECT().MintCoins(ctx, gomock.Any(), gomock.Any()).Return(nil)
				bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("bank send coins from module to account error"))
			},
		},
		{
			name:             "should pass - MsgAddValidator",
			validatorAddress: validatorAccAddress,
			pubKey:           msgPubKey,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom:     "BND",
					MaxValidators: 2,
				}, nil)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, nil)
				stakingKeeper.EXPECT().GetAllDelegatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegations(ctx, gomock.Any(), gomock.Any()).Return([]stakingtypes.UnbondingDelegation{}, nil)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
				bankKeeper.EXPECT().MintCoins(ctx, gomock.Any(), gomock.Any()).Return(nil)
				bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:             "should pass - validator not found when iterating over delegator delegations",
			validatorAddress: validatorAccAddress,
			pubKey:           msgPubKey,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom:     "BND",
					MaxValidators: 2,
				}, nil)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, nil).Times(1)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, stakingtypes.ErrNoValidatorFound).Times(1)
				stakingKeeper.EXPECT().GetAllDelegatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{
					{
						ValidatorAddress: validatorValAddress,
						DelegatorAddress: validatorAccAddress,
						Shares:           sdk.DefaultPowerReduction.ToLegacyDec(),
					},
				}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegations(ctx, gomock.Any(), gomock.Any()).Return([]stakingtypes.UnbondingDelegation{}, nil)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
				bankKeeper.EXPECT().MintCoins(ctx, gomock.Any(), gomock.Any()).Return(nil)
				bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			keeper, ctx := setupPoaKeeper(t, tc.stakingMocks, tc.bankMocks)

			msg := &types.MsgAddValidator{
				Authority:        keeper.GetAuthority(),
				ValidatorAddress: tc.validatorAddress,
				Description: stakingtypes.Description{
					Moniker:         "test",
					Identity:        "test",
					Website:         "test",
					SecurityContact: "test",
					Details:         "test",
				},
				Pubkey: tc.pubKey,
			}

			err := keeper.ExecuteAddValidator(ctx, msg)
			if tc.expectedError != nil {
				require.Contains(t, err.Error(), tc.expectedError.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestKeeper_ExecuteRemoveValidator(t *testing.T) {
	ctrl := gomock.NewController(t)
	selfDelegation := []stakingtypes.Delegation{{
		DelegatorAddress: validatorAccAddress,
		Shares:           math.LegacyOneDec(),
	}}

	tt := []struct {
		name             string
		validatorAddress string
		stakingMocks     func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper)
		bankMocks        func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper)
		expectedError    error
	}{
		{
			name:             "should fail - invalid validator address",
			validatorAddress: "invalidnaddress",
			expectedError:    errors.New("decoding bech32 failed"),
			stakingMocks:     func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
			bankMocks:        func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		},
		{
			name:             "should fail - staking keeper returns error on GetParams",
			validatorAddress: validatorAccAddress,
			expectedError:    errors.New("staking params error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{}, errors.New("staking params error"))
			},
			bankMocks: func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		},
		{
			name:             "should fail - staking keeper returns error on GetValidator",
			expectedError:    types.ErrAddressIsNotAValidator,
			validatorAddress: validatorAccAddress,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{}, errors.New("staking keeper get validator error"))
			},
			bankMocks: func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		},
		{
			name:             "should fail - staking keeper returns error on call GetUnbondingDelegationsFromValidator",
			validatorAddress: validatorAccAddress,
			expectedError:    errors.New("staking keeper get unbonding delegations from validator error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetValidatorDelegations(ctx, gomock.Any()).Return(selfDelegation, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return(
					[]stakingtypes.UnbondingDelegation{}, errors.New("staking keeper get unbonding delegations from validator error"))
				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(ctx, gomock.Any()).Return(errors.New("staking keeper hooks error"))
				stakingKeeper.EXPECT().Hooks().Return(hooks)
			},
			bankMocks: func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		},
		{
			name:             "should fail - staking keeper returns error on call SlashUnbondingDelegation",
			validatorAddress: validatorAccAddress,
			expectedError:    errors.New("staking keeper slash unbonding delegation error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetValidatorDelegations(ctx, gomock.Any()).Return(selfDelegation, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return(
					[]stakingtypes.UnbondingDelegation{
						{
							ValidatorAddress: validatorValAddress,
						},
					}, nil)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(ctx, gomock.Any()).Return(nil)
				stakingKeeper.EXPECT().Hooks().Return(hooks)

				stakingKeeper.EXPECT().SlashUnbondingDelegation(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(
					math.NewInt(0), errors.New("staking keeper slash unbonding delegation error"))
			},
			bankMocks: func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		},
		{
			name:             "should fail - staking keeper returns error on RemoveValidatorTokens call",
			validatorAddress: validatorAccAddress,
			expectedError:    errors.New("staking keeper remove validator tokens error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{
					Tokens: sdk.DefaultPowerReduction,
				}, nil)
				stakingKeeper.EXPECT().GetValidatorDelegations(ctx, gomock.Any()).Return(selfDelegation, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return(
					[]stakingtypes.UnbondingDelegation{}, nil,
				)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(ctx, gomock.Any()).Return(nil)
				hooks.EXPECT().BeforeValidatorSlashed(ctx, gomock.Any(), gomock.Any()).Return(errors.New("staking keeper hook error"))
				stakingKeeper.EXPECT().Hooks().Return(hooks).AnyTimes()

				stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, gomock.Any(), gomock.Any()).Return(
					stakingtypes.Validator{}, errors.New("staking keeper remove validator tokens error"),
				)
			},
			bankMocks: func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		},
		//nolint:dupl
		{
			name:             "should fail - bank keeper returns error on call BurnCoins for status bonded",
			validatorAddress: validatorAccAddress,
			expectedError:    errors.New("bank keeper burn coins error"),
			//nolint:dupl
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{
					Tokens: math.NewInt(0),
				}, nil)
				stakingKeeper.EXPECT().GetValidatorDelegations(ctx, gomock.Any()).Return(selfDelegation, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return(
					[]stakingtypes.UnbondingDelegation{}, nil,
				)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(ctx, gomock.Any()).Return(nil)
				stakingKeeper.EXPECT().Hooks().Return(hooks).AnyTimes()

				stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, gomock.Any(), gomock.Any()).Return(
					stakingtypes.Validator{
						Status: stakingtypes.Bonded,
					}, nil,
				)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().BurnCoins(ctx, gomock.Any(), gomock.Any()).Return(errors.New("bank keeper burn coins error"))
			},
		},
		//nolint:dupl
		{
			name:             "should fail - bank keeper returns error on call BurnCoins for status unbonding/unbonded",
			validatorAddress: validatorAccAddress,
			expectedError:    errors.New("bank keeper burn coins error"),
			//nolint:dupl
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{
					Tokens: math.NewInt(0),
				}, nil)
				stakingKeeper.EXPECT().GetValidatorDelegations(ctx, gomock.Any()).Return(selfDelegation, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return(
					[]stakingtypes.UnbondingDelegation{}, nil,
				)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(ctx, gomock.Any()).Return(nil)
				stakingKeeper.EXPECT().Hooks().Return(hooks).AnyTimes()

				stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, gomock.Any(), gomock.Any()).Return(
					stakingtypes.Validator{
						Status: stakingtypes.Unbonding,
					}, nil,
				)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().BurnCoins(ctx, gomock.Any(), gomock.Any()).Return(errors.New("bank keeper burn coins error"))
			},
		},
		{
			name:             "should fail - bank keeper returns error for invalid validator status",
			validatorAddress: validatorAccAddress,
			expectedError:    types.ErrInvalidValidatorStatus,
			//nolint:dupl
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{
					Tokens: math.NewInt(0),
				}, nil)
				stakingKeeper.EXPECT().GetValidatorDelegations(ctx, gomock.Any()).Return(selfDelegation, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return(
					[]stakingtypes.UnbondingDelegation{}, nil,
				)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(ctx, gomock.Any()).Return(nil)
				stakingKeeper.EXPECT().Hooks().Return(hooks).AnyTimes()

				stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, gomock.Any(), gomock.Any()).Return(
					stakingtypes.Validator{
						Status: stakingtypes.Unspecified,
					}, nil,
				)
			},
			bankMocks: func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		},
		{
			name:             "should fail - staking keeper returns error on call Unbond",
			validatorAddress: validatorAccAddress,
			expectedError:    errors.New("staking keeper unbond error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{
					Tokens: math.NewInt(0),
				}, nil)
				stakingKeeper.EXPECT().GetValidatorDelegations(ctx, gomock.Any()).Return(selfDelegation, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return(
					[]stakingtypes.UnbondingDelegation{}, nil,
				)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(ctx, gomock.Any()).Return(nil)
				stakingKeeper.EXPECT().Hooks().Return(hooks).AnyTimes()

				stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, gomock.Any(), gomock.Any()).Return(
					stakingtypes.Validator{
						Status: stakingtypes.Bonded,
					}, nil,
				)

				stakingKeeper.EXPECT().Unbond(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(
					math.NewInt(0), errors.New("staking keeper unbond error"),
				)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().BurnCoins(ctx, gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:             "should fail - validator without delegations was already removed",
			validatorAddress: validatorAccAddress,
			expectedError:    stakingtypes.ErrNoDelegatorForAddress,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{BondDenom: "BND"}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetValidatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{}, nil)
			},
			bankMocks: func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		},
		{
			name:             "should succeed - burns tokens and unbonds self plus foreign delegators",
			validatorAddress: validatorAccAddress,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				selfDelegatorAddr := sdk.MustAccAddressFromBech32(validatorAccAddress)
				foreignDelegatorAddr := sdk.AccAddress("foreign_delegator___")

				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{BondDenom: "BND"}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{
					Tokens:          math.NewInt(3),
					DelegatorShares: math.LegacyNewDec(3),
					Status:          stakingtypes.Bonded,
				}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return(
					[]stakingtypes.UnbondingDelegation{}, nil,
				)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(ctx, gomock.Any()).Return(nil)
				hooks.EXPECT().BeforeValidatorSlashed(ctx, gomock.Any(), math.LegacyOneDec()).Return(nil)
				stakingKeeper.EXPECT().Hooks().Return(hooks).AnyTimes()

				removeTokens := stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, gomock.Any(), gomock.Any()).Return(
					stakingtypes.Validator{Status: stakingtypes.Bonded}, nil,
				)
				stakingKeeper.EXPECT().GetValidatorDelegations(ctx, gomock.Any()).Return(
					[]stakingtypes.Delegation{
						{DelegatorAddress: selfDelegatorAddr.String(), Shares: math.LegacyOneDec()},
						{DelegatorAddress: foreignDelegatorAddr.String(), Shares: math.LegacyNewDec(2)},
					}, nil,
				)

				stakingKeeper.EXPECT().Unbond(ctx, selfDelegatorAddr, gomock.Any(), math.LegacyOneDec()).Return(math.NewInt(0), nil).After(removeTokens)
				stakingKeeper.EXPECT().Unbond(ctx, foreignDelegatorAddr, gomock.Any(), math.LegacyNewDec(2)).Return(math.NewInt(0), nil).After(removeTokens)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().BurnCoins(
					ctx,
					stakingtypes.BondedPoolName,
					sdk.NewCoins(sdk.NewCoin("BND", math.NewInt(3))),
				).Return(nil)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			keeper, ctx := setupPoaKeeper(t, tc.stakingMocks, tc.bankMocks)

			err := keeper.ExecuteRemoveValidator(ctx, tc.validatorAddress)
			if tc.expectedError != nil {
				require.Contains(t, err.Error(), tc.expectedError.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
