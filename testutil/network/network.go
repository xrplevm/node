package network

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Peersyst/exrp/app"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	tmcfg "github.com/cometbft/cometbft/config"
	tmflags "github.com/cometbft/cometbft/libs/cli/flags"
	"github.com/cometbft/cometbft/libs/log"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	"github.com/cometbft/cometbft/node"
	tmclient "github.com/cometbft/cometbft/rpc/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	pruningtypes "github.com/cosmos/cosmos-sdk/store/pruning/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/evmos/ethermint/crypto/hd"
	"github.com/evmos/ethermint/server/config"
	ethermint "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// network lock to only allow one test network at a time
var lock = new(sync.Mutex)

// AppConstructor defines a function which accepts a network configuration and
// creates an ABCI Application to provide to Tendermint.
type AppConstructor = func(val Validator) servertypes.Application

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
	AccountTokens       sdkmath.Int         // the amount of unique validator tokens (e.g. 1000node0)
	StakingTokens       sdkmath.Int         // the amount of tokens each validator has available to stake
	BondedTokens        sdkmath.Int         // the amount of tokens each validator stakes
	NumValidators       int                 // the total number of validators to create
	NumBondedValidators int                 // the total number of validators with bonded tokens to create
	ChainID             string              // the network chain-id
	TokenDenom          string              // the fees token denomination
	BondDenom           string              // the staking bond denomination
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

var (
	TokenDenom = "token"
	BondDenom  = "stake"
)

// DefaultConfig returns a sane default configuration suitable for nearly all
// testing requirements.
func DefaultConfig(numValidators int, numBondedValidators int) Config {
	encCfg := app.MakeEncodingConfig()
	cdc := encCfg.Codec
	genesisState := app.ModuleBasics.DefaultGenesis(cdc)

	return Config{
		Codec:               encCfg.Codec,
		TxConfig:            encCfg.TxConfig,
		LegacyAmino:         encCfg.Amino,
		InterfaceRegistry:   encCfg.InterfaceRegistry,
		AccountRetriever:    authtypes.AccountRetriever{},
		AppConstructor:      NewAppConstructor(encCfg),
		GenesisState:        genesisState,
		TimeoutCommit:       2 * time.Second,
		ChainID:             fmt.Sprintf("ethermint_%d-1", tmrand.Int63n(9999999999999)+1),
		NumValidators:       numValidators,
		NumBondedValidators: numBondedValidators,
		BondDenom:           BondDenom,
		TokenDenom:          TokenDenom,
		UnBoundingTime:      time.Second * 10,
		MinGasPrices:        fmt.Sprintf("0.000006%s", TokenDenom),
		AccountTokens:       sdkmath.NewInt(2000000000000000000),
		StakingTokens:       sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction),
		BondedTokens:        sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction),
		PruningStrategy:     pruningtypes.PruningOptionNothing,
		CleanupDir:          true,
		SigningAlgo:         string(hd.EthSecp256k1Type),
		KeyringOptions:      []keyring.Option{hd.EthSecp256k1Option()},
		PrintMnemonic:       false,
		EnableTMLogging:     true,
	}
}

type (
	Network struct {
		Logger     Logger
		BaseDir    string
		Validators []*Validator

		Config Config
	}

	Validator struct {
		AppConfig     *config.Config
		ClientCtx     client.Context
		Ctx           *server.Context
		Dir           string
		NodeID        string
		PubKey        cryptotypes.PubKey
		Moniker       string
		APIAddress    string
		RPCAddress    string
		P2PAddress    string
		Address       sdk.AccAddress
		ValAddress    sdk.ValAddress
		RPCClient     tmclient.Client
		JSONRPCClient *ethclient.Client

		tmNode      *node.Node
		api         *api.Server
		grpc        *grpc.Server
		grpcWeb     *http.Server
		jsonrpc     *http.Server
		jsonrpcDone chan struct{}
	}
)

type Logger interface {
	Log(args ...interface{})
	Logf(format string, args ...interface{})
}

var (
	_ Logger = (*testing.T)(nil)
	_ Logger = (*CLILogger)(nil)
)

type CLILogger struct {
	cmd *cobra.Command
}

func (s CLILogger) Log(args ...interface{}) {
	fmt.Println(args...)
	s.cmd.Println(args...)
}

func (s CLILogger) Logf(format string, args ...interface{}) {
	fmt.Println(format, args)
	s.cmd.Printf(format, args...)
}

func NewCLILogger(cmd *cobra.Command) CLILogger {
	return CLILogger{cmd}
}

