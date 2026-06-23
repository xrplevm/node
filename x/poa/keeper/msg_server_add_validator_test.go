package keeper

import (
	"errors"
	"testing"

	"cosmossdk.io/math"
	types1 "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v10/x/poa/testutil"
	"github.com/xrplevm/node/v10/x/poa/types"
	"go.uber.org/mock/gomock"
)

func TestMsgServer_AddValidator(t *testing.T) {
	ctrl := gomock.NewController(t)
	pubKey := testutil.NewMockPubKey(ctrl)
	msgPubKey, _ := types1.NewAnyWithValue(pubKey)

	tt := []struct {
		name             string
		authority        string
		validatorAddress string
		stakingMocks     func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper)
		bankMocks        func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper)
		expectedErr      error
	}{
		{
			name:             "should fail - invalid authority address",
			authority:        "invalidauthority",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			stakingMocks:     func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
			bankMocks:        func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
			expectedErr:      govtypes.ErrInvalidSigner,
		},
		{
			name:             "should fail - invalid validator address",
			authority:        poaAuthority,
			validatorAddress: "invalidvalidatoraddress",
			stakingMocks:     func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
			bankMocks:        func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
			expectedErr:      errors.New("decoding bech32 failed"),
		},
		{
			name:             "should pass",
			authority:        poaAuthority,
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
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
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			poaKeeper, ctx := setupPoaKeeper(t, tc.stakingMocks, tc.bankMocks)
			msgServer := NewMsgServerImpl(*poaKeeper)

			msg := &types.MsgAddValidator{
				Authority:        tc.authority,
				ValidatorAddress: tc.validatorAddress,
				Description: stakingtypes.Description{
					Moniker:         "test",
					Identity:        "test",
					Website:         "test",
					SecurityContact: "test",
					Details:         "test",
				},
				Pubkey: msgPubKey,
			}

			_, err := msgServer.AddValidator(ctx, msg)
			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
