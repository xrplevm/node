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
	"github.com/xrplevm/node/v5/x/poa/testutil"
	"github.com/xrplevm/node/v5/x/poa/types"
)

func poaKeeperTestSetup(t *testing.T) (*Keeper, sdk.Context) {
	stakingExpectations := func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
		stakingHooks := testutil.NewMockStakingHooks(gomock.NewController(t))
		stakingHooks.EXPECT().BeforeValidatorModified(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		stakingHooks.EXPECT().BeforeValidatorSlashed(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
			BondDenom: "XRP",
		}, nil).AnyTimes()
		stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, nil).AnyTimes()
		stakingKeeper.EXPECT().GetAllDelegatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{}, nil).AnyTimes()
		stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return([]stakingtypes.UnbondingDelegation{}, nil).AnyTimes()
		stakingKeeper.EXPECT().SlashUnbondingDelegation(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(math.ZeroInt(), nil).AnyTimes()
		stakingKeeper.EXPECT().RemoveDelegation(ctx, gomock.Any()).Return(nil).AnyTimes()
		stakingKeeper.EXPECT().RemoveValidatorTokensAndShares(ctx, gomock.Any(), gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0), Status: stakingtypes.Bonded}, math.ZeroInt(), nil).AnyTimes()
		stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, gomock.Any(), gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0), Status: stakingtypes.Bonded}, nil).AnyTimes()
		stakingKeeper.EXPECT().BondDenom(ctx).Return("XRP", nil).AnyTimes()
		stakingKeeper.EXPECT().Unbond(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(math.ZeroInt(), nil).AnyTimes()
		stakingKeeper.EXPECT().Hooks().Return(stakingHooks).AnyTimes()
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

	slashingExpectations := func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {}

	return setupPoaKeeper(t, stakingExpectations, bankExpectations, slashingExpectations)
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
		slashingMocks    func(ctx sdk.Context, slashingKeeper *testutil.MockSlashingKeeper)
		expectedError    error
	}{
		{
			name:             "should fail - invalid validator address",
			validatorAddress: "invalidnaddress",
			expectedError:    errors.New("decoding bech32 failed"),
			stakingMocks:     func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
			bankMocks:        func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
			slashingMocks:    func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
		},
		{
			name:             "should fail - staking keeper returns error on GetParams",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    errors.New("staking params error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{}, errors.New("staking params error"))
			},
			bankMocks:     func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
			slashingMocks: func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
		},
		{
			name:             "should fail - validator has bonded tokens",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    types.ErrAddressHasBankTokens,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: sdk.DefaultPowerReduction,
				})
			},
			slashingMocks: func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
		},
		{
			name:             "should fail - staking keeper returns error on GetValidator",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    errors.New("staking validator error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{}, errors.New("staking validator error"))
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
			},
			slashingMocks: func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
		},
		{
			name:             "should fail - staking keeper returns validator with tokens",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    types.ErrAddressHasBondedTokens,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: sdk.DefaultPowerReduction}, nil)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
			},
			slashingMocks: func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
		},
		{
			name:             "should fail - staking keeper returns error on GetAllDelegatorDelegations",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    errors.New("staking delegations error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, nil)
				stakingKeeper.EXPECT().GetAllDelegatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{}, errors.New("staking delegations error"))
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
			},
			slashingMocks: func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
		},
		{
			name:             "should fail - delegations are greater than 0 with invalid delegation validator address",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    errors.New("decoding bech32 failed"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
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
			slashingMocks: func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
		},
		// {
		// 	name: "should fail - delegations are greater than 0 with error on GetValidator call",
		// 	validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
		// 	expectedError: errors.New("staking validator error"),
		// 	stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
		// 		valAddr, _ := sdk.ValAddressFromBech32("ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp")
		// 		delAddr, _ := sdk.ValAddressFromBech32("ethm1dakgyqjulg29m5fmv992g2y66m9g2mjn6hahwg")

		// 		stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
		// 			BondDenom: "BND",
		// 		}, nil)
		// 		stakingKeeper.EXPECT().GetValidator(ctx, valAddr).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, nil)
		// 		stakingKeeper.EXPECT().GetValidator(ctx, delAddr).Return(stakingtypes.Validator{}, errors.New("staking validator error"))
		// 		stakingKeeper.EXPECT().GetAllDelegatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{
		// 			{
		// 				ValidatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
		// 				Shares: sdk.DefaultPowerReduction.ToLegacyDec(),
		// 			},
		// 		}, nil)
		// 	},
		// 	bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
		// 		bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
		// 			Denom:  "BND",
		// 			Amount: math.NewInt(0),
		// 		})
		// 	},
		// 	slashingMocks: func(ctx sdk.Context, slashingKeeper *testutil.MockSlashingKeeper) {},
		// },
		{
			name:             "should fail - GetUnbondingDelegationsFromValidator returns error",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    errors.New("staking unbonding delegations error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
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
			slashingMocks: func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
		},
		{
			name:             "should fail - unbonding delegations balances are greater than 0",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    types.ErrAddressHasUnbondingTokens,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
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
			slashingMocks: func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
		},
		{
			name:             "should fail - bank keeper MintCoins returns error",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    errors.New("bank mint coins error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
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
			slashingMocks: func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
		},
		{
			name:             "should fail - bank keeper SendCoinsFromModuleToAccount returns error",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    errors.New("bank send coins from module to account error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
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
			slashingMocks: func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
		},
		{
			name:             "should pass - MsgAddValidator",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			pubKey:           msgPubKey,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
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
			slashingMocks: func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			keeper, ctx := setupPoaKeeper(t, tc.stakingMocks, tc.bankMocks, tc.slashingMocks)

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
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    errors.New("staking params error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{}, errors.New("staking params error"))
			},
			bankMocks: func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		},
		{
			name:             "should fail - bank keeper returns error on SendCoinsFromAccountToModule when validator has balance",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    errors.New("bank send coins from account to module error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: sdk.DefaultPowerReduction,
				})
				bankKeeper.EXPECT().SendCoinsFromAccountToModule(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("bank send coins from account to module error"))
			},
		},
		{
			name:             "should fail - bank keeper returns error on BurnCoins when validator has balance",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    errors.New("bank burn coins error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: sdk.DefaultPowerReduction,
				})
				bankKeeper.EXPECT().SendCoinsFromAccountToModule(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				bankKeeper.EXPECT().BurnCoins(ctx, gomock.Any(), gomock.Any()).Return(errors.New("bank burn coins error"))
			},
		},
		{
			name:             "should fail - staking keeper returns error on GetValidator with balance",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{}, errors.New("staking keeper get validator error"))
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: sdk.DefaultPowerReduction,
				})
				bankKeeper.EXPECT().SendCoinsFromAccountToModule(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				bankKeeper.EXPECT().BurnCoins(ctx, gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:             "should fail - staking keeper returns error on GetValidator with no balance",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    types.ErrAddressHasNoTokens,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{}, errors.New("staking keeper get validator error"))
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
			},
		},
		{
			name:             "should fail - staking keeper returns error on call GetUnbondingDelegationsFromValidator",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
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
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
			},
		},
		{
			name:             "should fail - staking keeper returns error on call SlashUnbondingDelegation",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    errors.New("staking keeper slash unbonding delegation error"),
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return(
					[]stakingtypes.UnbondingDelegation{
						{
							ValidatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
						},
					}, nil)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(ctx, gomock.Any()).Return(nil)
				stakingKeeper.EXPECT().Hooks().Return(hooks)

				stakingKeeper.EXPECT().SlashUnbondingDelegation(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(
					math.NewInt(0), errors.New("staking keeper slash unbonding delegation error"))
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
			},
		},
		{
			name:             "should fail - staking keeper returns error on RemoveValidatorTokens call",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
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

				stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, gomock.Any(), gomock.Any()).Return(
					stakingtypes.Validator{}, errors.New("staking keeper remove validator tokens error"),
				)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
			},
		},
		//nolint:dupl
		{
			name:             "should fail - bank keeper returns error on call BurnCoins for status bonded",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    errors.New("bank keeper burn coins error"),
			//nolint:dupl
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

				stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, gomock.Any(), gomock.Any()).Return(
					stakingtypes.Validator{
						Status: stakingtypes.Bonded,
					}, nil,
				)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
				bankKeeper.EXPECT().BurnCoins(ctx, gomock.Any(), gomock.Any()).Return(errors.New("bank keeper burn coins error"))
			},
		},
		//nolint:dupl
		{
			name:             "should fail - bank keeper returns error on call BurnCoins for status unbonding/unbonded",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    errors.New("bank keeper burn coins error"),
			//nolint:dupl
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

				stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, gomock.Any(), gomock.Any()).Return(
					stakingtypes.Validator{
						Status: stakingtypes.Unbonding,
					}, nil,
				)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
				bankKeeper.EXPECT().BurnCoins(ctx, gomock.Any(), gomock.Any()).Return(errors.New("bank keeper burn coins error"))
			},
		},
		{
			name:             "should fail - bank keeper returns error for invalid validator status",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedError:    types.ErrInvalidValidatorStatus,
			//nolint:dupl
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

				stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, gomock.Any(), gomock.Any()).Return(
					stakingtypes.Validator{
						Status: stakingtypes.Unspecified,
					}, nil,
				)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
			},
		},
		{
			name:             "should fail - staking keeper returns error on call Unbond",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
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
				bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
					Denom:  "BND",
					Amount: math.NewInt(0),
				})
				bankKeeper.EXPECT().BurnCoins(ctx, gomock.Any(), gomock.Any()).Return(nil)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			keeper, ctx := setupPoaKeeper(t, tc.stakingMocks, tc.bankMocks, func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {})

			err := keeper.ExecuteRemoveValidator(ctx, tc.validatorAddress)
			if tc.expectedError != nil {
				require.Contains(t, err.Error(), tc.expectedError.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
