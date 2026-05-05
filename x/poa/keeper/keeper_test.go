package keeper

import (
	"errors"
	"testing"

	"cosmossdk.io/math"
	types1 "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v10/x/poa/testutil"
	"github.com/xrplevm/node/v10/x/poa/types"
)

func poaKeeperTestSetup(t *testing.T) (*Keeper, sdk.Context) {
	stakingExpectations := func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
		stakingHooks := testutil.NewMockStakingHooks(gomock.NewController(t))
		stakingHooks.EXPECT().BeforeValidatorModified(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		stakingHooks.EXPECT().BeforeValidatorSlashed(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
			BondDenom:     "XRP",
			MaxValidators: 32,
		}, nil).AnyTimes()
		stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, nil).AnyTimes()
		stakingKeeper.EXPECT().GetAllDelegatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{}, nil).AnyTimes()
		stakingKeeper.EXPECT().GetValidatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{}, nil).AnyTimes()
		stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return([]stakingtypes.UnbondingDelegation{}, nil).AnyTimes()
		stakingKeeper.EXPECT().SlashUnbondingDelegation(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(math.ZeroInt(), nil).AnyTimes()
		stakingKeeper.EXPECT().RemoveDelegation(ctx, gomock.Any()).Return(nil).AnyTimes()
		stakingKeeper.EXPECT().RemoveValidatorTokensAndShares(ctx, gomock.Any(), gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0), Status: stakingtypes.Bonded}, math.ZeroInt(), nil).AnyTimes()
		stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, gomock.Any(), gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0), Status: stakingtypes.Bonded}, nil).AnyTimes()
		stakingKeeper.EXPECT().BondDenom(ctx).Return("XRP", nil).AnyTimes()
		stakingKeeper.EXPECT().Unbond(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(math.ZeroInt(), nil).AnyTimes()
		stakingKeeper.EXPECT().Hooks().Return(stakingHooks).AnyTimes()
		stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{}, nil).AnyTimes()
	}

	bankExpectations := func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
		bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
			Amount: math.NewInt(0),
		}).AnyTimes()
		bankKeeper.EXPECT().MintCoins(ctx, gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		bankKeeper.EXPECT().BurnCoins(ctx, stakingtypes.BondedPoolName, gomock.Any()).Return(nil).AnyTimes()
		bankKeeper.EXPECT().SendCoinsFromAccountToModule(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	}

	return setupPoaKeeper(t, stakingExpectations, bankExpectations)
}

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
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    errors.New("staking params error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{}, errors.New("staking params error"))
			},
			bankMocks: func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		},
		{
			name:             "should fail - maximum validators reached",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
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
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
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
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
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
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
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
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
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
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
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
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
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
						ValidatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
						DelegatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
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
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
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
						ValidatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
						DelegatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
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
			name:             "should fail - GetUnbondingDelegationsFromValidator returns error",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    errors.New("staking unbonding delegations error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom:     "BND",
					MaxValidators: 2,
				}, nil)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, nil)
				stakingKeeper.EXPECT().GetAllDelegatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return([]stakingtypes.UnbondingDelegation{}, errors.New("staking unbonding delegations error"))
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
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    types.ErrAddressHasUnbondingTokens,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom:     "BND",
					MaxValidators: 2,
				}, nil)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, nil)
				stakingKeeper.EXPECT().GetAllDelegatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return([]stakingtypes.UnbondingDelegation{
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
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    errors.New("bank mint coins error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom:     "BND",
					MaxValidators: 2,
				}, nil)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, nil)
				stakingKeeper.EXPECT().GetAllDelegatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return([]stakingtypes.UnbondingDelegation{}, nil)
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
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    errors.New("bank send coins from module to account error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom:     "BND",
					MaxValidators: 2,
				}, nil)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, nil)
				stakingKeeper.EXPECT().GetAllDelegatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return([]stakingtypes.UnbondingDelegation{}, nil)
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
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			pubKey:           msgPubKey,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom:     "BND",
					MaxValidators: 2,
				}, nil)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, nil)
				stakingKeeper.EXPECT().GetAllDelegatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return([]stakingtypes.UnbondingDelegation{}, nil)
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
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
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
						ValidatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
						DelegatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
						Shares:           sdk.DefaultPowerReduction.ToLegacyDec(),
					},
				}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return([]stakingtypes.UnbondingDelegation{}, nil)
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

