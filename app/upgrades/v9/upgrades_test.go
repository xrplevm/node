package v9

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	legacyevmtypes "github.com/xrplevm/node/v10/types/legacy/ethermint/evm"
	legacytypes "github.com/xrplevm/node/v10/types/legacy/ethermint/types"
	"google.golang.org/grpc"
)

var (
	_ ERC20Keeper = (*mockERC20Keeper)(nil)
	_ EvmKeeper   = (*mockEvmKeeper)(nil)
)

type mockERC20Keeper struct {
	dynamicPrecompiles []common.Address
	nativePrecompiles  []common.Address
}

func (m *mockERC20Keeper) SetDynamicPrecompile(_ sdk.Context, precompile common.Address) {
	m.dynamicPrecompiles = append(m.dynamicPrecompiles, precompile)
}

func (m *mockERC20Keeper) SetNativePrecompile(_ sdk.Context, precompile common.Address) {
	m.nativePrecompiles = append(m.nativePrecompiles, precompile)
}

type mockEvmKeeper struct {
	params       evmtypes.Params
	setParams    bool
	codeHashes   map[common.Address]common.Hash
	setParamsErr error
}

func (m *mockEvmKeeper) GetParams(sdk.Context) evmtypes.Params { return m.params }
func (m *mockEvmKeeper) SetParams(_ sdk.Context, params evmtypes.Params) error {
	m.setParams = true
	m.params = params
	return m.setParamsErr
}

func (m *mockEvmKeeper) SetCodeHash(_ sdk.Context, addrBytes, hashBytes []byte) {
	if m.codeHashes == nil {
		m.codeHashes = make(map[common.Address]common.Hash)
	}
	m.codeHashes[common.BytesToAddress(addrBytes)] = common.BytesToHash(hashBytes)
}

func TestCreateUpgradeHandlerMigratesLegacyState(t *testing.T) {
	const dynamicPrecompile = "0x0000000000000000000000000000000000000001"

	appCodec := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	configurator := module.NewConfigurator(appCodec, grpc.NewServer(), grpc.NewServer())
	moduleManager := module.NewManager()
	versions := module.VersionMap{"bank": 1}

	ctx, storeKeys := testContext(t, authtypes.StoreKey, erc20types.StoreKey, evmtypes.StoreKey)
	setLegacyEvmParams(ctx, storeKeys, appCodec, legacyevmtypes.Params{
		EvmDenom: "axrp",
	})
	ctx.KVStore(storeKeys[erc20types.StoreKey]).Set([]byte("DynamicPrecompiles"), []byte(dynamicPrecompile))

	accountKeeper := testAccountKeeper(storeKeys[authtypes.StoreKey])
	evmKeeper := &mockEvmKeeper{}
	erc20Keeper := &mockERC20Keeper{}

	updatedVersions, err := CreateUpgradeHandler(
		moduleManager,
		configurator,
		storeKeys,
		appCodec,
		accountKeeper,
		evmKeeper,
		erc20Keeper,
	)(
		ctx,
		upgradetypes.Plan{Name: UpgradeName},
		versions,
	)

	require.NoError(t, err)
	require.Empty(t, updatedVersions)
	require.Equal(t, []common.Address{common.HexToAddress(dynamicPrecompile)}, erc20Keeper.dynamicPrecompiles)
	require.True(t, evmKeeper.setParams)
}

func TestMigrateEvmModuleMigratesLegacyParams(t *testing.T) {
	const (
		createAllowedAddress  = "0x0000000000000000000000000000000000000001"
		callRestrictedAddress = "0x0000000000000000000000000000000000000002"
		staticPrecompile      = "0x0000000000000000000000000000000000000003"
	)
	ctx, storeKeys := testContext(t, evmtypes.StoreKey)
	appCodec := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	legacyParams := legacyevmtypes.Params{
		EvmDenom:    "axrp",
		ExtraEIPs:   []string{"ethereum_3855", "ethereum_3860"},
		EVMChannels: []string{"channel-0"},
		AccessControl: legacyevmtypes.AccessControl{
			Create: legacyevmtypes.AccessControlType{
				AccessType:        legacyevmtypes.AccessTypePermissioned,
				AccessControlList: []string{createAllowedAddress},
			},
			Call: legacyevmtypes.AccessControlType{
				AccessType:        legacyevmtypes.AccessTypeRestricted,
				AccessControlList: []string{callRestrictedAddress},
			},
		},
		ActiveStaticPrecompiles: []string{staticPrecompile},
	}
	expectedParams := evmtypes.Params{
		EvmDenom:    "axrp",
		ExtraEIPs:   []int64{3855, 3860},
		EVMChannels: []string{"channel-0"},
		AccessControl: evmtypes.AccessControl{
			Create: evmtypes.AccessControlType{
				AccessType:        evmtypes.AccessTypePermissioned,
				AccessControlList: []string{createAllowedAddress},
			},
			Call: evmtypes.AccessControlType{
				AccessType:        evmtypes.AccessTypeRestricted,
				AccessControlList: []string{callRestrictedAddress},
			},
		},
		ActiveStaticPrecompiles: []string{staticPrecompile},
	}
	setLegacyEvmParams(ctx, storeKeys, appCodec, legacyParams)
	keeper := &mockEvmKeeper{}

	err := MigrateEvmModule(ctx, storeKeys, appCodec, keeper)

	require.NoError(t, err)
	require.True(t, keeper.setParams)
	require.Equal(t, expectedParams, keeper.params)
}

