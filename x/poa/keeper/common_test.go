package keeper

import (
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"

	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/golang/mock/gomock"
	"github.com/xrplevm/node/v8/x/poa/testutil"
	"github.com/xrplevm/node/v8/x/poa/types"

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

func getStakingKeeperMock(t *testing.T, ctx sdk.Context, setExpectations func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper)) *testutil.MockStakingKeeper {
	ctrl := gomock.NewController(t)
	stakingKeeper := testutil.NewMockStakingKeeper(ctrl)
	setExpectations(ctx, stakingKeeper)
	return stakingKeeper
}

func getBankKeeperMock(t *testing.T, ctx sdk.Context, setExpectations func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper)) *testutil.MockBankKeeper {
	ctrl := gomock.NewController(t)
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	setExpectations(ctx, bankKeeper)
	return bankKeeper
}

func getCtxMock(t *testing.T, key *storetypes.KVStoreKey, tsKey *storetypes.TransientStoreKey) sdk.Context {
	setupSdkConfig()

	testCtx := sdktestutil.DefaultContextWithDB(t, key, tsKey)
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: time.Now()})
	return ctx
}

func getMockedPoAKeeper(t *testing.T, key *storetypes.KVStoreKey, tsKey *storetypes.TransientStoreKey, ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper, bankKeeper *testutil.MockBankKeeper) *Keeper {
	encCfg := moduletestutil.MakeTestEncodingConfig()

	types.RegisterInterfaces(encCfg.InterfaceRegistry)
	stakingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)

	msr := baseapp.NewMsgServiceRouter()
	msr.SetInterfaceRegistry(encCfg.InterfaceRegistry)

	ctrl := gomock.NewController(t)
	stakingMsr := testutil.NewMockStakingMsgServer(ctrl)

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

	return poaKeeper
}

func setupPoaKeeper(t *testing.T, setStakingExpectations func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper), setBankExpectations func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper)) (*Keeper, sdk.Context) {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	tsKey := storetypes.NewTransientStoreKey("test")

	ctx := getCtxMock(t, key, tsKey)
	stakingKeeper := getStakingKeeperMock(t, ctx, setStakingExpectations)
	bankKeeper := getBankKeeperMock(t, ctx, setBankExpectations)

	return getMockedPoAKeeper(t, key, tsKey, ctx, stakingKeeper, bankKeeper), ctx
}
