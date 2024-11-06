package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"

	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/golang/mock/gomock"
	"github.com/xrplevm/node/v3/x/poa/testutil"
	"github.com/xrplevm/node/v3/x/poa/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	accountAddressPrefix = "ethm"
	bip44CoinType        = 60
)

func setupSdkConfig() {
	accountPubKeyPrefix := accountAddressPrefix + "pub"
	validatorAddressPrefix := accountAddressPrefix + "valoper"
	validatorPubKeyPrefix := accountAddressPrefix + "valoperpub"
	consNodeAddressPrefix := accountAddressPrefix + "valcons"
	consNodePubKeyPrefix := accountAddressPrefix + "valconspub"

	// Set config
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(accountAddressPrefix, accountPubKeyPrefix)
	config.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)
	config.SetCoinType(bip44CoinType)
	config.SetPurpose(sdk.Purpose) // Shared
}

func setupPoAKeeper(t *testing.T) (
	*Keeper,
	sdk.Context,
	*testutil.MockPubKey,
) {
	setupSdkConfig()

	key := sdk.NewKVStoreKey(types.StoreKey)

	tsKey := sdk.NewTransientStoreKey("transient_test")
	testCtx := sdktestutil.DefaultContextWithDB(t, key, tsKey)
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: time.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	types.RegisterInterfaces(encCfg.InterfaceRegistry)
	stakingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)

	msr := baseapp.NewMsgServiceRouter()
	msr.SetInterfaceRegistry(encCfg.InterfaceRegistry)

	// gomock initializations
	ctrl := gomock.NewController(t)
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	stakingKeeper := testutil.NewMockStakingKeeper(ctrl)
	pubKey := testutil.NewMockPubKey(ctrl)
	stakingMsr := testutil.NewMockStakingMsgServer(ctrl)
	// bank keeper expectations
	bankKeeper.EXPECT().GetBalance(ctx, gomock.Any(), gomock.Any()).Return(sdk.Coin{
		Amount: math.NewInt(0),
	}).AnyTimes()
	bankKeeper.EXPECT().MintCoins(ctx, gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	bankKeeper.EXPECT().BurnCoins(ctx, stakingtypes.BondedPoolName, gomock.Any()).Return(nil).AnyTimes()
	bankKeeper.EXPECT().SendCoinsFromAccountToModule(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// staking keeper expectations
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

	stakingMsr.EXPECT().CreateValidator(gomock.Any(), gomock.Any()).Return(&stakingtypes.MsgCreateValidatorResponse{}, nil).AnyTimes()

	poaKeeper := NewKeeper(
		encCfg.Codec,
		paramtypes.NewSubspace(encCfg.Codec, encCfg.Amino, key, tsKey, "poa"),
		msr,
		bankKeeper,
		stakingKeeper,
		"ethm1wunfhl05vc8r8xxnnp8gt62wa54r6y52pg03zq",
	)
	poaKeeper.SetParams(ctx, types.DefaultParams())
	types.RegisterMsgServer(msr, NewMsgServerImpl(*poaKeeper))
	stakingtypes.RegisterMsgServer(msr, stakingMsr)

	return poaKeeper, ctx, pubKey
}
