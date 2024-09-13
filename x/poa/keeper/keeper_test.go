package keeper

import (
	"testing"

	"cosmossdk.io/math"
	types1 "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v3/x/poa/testutil"
	"github.com/xrplevm/node/v3/x/poa/types"
)

func poaKeeperTestSetup(t *testing.T) (*Keeper, sdk.Context) {
	stakingExpectations := func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
		stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
			BondDenom: "XRP",
		}).AnyTimes()
		stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0)}, true).AnyTimes()
		stakingKeeper.EXPECT().GetAllDelegatorDelegations(ctx, gomock.Any()).Return([]stakingtypes.Delegation{}).AnyTimes()
		stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return([]stakingtypes.UnbondingDelegation{}).AnyTimes()
		stakingKeeper.EXPECT().SlashUnbondingDelegation(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(sdk.ZeroInt()).AnyTimes()
		stakingKeeper.EXPECT().RemoveDelegation(ctx, gomock.Any()).Return(nil).AnyTimes()
		stakingKeeper.EXPECT().RemoveValidatorTokensAndShares(ctx, gomock.Any(), gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0), Status: stakingtypes.Bonded}, sdk.ZeroInt()).AnyTimes()
		stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, gomock.Any(), gomock.Any()).Return(stakingtypes.Validator{Tokens: math.NewInt(0), Status: stakingtypes.Bonded}).AnyTimes()
		stakingKeeper.EXPECT().BondDenom(ctx).Return("XRP").AnyTimes()
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
func TestPoAKeeper_ExecuteAddValidator(t *testing.T) {
	keeper, ctx := poaKeeperTestSetup(t)
	ctrl := gomock.NewController(t)
	pubKey := testutil.NewMockPubKey(ctrl)
	msgPubKey, _ := types1.NewAnyWithValue(pubKey)

	msg := &types.MsgAddValidator{
		Authority:        keeper.GetAuthority(),
		ValidatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
		Description: stakingtypes.Description{
			Moniker:         "test",
			Identity:        "test",
			Website:         "test",
			SecurityContact: "test",
			Details:         "test",
		},
		Pubkey: msgPubKey,
	}

	err := keeper.ExecuteAddValidator(ctx, msg)
	require.NoError(t, err)
}

func TestPoAKeeper_ExecuteRemoveValidator(t *testing.T) {
	keeper, ctx := poaKeeperTestSetup(t)

	err := keeper.ExecuteRemoveValidator(ctx, "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp")
	require.NoError(t, err)
}
