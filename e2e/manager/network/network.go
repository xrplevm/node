package network

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	tmcfg "github.com/cometbft/cometbft/config"
	tmflags "github.com/cometbft/cometbft/libs/cli/flags"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/ethereum/go-ethereum/common"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/evmos/evmos/v19/server/config"
	evmostypes "github.com/evmos/evmos/v19/types"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
)

// Network defines a local in-process testing network using SimApp. It can be
// configured to start any number of validators, each with its own RPC and API
// clients. Typically, this test network would be used in client and integration
// testing where user input is expected.
//
// Note, due to Tendermint constraints in regards to RPC functionality, there
// may only be one test network running at a time. Thus, any caller must be
// sure to Cleanup after testing is finished in order to allow other tests
// to create networks. In addition, only the first validator will have a valid
// RPC and API server/client.
type Network struct {
	Logger     Logger
	BaseDir    string
	Validators []*Validator

	Config Config
}

func NewNetwork(l Logger, baseDir string, cfg Config) (*Network, error) {
	lock.Lock()

	if !evmostypes.IsValidChainID(cfg.ChainID) {
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
			appCfg.JSONRPC.Address = fmt.Sprintf("0.0.0.0:%s", jsonRPCPort)
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
			sdk.NewCoin(cfg.TokenDenom, cfg.AccountTokens),
			sdk.NewCoin(cfg.BondDenom, cfg.StakingTokens),
		)

		genFiles = append(genFiles, tmCfg.GenesisFile())
		genBalances = append(genBalances, banktypes.Balance{Address: addr.String(), Coins: balances.Sort()})
		genAccounts = append(genAccounts, &evmostypes.EthAccount{
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
			fee := sdk.NewCoins(sdk.NewCoin(cfg.BondDenom, sdk.NewInt(0)))
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

		customAppTemplate, _ := config.AppConfig(cfg.TokenDenom)
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

func (n *Network) Cleanup() {
	defer func() {
		time.Sleep(10 * time.Second)
		lock.Unlock()
		n.Logger.Log("released test network lock")
	}()

	n.Logger.Log("cleaning up test network...")

	for _, v := range n.Validators {
		if v.TmNode != nil && v.TmNode.IsRunning() {
			_ = v.TmNode.Stop()
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
				v.TmNode.Logger.Error("HTTP server shutdown produced a warning", "error", err.Error())
			} else {
				v.TmNode.Logger.Info("HTTP server shut down, waiting 5 sec")
				select {
				case <-time.NewTicker(5 * time.Second).C:
				case <-v.jsonrpcDone:
				}
			}
		}
	}

	if n.Config.CleanupDir {
		time.Sleep(10 * time.Second)
		_ = os.RemoveAll(n.BaseDir)
	}

	n.Logger.Log("finished cleaning up test network")
}
