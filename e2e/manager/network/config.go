package network

import (
	"fmt"
	"os"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/simapp"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	pruningtypes "github.com/cosmos/cosmos-sdk/store/pruning/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/evmos/evmos/v19/crypto/hd"
	"github.com/xrplevm/node/v3/app"
)

// Config defines the necessary configuration used to bootstrap and start an
// in-process local testing network.
type Config struct {
	KeyringOptions      []keyring.Option // keyring configuration options
	Codec               codec.Codec
	LegacyAmino         *codec.LegacyAmino // TODO: Remove!
	InterfaceRegistry   codectypes.InterfaceRegistry
	TxConfig            client.TxConfig
	AccountRetriever    client.AccountRetriever
	AppConstructor      AppConstructor      // the ABCI application constructor
	GenesisState        simapp.GenesisState // custom gensis state to provide
	TimeoutCommit       time.Duration       // the consensus commitment timeout
	AccountTokens       math.Int            // the amount of unique validator tokens (e.g. 1000node0)
	StakingTokens       math.Int            // the amount of tokens each validator has available to stake
	BondedTokens        math.Int            // the amount of tokens each validator stakes
	SignedBlocksWindow  int64               // slashing signed blocks window
	NumValidators       int                 // the total number of validators to create and bond
	NumBondedValidators int                 // the total number of validators with bonded tokens to create
	ChainID             string              // the network chain-id
	BondDenom           string              // the staking bond denomination
	TokenDenom          string              // the fees token denomination
	UnBoundingTime      time.Duration       // the time to unbound and accept governance proposals
	MinGasPrices        string              // the minimum gas prices each validator will accept
	PruningStrategy     string              // the pruning strategy each validator will have
	SigningAlgo         string              // signing algorithm for keys
	RPCAddress          string              // RPC listen address (including port)
	JSONRPCAddress      string              // JSON-RPC listen address (including port)
	APIAddress          string              // REST API listen address (including port)
	GRPCAddress         string              // GRPC server listen address (including port)
	EnableTMLogging     bool                // enable Tendermint logging to STDOUT
	CleanupDir          bool                // remove base temporary directory during cleanup
	PrintMnemonic       bool                // print the mnemonic of first validator as log output for testing
}

// DefaultConfig returns a sane default configuration suitable for nearly all
// testing requirements.
func DefaultConfig(numValidators int, numBondedValidators int, blockTime time.Duration, unbondingBlocks int64) Config {
	encCfg := app.MakeEncodingConfig()
	chainID := fmt.Sprintf("exrp_%d-1", tmrand.Int63n(9999999999999)+1)
	return Config{
		Codec:               encCfg.Codec,
		TxConfig:            encCfg.TxConfig,
		LegacyAmino:         encCfg.Amino,
		InterfaceRegistry:   encCfg.InterfaceRegistry,
		AccountRetriever:    authtypes.AccountRetriever{},
		AppConstructor:      NewAppConstructor(encCfg, chainID),
		GenesisState:        app.ModuleBasics.DefaultGenesis(encCfg.Codec),
		TimeoutCommit:       blockTime,
		ChainID:             chainID,
		NumValidators:       numValidators,
		NumBondedValidators: numBondedValidators,
		BondDenom:           "apoa",
		MinGasPrices:        fmt.Sprintf("0.000006%s", "apoa"),
		AccountTokens:       sdk.TokensFromConsensusPower(1000000000000000000, sdk.DefaultPowerReduction),
		StakingTokens:       sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction),
		BondedTokens:        sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction),
		TokenDenom:          "axrp",
		UnBoundingTime:      (time.Duration(unbondingBlocks) * blockTime) + time.Second,
		PruningStrategy:     pruningtypes.PruningOptionNothing,
		CleanupDir:          os.Getenv("TEST_CLEANUP_DIR") != "false",
		SigningAlgo:         string(hd.EthSecp256k1Type),
		KeyringOptions:      []keyring.Option{hd.EthSecp256k1Option()},
		PrintMnemonic:       false,
		EnableTMLogging:     true,
		SignedBlocksWindow:  2,
	}
}