// New creates a new Network for integration tests or in-process testnets run via the CLI
func New(l Logger, baseDir string, cfg Config) (*Network, error) {
	// only one caller/test can create and use a network at a time
	l.Log("acquiring test network lock")
	lock.Lock()

	if !ethermint.IsValidChainID(cfg.ChainID) {
		return nil, fmt.Errorf("invalid chain-id: %s", cfg.ChainID)
	}

	network := &Network{
		Logger:     l,
		BaseDir:    baseDir,
		Validators: make([]*Validator, cfg.NumValidators),
		Config:     cfg,
	}

	l.Logf("preparing test network with chain-id \"%s\"\n", cfg.ChainID)

	monikers := make([]string, cfg.NumValidators)
	nodeIDs := make([]string, cfg.NumValidators)
	valPubKeys := make([]cryptotypes.PubKey, cfg.NumValidators)

	var (
		genAccounts []authtypes.GenesisAccount
		genBalances []banktypes.Balance
		genFiles    []string
	)

	buf := bufio.NewReader(os.Stdin)

	// generate private keys, node IDs, and initial transactions
	for i := 0; i < cfg.NumValidators; i++ {
		appCfg := config.DefaultConfig()
		appCfg.Pruning = cfg.PruningStrategy
		appCfg.MinGasPrices = cfg.MinGasPrices
		appCfg.API.Enable = true
		appCfg.API.Swagger = false
		appCfg.Telemetry.Enabled = false
		appCfg.Telemetry.GlobalLabels = [][]string{{"chain_id", cfg.ChainID}}

		ctx := server.NewDefaultContext()
		tmCfg := ctx.Config
		tmCfg.Consensus.TimeoutCommit = cfg.TimeoutCommit

		// Only allow the first validator to expose an RPC, API and gRPC
		// server/client due to Tendermint in-process constraints.
		apiAddr := ""
		tmCfg.RPC.ListenAddress = ""
		appCfg.GRPC.Enable = false
		appCfg.GRPCWeb.Enable = false
		appCfg.JSONRPC.Enable = false
		apiListenAddr := ""

		if cfg.APIAddress != "" {
			apiListenAddr = cfg.APIAddress
		} else {
			var err error
			apiListenAddr, _, err = server.FreeTCPAddr()
			if err != nil {
				return nil, err
			}
		}

		appCfg.API.Address = apiListenAddr
		apiURL, err := url.Parse(apiListenAddr)
		if err != nil {
			return nil, err
		}
		apiAddr = fmt.Sprintf("http://%s:%s", apiURL.Hostname(), apiURL.Port())

		if cfg.RPCAddress != "" {
			tmCfg.RPC.ListenAddress = cfg.RPCAddress
		} else {
			rpcAddr, _, err := server.FreeTCPAddr()
			if err != nil {
				return nil, err
			}
			tmCfg.RPC.ListenAddress = rpcAddr
		}

		if cfg.GRPCAddress != "" {
			appCfg.GRPC.Address = cfg.GRPCAddress
		} else {
			_, grpcPort, err := server.FreeTCPAddr()
			if err != nil {
				return nil, err
			}
			appCfg.GRPC.Address = fmt.Sprintf("0.0.0.0:%s", grpcPort)
		}
		appCfg.GRPC.Enable = false

		_, grpcWebPort, err := server.FreeTCPAddr()
		if err != nil {
			return nil, err
		}
		appCfg.GRPCWeb.Address = fmt.Sprintf("0.0.0.0:%s", grpcWebPort)
		appCfg.GRPCWeb.Enable = false

		if cfg.JSONRPCAddress != "" {
			appCfg.JSONRPC.Address = cfg.JSONRPCAddress
		} else {
			_, jsonRPCPort, err := server.FreeTCPAddr()
			if err != nil {
				return nil, err
			}
			appCfg.JSONRPC.Address = fmt.Sprintf("127.0.0.1:%s", jsonRPCPort)
		}
		appCfg.JSONRPC.Enable = false
		appCfg.JSONRPC.API = config.GetAPINamespaces()

		logger := log.NewNopLogger()
		if cfg.EnableTMLogging && i == 0 {
			logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
			logger, _ = tmflags.ParseLogLevel("info", logger, tmcfg.DefaultLogLevel)
		}

		ctx.Logger = logger

		nodeDirName := fmt.Sprintf("node%d", i)
		nodeDir := filepath.Join(network.BaseDir, nodeDirName, "evmosd")
		clientDir := filepath.Join(network.BaseDir, nodeDirName, "evmoscli")
		gentxsDir := filepath.Join(network.BaseDir, "gentxs")

		err = os.MkdirAll(filepath.Join(nodeDir, "config"), 0o750)
		if err != nil {
			return nil, err
		}

		err = os.MkdirAll(clientDir, 0o750)
		if err != nil {
			return nil, err
		}

		tmCfg.SetRoot(nodeDir)
		tmCfg.Moniker = nodeDirName
		monikers[i] = nodeDirName

		proxyAddr, _, err := server.FreeTCPAddr()
		if err != nil {
			return nil, err
		}
		tmCfg.ProxyApp = proxyAddr

		p2pAddr, _, err := server.FreeTCPAddr()
		if err != nil {
			return nil, err
		}
		tmCfg.P2P.ListenAddress = p2pAddr
		tmCfg.P2P.AddrBookStrict = false
		tmCfg.P2P.AllowDuplicateIP = true

		nodeID, pubKey, err := genutil.InitializeNodeValidatorFiles(tmCfg)
		if err != nil {
			return nil, err
		}
		nodeIDs[i] = nodeID
		valPubKeys[i] = pubKey

		kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, clientDir, buf, cfg.Codec, cfg.KeyringOptions...)
		if err != nil {
			return nil, err
		}

		keyringAlgos, _ := kb.SupportedAlgorithms()
		algo, err := keyring.NewSigningAlgoFromString(cfg.SigningAlgo, keyringAlgos)
		if err != nil {
			return nil, err
		}

		addr, secret, err := testutil.GenerateSaveCoinKey(kb, nodeDirName, "", true, algo)
		if err != nil {
			return nil, err
		}

		// if PrintMnemonic is set to true, we print the first validator node's secret to the network's logger
		// for debugging and manual testing
		if cfg.PrintMnemonic && i == 0 {
			printMnemonic(l, secret)
		}

		info := map[string]string{"secret": secret}
		infoBz, err := json.Marshal(info)
		if err != nil {
			return nil, err
		}

		// save private key seed words
		err = WriteFile(fmt.Sprintf("%v.json", "key_seed"), clientDir, infoBz)
		if err != nil {
			return nil, err
		}

		balances := sdk.NewCoins(
			sdk.NewCoin(TokenDenom, cfg.AccountTokens),
			sdk.NewCoin(cfg.BondDenom, cfg.StakingTokens),
		)

		genFiles = append(genFiles, tmCfg.GenesisFile())
		genBalances = append(genBalances, banktypes.Balance{Address: addr.String(), Coins: balances.Sort()})
		genAccounts = append(genAccounts, &ethermint.EthAccount{
			BaseAccount: authtypes.NewBaseAccount(addr, nil, 0, 0),
			CodeHash:    common.BytesToHash(evmtypes.EmptyCodeHash).Hex(),
		})

		commission, err := sdk.NewDecFromStr("0.5")
		if err != nil {
			return nil, err
		}

		if i < cfg.NumBondedValidators {

			createValMsg, err := stakingtypes.NewMsgCreateValidator(
				sdk.ValAddress(addr),
				valPubKeys[i],
				sdk.NewCoin(cfg.BondDenom, cfg.BondedTokens),
				stakingtypes.NewDescription(nodeDirName, "", "", "", ""),
				stakingtypes.NewCommissionRates(commission, sdk.OneDec(), sdk.OneDec()),
				sdk.OneInt(),
			)
			if err != nil {
				return nil, err
			}

			p2pURL, err := url.Parse(p2pAddr)
			if err != nil {
				return nil, err
			}

			memo := fmt.Sprintf("%s@%s:%s", nodeIDs[i], p2pURL.Hostname(), p2pURL.Port())
			fee := sdk.NewCoins(sdk.NewCoin(cfg.BondDenom, sdkmath.NewInt(0)))
			txBuilder := cfg.TxConfig.NewTxBuilder()
			err = txBuilder.SetMsgs(createValMsg)
			if err != nil {
				return nil, err
			}
			txBuilder.SetFeeAmount(fee)    // Arbitrary fee
			txBuilder.SetGasLimit(1000000) // Need at least 100386
			txBuilder.SetMemo(memo)

			txFactory := tx.Factory{}
			txFactory = txFactory.
				WithChainID(cfg.ChainID).
				WithMemo(memo).
				WithKeybase(kb).
				WithTxConfig(cfg.TxConfig)

			if err := tx.Sign(txFactory, nodeDirName, txBuilder, true); err != nil {
				return nil, err
			}

			txBz, err := cfg.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
			if err != nil {
				return nil, err
			}

			if err := WriteFile(fmt.Sprintf("%v.json", nodeDirName), gentxsDir, txBz); err != nil {
				return nil, err
			}
		}

		customAppTemplate, _ := config.AppConfig(TokenDenom)
		srvconfig.SetConfigTemplate(customAppTemplate)
		srvconfig.WriteConfigFile(filepath.Join(nodeDir, "config/app.toml"), appCfg)

		ctx.Viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
		ctx.Viper.SetConfigFile(filepath.Join(nodeDir, "config/app.toml"))
		err = ctx.Viper.ReadInConfig()
		if err != nil {
			return nil, err
		}

		clientCtx := client.Context{}.
			WithKeyringDir(clientDir).
			WithKeyring(kb).
			WithHomeDir(tmCfg.RootDir).
			WithChainID(cfg.ChainID).
			WithInterfaceRegistry(cfg.InterfaceRegistry).
			WithCodec(cfg.Codec).
			WithLegacyAmino(cfg.LegacyAmino).
			WithTxConfig(cfg.TxConfig).
			WithAccountRetriever(cfg.AccountRetriever)

		network.Validators[i] = &Validator{
			AppConfig:  appCfg,
			ClientCtx:  clientCtx,
			Ctx:        ctx,
			Dir:        filepath.Join(network.BaseDir, nodeDirName),
			NodeID:     nodeID,
			PubKey:     pubKey,
			Moniker:    nodeDirName,
			RPCAddress: tmCfg.RPC.ListenAddress,
			P2PAddress: tmCfg.P2P.ListenAddress,
			APIAddress: apiAddr,
			Address:    addr,
			ValAddress: sdk.ValAddress(addr),
		}
	}

	err := initGenFiles(cfg, genAccounts, genBalances, genFiles)
	if err != nil {
		return nil, err
	}
	err = collectGenFiles(cfg, network.Validators, network.BaseDir)
	if err != nil {
		return nil, err
	}

	l.Log("starting test network...")
	for _, v := range network.Validators {
		err := startInProcess(cfg, v)
		if err != nil {
			return nil, err
		}
	}

	l.Log("started test network")

	// Ensure we cleanup incase any test was abruptly halted (e.g. SIGINT) as any
	// defer in a test would not be called.
	server.TrapSignal(network.Cleanup)

	return network, nil
}

