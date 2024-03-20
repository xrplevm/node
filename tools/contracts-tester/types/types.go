package types

import (
	"context"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type ContractTestSuite interface {
	SetupEnv(ctx context.Context) error
	RunTests()
}

type TestAccount struct {
	EvmAddress common.Address
	PrivateKey string
	CallOpts   *bind.CallOpts
}

type TestAccountSigner struct {
	TestAccount
	Signer bind.SignerFn
}

const (
	EnvRunTests       = "BRIDGE_RUN_TESTS"
	EnvNodeUrl        = "NODE_URL"
	EnvClaimerAccount = "CLAIMER_ACCOUNT"
	EnvWitnesses      = "TEST_WITNESSES"
	EnvSafeThreshold  = "SAFE_THRESHOLD"
	EnvBridgeValues   = "BRIDGE"
	EnvBridgeKey      = "BRIDGE_KEY"
)

const (
	DefaultRunTests       = "false"
	DefaultNodeUrl        = "http://localhost:8545"
	DefaultClaimerAccount = "fCD6E97a730C114a6D8fd4684c0CaC15b2F81088:54170bebb7d2dae201dacdcb4b7e506a3d57b07f858877c52dbfdae342ab8a4a"
	DefaultWitnesses      = "C255f8ED0Fb0af012c9Dc33D5712FA8f672eE85f:29b10d36d005d4c540da657081bba9f737e195126f14821f227cb1f5dbd669c8"
	DefaultSafeThreshold  = "1"
)

func getEnv(envVar string) string {
	return strings.TrimSpace(os.Getenv(envVar))
}

func getEnvWithDefault(envVar, defaultValue string) string {
	envValue := getEnv(envVar)
	if envValue != "" {
		return envValue
	}

	return defaultValue
}

func GetRunTests() bool {
	value := getEnvWithDefault(EnvRunTests, DefaultRunTests)
	return value == "true"
}

func GetNodeUrl() string {
	return getEnvWithDefault(EnvNodeUrl, DefaultNodeUrl)
}

func GetClaimerAccount(ctx context.Context) TestAccount {
	envClaimerAccount := getEnv(EnvClaimerAccount)
	claimerInfo := strings.Split(envClaimerAccount, ":")
	if len(claimerInfo) != 2 {
		claimerInfo = strings.Split(DefaultClaimerAccount, ":")
	}

	claimerAddress := common.HexToAddress(claimerInfo[0])
	callOpts := bind.CallOpts{Pending: true, From: claimerAddress, Context: ctx}

	return TestAccount{claimerAddress, claimerInfo[1], &callOpts}
}

func GetWitnesses(ctx context.Context) []TestAccount {
	witnessesStr := getEnvWithDefault(EnvWitnesses, DefaultWitnesses)
	witnessesStrs := strings.Split(witnessesStr, ",")

	witnesses := []TestAccount{}
	for _, witnessStr := range witnessesStrs {
		witnessInfo := strings.Split(witnessStr, ":")
		evmAddress := common.HexToAddress(witnessInfo[0])
		privateKey := ""
		if len(witnessInfo) > 1 {
			privateKey = witnessInfo[1]
		}

		callOpts := bind.CallOpts{Pending: true, From: evmAddress, Context: ctx}
		witnesses = append(witnesses, TestAccount{evmAddress, privateKey, &callOpts})
	}

	return witnesses
}

func AddSignerToAccount(account TestAccount, chainId *big.Int) (TestAccountSigner, error) {
	ecdsaPrivateKey, err := crypto.HexToECDSA(account.PrivateKey)
	if err != nil {
		return TestAccountSigner{}, err
	}
	signer := types.LatestSignerForChainID(chainId)

	testAccountSigner := TestAccountSigner{
		account,
		func(address common.Address, transaction *types.Transaction) (*types.Transaction, error) {
			return types.SignTx(transaction, signer, ecdsaPrivateKey)
		},
	}

	return testAccountSigner, nil
}

func GetSafeThreshold() int {
	thresholdStr := getEnv(EnvSafeThreshold)
	threshold, err := strconv.Atoi(thresholdStr)
	if err == nil {
		return threshold
	}

	threshold, _ = strconv.Atoi(DefaultSafeThreshold)
	return threshold
}

func GetBridgeValues() (*common.Address, *big.Int, *big.Int) {
	bridgeStr := getEnv(EnvBridgeValues)
	if bridgeStr == "" {
		return nil, nil, nil
	}

	bridgeValues := strings.Split(bridgeStr, ",")
	if len(bridgeValues) != 3 {
		return nil, nil, nil
	}

	address := common.HexToAddress(bridgeValues[0])

	minCreateAmount, _ := big.NewInt(0).SetString(bridgeValues[1], 10)
	sigReward, _ := big.NewInt(0).SetString(bridgeValues[2], 10)

	return &address, minCreateAmount, sigReward
}

func GetBridgeKey() string {
	return getEnv(EnvBridgeKey)
}