func mustAccAddress(t testing.TB, address string) sdk.AccAddress {
	t.Helper()

	accAddr, err := sdk.AccAddressFromBech32(address)
	require.NoError(t, err)

	return accAddr
}

func TestKeeper_ExecuteRemoveValidator(t *testing.T) {
	ctrl := gomock.NewController(t)

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
			validatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
			expectedError:    errors.New("staking params error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{}, errors.New("staking params error"))
			},
			bankMocks: func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		},
		{
			name:             "should fail - staking keeper returns error on GetValidator",
			expectedError:    types.ErrAddressIsNotAValidator,
			validatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
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
			validatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
			expectedError:    errors.New("staking keeper get unbonding delegations from validator error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{}, nil)
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
			validatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
			expectedError:    errors.New("staking keeper slash unbonding delegation error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return(
					[]stakingtypes.UnbondingDelegation{
						{
							ValidatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
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
			validatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
			expectedError:    errors.New("staking keeper remove validator tokens error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{
					Tokens: sdk.DefaultPowerReduction,
				}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return(
					[]stakingtypes.UnbondingDelegation{}, nil,
				)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(ctx, gomock.Any()).Return(nil)
				hooks.EXPECT().BeforeValidatorSlashed(ctx, gomock.Any(), gomock.Any()).Return(errors.New("staking keeper hook error"))
				stakingKeeper.EXPECT().Hooks().Return(hooks).AnyTimes()

				stakingKeeper.EXPECT().GetValidatorDelegations(ctx, gomock.Any()).Return(
					[]stakingtypes.Delegation{}, nil,
				)

				stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, gomock.Any(), gomock.Any()).Return(
					stakingtypes.Validator{}, errors.New("staking keeper remove validator tokens error"),
				)
			},
			bankMocks: func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		},
		{
			name:             "should fail - staking keeper returns error on GetValidatorDelegations",
			validatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
			expectedError:    errors.New("staking keeper get validator delegations error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{
					Tokens: math.NewInt(0),
				}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return(
					[]stakingtypes.UnbondingDelegation{}, nil,
				)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(ctx, gomock.Any()).Return(nil)
				stakingKeeper.EXPECT().Hooks().Return(hooks).AnyTimes()

				stakingKeeper.EXPECT().GetValidatorDelegations(ctx, gomock.Any()).Return(
					nil, errors.New("staking keeper get validator delegations error"),
				)
			},
			bankMocks: func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		},
		//nolint:dupl
		{
			name:             "should fail - bank keeper returns error on call BurnCoins for status bonded",
			validatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
			expectedError:    errors.New("bank keeper burn coins error"),
			//nolint:dupl
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				validatorTokens := math.NewInt(10)
				validator := stakingtypes.Validator{
					OperatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
					Tokens:          validatorTokens,
					DelegatorShares: math.LegacyNewDec(10),
					Status:          stakingtypes.Bonded,
				}

				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(validator, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return(
					[]stakingtypes.UnbondingDelegation{}, nil,
				)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(ctx, gomock.Any()).Return(nil)
				hooks.EXPECT().BeforeValidatorSlashed(ctx, gomock.Any(), math.LegacyOneDec()).Return(nil)
				stakingKeeper.EXPECT().Hooks().Return(hooks).AnyTimes()

				stakingKeeper.EXPECT().GetValidatorDelegations(ctx, gomock.Any()).Return(
					[]stakingtypes.Delegation{}, nil,
				)

				stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, validator, validatorTokens).Return(
					stakingtypes.Validator{
						OperatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
						DelegatorShares: validator.DelegatorShares,
						Status:          stakingtypes.Bonded,
					}, nil,
				)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				validatorTokens := math.NewInt(10)
				bankKeeper.EXPECT().BurnCoins(
					ctx,
					stakingtypes.BondedPoolName,
					sdk.NewCoins(sdk.NewCoin("BND", validatorTokens)),
				).Return(errors.New("bank keeper burn coins error"))
			},
		},
		//nolint:dupl
		{
			name:             "should fail - bank keeper returns error on call BurnCoins for status unbonding/unbonded",
			validatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
			expectedError:    errors.New("bank keeper burn coins error"),
			//nolint:dupl
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				validatorTokens := math.NewInt(10)
				validator := stakingtypes.Validator{
					OperatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
					Tokens:          validatorTokens,
					DelegatorShares: math.LegacyNewDec(10),
					Status:          stakingtypes.Unbonding,
				}

				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(validator, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return(
					[]stakingtypes.UnbondingDelegation{}, nil,
				)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(ctx, gomock.Any()).Return(nil)
				hooks.EXPECT().BeforeValidatorSlashed(ctx, gomock.Any(), math.LegacyOneDec()).Return(nil)
				stakingKeeper.EXPECT().Hooks().Return(hooks).AnyTimes()

				stakingKeeper.EXPECT().GetValidatorDelegations(ctx, gomock.Any()).Return(
					[]stakingtypes.Delegation{}, nil,
				)

				stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, validator, validatorTokens).Return(
					stakingtypes.Validator{
						OperatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
						DelegatorShares: validator.DelegatorShares,
						Status:          stakingtypes.Unbonding,
					}, nil,
				)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				validatorTokens := math.NewInt(10)
				bankKeeper.EXPECT().BurnCoins(
					ctx,
					stakingtypes.NotBondedPoolName,
					sdk.NewCoins(sdk.NewCoin("BND", validatorTokens)),
				).Return(errors.New("bank keeper burn coins error"))
			},
		},
		{
			name:             "should fail - bank keeper returns error for invalid validator status",
			validatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
			expectedError:    types.ErrInvalidValidatorStatus,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{
					Tokens: math.NewInt(0),
				}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return(
					[]stakingtypes.UnbondingDelegation{}, nil,
				)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(ctx, gomock.Any()).Return(nil)
				stakingKeeper.EXPECT().Hooks().Return(hooks).AnyTimes()

				stakingKeeper.EXPECT().GetValidatorDelegations(ctx, gomock.Any()).Return(
					[]stakingtypes.Delegation{}, nil,
				)

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
			validatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
			expectedError:    errors.New("staking keeper unbond error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{
					Tokens: math.NewInt(0),
				}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return(
					[]stakingtypes.UnbondingDelegation{}, nil,
				)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(ctx, gomock.Any()).Return(nil)
				stakingKeeper.EXPECT().Hooks().Return(hooks).AnyTimes()

				stakingKeeper.EXPECT().GetValidatorDelegations(ctx, gomock.Any()).Return(
					[]stakingtypes.Delegation{{
						DelegatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
						ValidatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
						Shares:           math.LegacyOneDec(),
					}}, nil,
				)

				stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, gomock.Any(), gomock.Any()).Return(
					stakingtypes.Validator{
						Status: stakingtypes.Bonded,
					}, nil,
				)

				stakingKeeper.EXPECT().Unbond(
					ctx,
					mustAccAddress(t, "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp"),
					gomock.Any(),
					math.LegacyOneDec(),
				).Return(math.NewInt(0), errors.New("staking keeper unbond error"))
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().BurnCoins(ctx, gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:             "should fail - invalid delegator bech32 in delegation list",
			validatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
			expectedError:    errors.New("decoding bech32 failed"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{
					Tokens: math.NewInt(0),
				}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return(
					[]stakingtypes.UnbondingDelegation{}, nil,
				)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(ctx, gomock.Any()).Return(nil)
				stakingKeeper.EXPECT().Hooks().Return(hooks).AnyTimes()

				stakingKeeper.EXPECT().GetValidatorDelegations(ctx, gomock.Any()).Return(
					[]stakingtypes.Delegation{{
						DelegatorAddress: "not-a-valid-bech32",
						ValidatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
						Shares:           math.LegacyOneDec(),
					}}, nil,
				)
			},
			bankMocks: func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		},
		{
			name:             "should succeed - no delegations skips unbond",
			validatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
			expectedError:    nil,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				validatorTokens := math.ZeroInt()
				validator := stakingtypes.Validator{
					OperatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
					Tokens:          validatorTokens,
				}

				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(validator, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return(
					[]stakingtypes.UnbondingDelegation{}, nil,
				)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(ctx, gomock.Any()).Return(nil)
				stakingKeeper.EXPECT().Hooks().Return(hooks).AnyTimes()

				stakingKeeper.EXPECT().GetValidatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{}, nil)

				stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, validator, validatorTokens).Return(
					stakingtypes.Validator{
						OperatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
						Status:          stakingtypes.Bonded,
					}, nil,
				)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().BurnCoins(
					ctx,
					stakingtypes.BondedPoolName,
					sdk.NewCoins(sdk.NewCoin("BND", math.ZeroInt())),
				).Return(nil)
			},
		},
		{
			name:             "should succeed - burns tokens and unbonds only self delegation",
			validatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
			expectedError:    nil,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				selfDelegatorAddr := mustAccAddress(t, "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp")
				validatorTokens := math.NewInt(1)
				validator := stakingtypes.Validator{
					OperatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
					Tokens:          validatorTokens,
					DelegatorShares: math.LegacyOneDec(),
					Status:          stakingtypes.Bonded,
				}

				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(validator, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return(
					[]stakingtypes.UnbondingDelegation{}, nil,
				)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(ctx, gomock.Any()).Return(nil)
				hooks.EXPECT().BeforeValidatorSlashed(ctx, gomock.Any(), math.LegacyOneDec()).Return(nil)
				stakingKeeper.EXPECT().Hooks().Return(hooks).AnyTimes()

				stakingKeeper.EXPECT().GetValidatorDelegations(ctx, gomock.Any()).Return(
					[]stakingtypes.Delegation{
						{
							DelegatorAddress: selfDelegatorAddr.String(),
							ValidatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
							Shares:           math.LegacyOneDec(),
						},
					}, nil,
				)

				stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, validator, validatorTokens).Return(
					stakingtypes.Validator{
						OperatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
						DelegatorShares: validator.DelegatorShares,
						Status:          stakingtypes.Bonded,
					}, nil,
				)

				stakingKeeper.EXPECT().Unbond(ctx, selfDelegatorAddr, gomock.Any(), math.LegacyOneDec()).Return(math.NewInt(0), nil)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				validatorTokens := math.NewInt(1)
				bankKeeper.EXPECT().BurnCoins(
					ctx,
					stakingtypes.BondedPoolName,
					sdk.NewCoins(sdk.NewCoin("BND", validatorTokens)),
				).Return(nil)
			},
		},
		{
			name:             "should succeed - burns tokens and unbonds self plus foreign delegators",
			validatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
			expectedError:    nil,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				selfDelegatorAddr := mustAccAddress(t, "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp")
				foreignDelegatorAddr := mustAccAddress(t, "ethm1wunfhl05vc8r8xxnnp8gt62wa54r6y52pg03zq")
				validatorTokens := math.NewInt(3)
				validator := stakingtypes.Validator{
					OperatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
					Tokens:          validatorTokens,
					DelegatorShares: math.LegacyNewDec(3),
					Status:          stakingtypes.Bonded,
				}

				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(validator, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return(
					[]stakingtypes.UnbondingDelegation{}, nil,
				)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(ctx, gomock.Any()).Return(nil)
				hooks.EXPECT().BeforeValidatorSlashed(ctx, gomock.Any(), math.LegacyOneDec()).Return(nil)
				stakingKeeper.EXPECT().Hooks().Return(hooks).AnyTimes()

				stakingKeeper.EXPECT().GetValidatorDelegations(ctx, gomock.Any()).Return(
					[]stakingtypes.Delegation{
						{
							DelegatorAddress: selfDelegatorAddr.String(),
							ValidatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
							Shares:           math.LegacyOneDec(),
						},
						{
							DelegatorAddress: foreignDelegatorAddr.String(),
							ValidatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
							Shares:           math.LegacyNewDec(2),
						},
					}, nil,
				)

				stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, validator, validatorTokens).Return(
					stakingtypes.Validator{
						OperatorAddress: "ethmvaloper1a0pd5cyew47pvgf7rd7axxy3humv9ev0urudmu",
						DelegatorShares: validator.DelegatorShares,
						Status:          stakingtypes.Bonded,
					}, nil,
				)

				stakingKeeper.EXPECT().Unbond(ctx, selfDelegatorAddr, gomock.Any(), math.LegacyOneDec()).Return(math.NewInt(0), nil)
				stakingKeeper.EXPECT().Unbond(ctx, foreignDelegatorAddr, gomock.Any(), math.LegacyNewDec(2)).Return(math.NewInt(0), nil)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				validatorTokens := math.NewInt(3)
				bankKeeper.EXPECT().BurnCoins(
					ctx,
					stakingtypes.BondedPoolName,
					sdk.NewCoins(sdk.NewCoin("BND", validatorTokens)),
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
