package bridge

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/Peersyst/exrp/cmd/exrpd/cmd"
	envTypes "github.com/Peersyst/exrp/tools/contracts-tester/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const executionRevertedStr = "desc = execution reverted"

var zeroAddress common.Address = common.HexToAddress("0x0000000000000000000000000000000000000000")
var valueMultiplier = big.NewInt(1000000000000000000)

type BridgeTestSuite struct {
	t            *testing.T
	client       *ethclient.Client
	bridge       *Bridge
	chainId      *big.Int
	initBlock    uint64
	threshold    int
	claimer      envTypes.TestAccountSigner
	witnesses    []envTypes.TestAccountSigner
	safeAddress  common.Address
	bridgeTests  bool
	bridgeConfig *XChainTypesBridgeConfig
	bridgeParams *XChainTypesBridgeParams
	bridgeKey    string
	claims       int
	ctx          context.Context
}

func CreateBridgeSuite(t *testing.T) envTypes.ContractTestSuite {
	return &BridgeTestSuite{t: t}
}

func (suite *BridgeTestSuite) SetupEnv(ctx context.Context) error {
	// Client
	client, err := ethclient.Dial(envTypes.GetNodeUrl())
	if err != nil {
		return fmt.Errorf("BRIDGE_TEST: Error creating client")
	}
	suite.client = client

	// Contract
	bridge, err := NewBridge(common.HexToAddress(cmd.BridgeProxyModuleAddress), client)
	if err != nil {
		return fmt.Errorf("Error instantiating bridge contract: '%+v'", err)
	}
	suite.bridge = bridge

	// Chain Id
	chainId, err := client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("Error getting chain id: '%+v'", err)
	}
	suite.chainId = chainId

	// Init block
	block, err := client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("Error getting chain id: '%+v'", err)
	}
	suite.initBlock = block

	// Claimer account
	claimer := envTypes.GetClaimerAccount(ctx)
	suite.claimer, err = envTypes.AddSignerToAccount(claimer, suite.chainId)
	if err != nil {
		return fmt.Errorf("Error adding signer to account: '%+v'", err)
	}

	// Witnesses accounts
	witnesses := envTypes.GetWitnesses(ctx)
	suite.witnesses = []envTypes.TestAccountSigner{}
	for _, witness := range witnesses {
		witnessSigner, err := envTypes.AddSignerToAccount(witness, suite.chainId)
		if err != nil {
			return fmt.Errorf("Error adding signer to account: '%+v'", err)
		}

		suite.witnesses = append(suite.witnesses, witnessSigner)
	}

	// Constant values
	suite.safeAddress = common.HexToAddress(cmd.SafeProxyAddress)
	suite.threshold = envTypes.GetSafeThreshold()
	suite.ctx = ctx

	// Bridge values
	bridgeLckAddress, minCreateAmount, signatureReward := envTypes.GetBridgeValues()
	if bridgeLckAddress != nil {
		suite.bridgeConfig = &XChainTypesBridgeConfig{
			LockingChainDoor: *bridgeLckAddress,
			LockingChainIssue: XChainTypesBridgeChainIssue{
				Issuer:   zeroAddress,
				Currency: "XRP",
			},
			IssuingChainDoor: suite.safeAddress,
			IssuingChainIssue: XChainTypesBridgeChainIssue{
				Issuer:   zeroAddress,
				Currency: "XRP",
			},
		}

		suite.bridgeParams = &XChainTypesBridgeParams{
			MinCreateAmount: minCreateAmount,
			SignatureReward: signatureReward,
		}

		suite.bridgeKey = envTypes.GetBridgeKey()
		suite.bridgeTests = true
	} else {
		suite.bridgeTests = false
	}

	suite.claims = 0

	return nil
}

func (suite *BridgeTestSuite) RunTests() {
	if suite.bridgeTests {
		suite.t.Logf("Running bridge tests...")

		suite.runAddClaimAttestationTest()
		suite.runAddCreateAccountAttestationTest()
		suite.runClaimTest()
		suite.runCommitTest()
		suite.runCommitWithoutAddressTest()
		suite.runCreateAccountCommitTest()
		suite.runCreateClaimIdTest()
		suite.runGetBridgeClaimTest(*suite.bridgeConfig, big.NewInt(0), zeroAddress, zeroAddress, false)
		suite.runGetBridgeConfigTest()
		suite.runGetBridgeCreateAccountTest(*suite.bridgeConfig, zeroAddress, big.NewInt(0), false, false)
		suite.runGetBridgeKeyTest(*suite.bridgeConfig, suite.bridgeKey)
		suite.runGetBridgeParamsTest(*suite.bridgeConfig, suite.bridgeParams.MinCreateAmount, suite.bridgeParams.SignatureReward)
		suite.runGetBridgeTokenFailTest(*suite.bridgeConfig)
		suite.runGetBridgesPaginatedTest(1)
	} else {
		suite.runGetBridgesPaginatedTest(0)
	}

	suite.runCreateBridgeTest()
	suite.runCreateBridgeRequestTest()
	suite.runWitnessesTest()
	suite.runTokenRegisteredTest("0x1", false)
	suite.runOwnerTest()
	suite.runPauseTest()
	suite.runPausedTest(false)
	suite.runSafeTest()
	suite.runUnpauseTest()
}