func TestMigrateErc20ModuleMigratesAndDeletesLegacyPrecompiles(t *testing.T) {
	const (
		dynamicPrecompile1 = "0x0000000000000000000000000000000000000001"
		dynamicPrecompile2 = "0x0000000000000000000000000000000000000002"
		nativePrecompile   = "0x0000000000000000000000000000000000000003"
	)
	dynamicPrecompiles := []byte(dynamicPrecompile1 + dynamicPrecompile2)
	nativePrecompiles := []byte(nativePrecompile)

	ctx, storeKeys := testContext(t, erc20types.StoreKey)
	store := ctx.KVStore(storeKeys[erc20types.StoreKey])

	store.Set([]byte("DynamicPrecompiles"), dynamicPrecompiles)
	store.Set([]byte("NativePrecompiles"), nativePrecompiles)

	keeper := &mockERC20Keeper{}

	MigrateErc20Module(ctx, storeKeys, keeper)

	require.Equal(t, []common.Address{
		common.HexToAddress(dynamicPrecompile1),
		common.HexToAddress(dynamicPrecompile2),
	}, keeper.dynamicPrecompiles)

	require.Equal(t, []common.Address{common.HexToAddress(nativePrecompile)}, keeper.nativePrecompiles)
	require.Nil(t, store.Get([]byte("DynamicPrecompiles")))
	require.Nil(t, store.Get([]byte("NativePrecompiles")))
}

func TestMigrateEthAccountsToBaseAccountsConvertsContracts(t *testing.T) {
	ctx, storeKeys := testContext(t, authtypes.StoreKey)
	accountKeeper := testAccountKeeper(storeKeys[authtypes.StoreKey])
	contractAddress := sdk.AccAddress(common.HexToAddress("0x0000000000000000000000000000000000000007").Bytes())
	codeHash := common.HexToHash("0x1234")
	ethAccount := &legacytypes.EthAccount{
		BaseAccount: authtypes.NewBaseAccount(contractAddress, nil, 0, 0),
		CodeHash:    codeHash.Hex(),
	}
	accountKeeper.SetAccount(ctx, ethAccount)
	keeper := &mockEvmKeeper{}

	MigrateEthAccountsToBaseAccounts(ctx, accountKeeper, keeper)

	require.IsType(t, &authtypes.BaseAccount{}, accountKeeper.GetAccount(ctx, contractAddress))
	require.Equal(t, codeHash, keeper.codeHashes[common.BytesToAddress(contractAddress)])
}

func testAccountKeeper(key *storetypes.KVStoreKey) authkeeper.AccountKeeper {
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{})
	encCfg.InterfaceRegistry.RegisterImplementations((*sdk.AccountI)(nil), &legacytypes.EthAccount{})

	return authkeeper.NewAccountKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(key),
		legacytypes.ProtoAccount,
		nil,
		authcodec.NewBech32Codec("cosmos"),
		"cosmos",
		authtypes.NewModuleAddress("gov").String(),
	)
}

func setLegacyEvmParams(
	ctx sdk.Context,
	storeKeys map[string]*storetypes.KVStoreKey,
	appCodec codec.Codec,
	params legacyevmtypes.Params,
) {
	ctx.KVStore(storeKeys[evmtypes.StoreKey]).Set(evmtypes.KeyPrefixParams, appCodec.MustMarshal(&params))
}

func testContext(t *testing.T, storeKeys ...string) (sdk.Context, map[string]*storetypes.KVStoreKey) {
	t.Helper()

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	keys := make(map[string]*storetypes.KVStoreKey, len(storeKeys))
	for _, storeKey := range storeKeys {
		key := storetypes.NewKVStoreKey(storeKey)
		keys[storeKey] = key
		ms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	}
	require.NoError(t, ms.LoadLatestVersion())

	return sdk.NewContext(ms, cmtproto.Header{}, false, log.NewNopLogger()), keys
}