// LatestHeight returns the latest height of the network or an error if the
// query fails or no validators exist.
func (n *Network) LatestHeight() (int64, error) {
	if len(n.Validators) == 0 {
		return 0, errors.New("no validators available")
	}

	status, err := n.Validators[0].RPCClient.Status(context.Background())
	if err != nil {
		return 0, err
	}

	return status.SyncInfo.LatestBlockHeight, nil
}

// WaitForHeight performs a blocking check where it waits for a block to be
// committed after a given block. If that height is not reached within a timeout,
// an error is returned. Regardless, the latest height queried is returned.
func (n *Network) WaitForHeight(h int64) (int64, error) {
	return n.WaitForHeightWithTimeout(h, 10*time.Second)
}

// WaitForHeightWithTimeout is the same as WaitForHeight except the caller can
// provide a custom timeout.
func (n *Network) WaitForHeightWithTimeout(h int64, t time.Duration) (int64, error) {
	ticker := time.NewTicker(time.Second)
	timeout := time.After(t)

	if len(n.Validators) == 0 {
		return 0, errors.New("no validators available")
	}

	var latestHeight int64
	val := n.Validators[0]

	for {
		select {
		case <-timeout:
			ticker.Stop()
			return latestHeight, errors.New("timeout exceeded waiting for block")
		case <-ticker.C:
			status, err := val.RPCClient.Status(context.Background())
			if err == nil && status != nil {
				latestHeight = status.SyncInfo.LatestBlockHeight
				if latestHeight >= h {
					return latestHeight, nil
				}
			}
		}
	}
}

// WaitForNextBlock waits for the next block to be committed, returning an error
// upon failure.
func (n *Network) WaitForNextBlock() error {
	lastBlock, err := n.LatestHeight()
	if err != nil {
		return err
	}

	_, err = n.WaitForHeight(lastBlock + 1)
	if err != nil {
		return err
	}

	return err
}

// Cleanup removes the root testing (temporary) directory and stops both the
// Tendermint and API services. It allows other callers to create and start
// test networks. This method must be called when a test is finished, typically
// in a defer.
func (n *Network) Cleanup() {
	defer func() {
		lock.Unlock()
		n.Logger.Log("released test network lock")
	}()

	n.Logger.Log("cleaning up test network...")

	for _, v := range n.Validators {
		if v.tmNode != nil && v.tmNode.IsRunning() {
			_ = v.tmNode.Stop()
		}

		if v.api != nil {
			_ = v.api.Close()
		}

		if v.grpc != nil {
			v.grpc.Stop()
			if v.grpcWeb != nil {
				_ = v.grpcWeb.Close()
			}
		}

		if v.jsonrpc != nil {
			shutdownCtx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancelFn()

			if err := v.jsonrpc.Shutdown(shutdownCtx); err != nil {
				v.tmNode.Logger.Error("HTTP server shutdown produced a warning", "error", err.Error())
			} else {
				v.tmNode.Logger.Info("HTTP server shut down, waiting 5 sec")
				select {
				case <-time.Tick(5 * time.Second):
				case <-v.jsonrpcDone:
				}
			}
		}
	}

	if n.Config.CleanupDir {
		_ = os.RemoveAll(n.BaseDir)
	}

	n.Logger.Log("finished cleaning up test network")
}
