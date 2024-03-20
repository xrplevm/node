package bridge

import (
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func (suite *BridgeTestSuite) runGetBridgeClaimTest(bridgeCfg XChainTypesBridgeConfig, claimId *big.Int, expCreator, expSender common.Address, expExists bool) {
	creator, sender, exists, err := suite.bridge.GetBridgeClaim(suite.claimer.CallOpts, bridgeCfg, claimId)
	if err != nil {
		suite.t.Errorf("Error getting bridge claim: '%+v'", err)
	} else {
		if creator.Hex() != expCreator.Hex() {
			suite.t.Errorf("Invalid bridge claim creator value - expected '%+v' got '%+v'", expCreator.Hex(), creator.Hex())
		}
		if sender.Hex() != expSender.Hex() {
			suite.t.Errorf("Invalid bridge claim sender value - expected '%+v' got '%+v'", expSender.Hex(), sender.Hex())
		}
		if exists != expExists {
			suite.t.Errorf("Invalid bridge claim exists value - expected '%+v' got '%+v'", expExists, exists)
		}
	}
}

func (suite *BridgeTestSuite) runGetBridgeConfigTest() {
	lckAddr, lckIssAddr, lckIssCurr, issAddr, issIssAddr, issIssCurr, err := suite.bridge.GetBridgeConfig(suite.claimer.CallOpts, *suite.bridgeConfig)
	if err != nil {
		suite.t.Errorf("Error getting bridge config: '%+v'", err)
	} else {
		if lckAddr.Hex() != suite.bridgeConfig.LockingChainDoor.Hex() {
			suite.t.Errorf("Invalid locking chain door value - expected '%+v' got '%+v'", suite.bridgeConfig.LockingChainDoor.Hex(), lckAddr.Hex())
		}
		if lckIssAddr.Hex() != suite.bridgeConfig.LockingChainIssue.Issuer.Hex() {
			suite.t.Errorf("Invalid locking chain issue issuer value - expected '%+v' got '%+v'", suite.bridgeConfig.LockingChainIssue.Issuer.Hex(), lckIssAddr.Hex())
		}
		if lckIssCurr != suite.bridgeConfig.LockingChainIssue.Currency {
			suite.t.Errorf("Invalid locking chain issue currency value - expected '%+v' got '%+v'", suite.bridgeConfig.LockingChainIssue.Currency, lckIssCurr)
		}
		if issAddr.Hex() != suite.bridgeConfig.IssuingChainDoor.Hex() {
			suite.t.Errorf("Invalid issuing chain door value - expected '%+v' got '%+v'", suite.bridgeConfig.IssuingChainDoor.Hex(), issAddr.Hex())
		}
		if issIssAddr.Hex() != suite.bridgeConfig.IssuingChainIssue.Issuer.Hex() {
			suite.t.Errorf("Invalid issuing chain issue issuer value - expected '%+v' got '%+v'", suite.bridgeConfig.IssuingChainIssue.Issuer.Hex(), issIssAddr.Hex())
		}
		if issIssCurr != suite.bridgeConfig.IssuingChainIssue.Currency {
			suite.t.Errorf("Invalid issuing chain issue currency value - expected '%+v' got '%+v'", suite.bridgeConfig.IssuingChainIssue.Currency, issIssCurr)
		}
	}
}

func (suite *BridgeTestSuite) runGetBridgeCreateAccountTest(bridgeCfg XChainTypesBridgeConfig, address common.Address, expSigReward *big.Int, expIsCreated, expExists bool) {
	sigReward, isCreated, exists, err := suite.bridge.GetBridgeCreateAccount(suite.claimer.CallOpts, bridgeCfg, address)
	if err != nil {
		suite.t.Errorf("Error getting bridge create account: '%+v'", err)
	} else {
		if sigReward.Cmp(expSigReward) != 0 {
			suite.t.Errorf("Invalid bridge create account signature reward value - expected '%+v' got '%+v'", expSigReward, sigReward)
		}
		if isCreated != expIsCreated {
			suite.t.Errorf("Invalid bridge create account isCreated value - expected '%+v' got '%+v'", expIsCreated, isCreated)
		}
		if exists != expExists {
			suite.t.Errorf("Invalid bridge create account exists value - expected '%+v' got '%+v'", expExists, exists)
		}
	}
}

func (suite *BridgeTestSuite) runGetBridgeKeyTest(bridgeCfg XChainTypesBridgeConfig, expected string) {
	key, err := suite.bridge.GetBridgeKey(suite.claimer.CallOpts, bridgeCfg)
	if err != nil {
		suite.t.Errorf("Error getting bridge key: '%+v'", err)
	} else {
		if hex.EncodeToString(key[:]) != expected {
			suite.t.Errorf("Invalid bridge key value - expected '%+v' got '%+v'", expected, hex.EncodeToString(key[:]))
		}
	}
}

func (suite *BridgeTestSuite) runGetBridgeParamsTest(bridgeCfg XChainTypesBridgeConfig, expMinCreateAmount, expSigReward *big.Int) {
	minCreateAmount, signatureReward, err := suite.bridge.GetBridgeParams(suite.claimer.CallOpts, bridgeCfg)
	if err != nil {
		suite.t.Errorf("Error getting bridge params: '%+v'", err)
	} else {
		if minCreateAmount.Cmp(expMinCreateAmount) != 0 {
			suite.t.Errorf("Invalid bridge min create amount - expected '%+v' got '%+v'", expMinCreateAmount, minCreateAmount)
		}
		if signatureReward.Cmp(expSigReward) != 0 {
			suite.t.Errorf("Invalid bridge signature reward - expected '%+v' got '%+v'", expSigReward, signatureReward)
		}
	}
}

func (suite *BridgeTestSuite) runGetBridgeTokenTest(bridgeCfg XChainTypesBridgeConfig, expected common.Address) {
	tokenAddress, err := suite.bridge.GetBridgeToken(suite.claimer.CallOpts, bridgeCfg)
	if err != nil {
		suite.t.Errorf("Error getting bridge token: '%+v'", err)
	} else {
		if tokenAddress.Hex() != expected.Hex() {
			suite.t.Errorf("Invalid bridge token value - expected '%+v' got '%+v'", expected.Hex(), tokenAddress.Hex())
		}
	}
}

func (suite *BridgeTestSuite) runGetBridgeTokenFailTest(bridgeCfg XChainTypesBridgeConfig) {
	_, err := suite.bridge.GetBridgeToken(suite.claimer.CallOpts, bridgeCfg)
	if err == nil {
		suite.t.Errorf("Getting bridge token should be reverted bridge: '%+v'", bridgeCfg)
	} else {
		if err.Error() != "execution reverted" {
			suite.t.Errorf("Invalid bridge token revert value - expected '%+v' got '%+v'", "execution reverted", err.Error())
		}
	}
}

func (suite *BridgeTestSuite) runGetBridgesPaginatedTest(expectedLength int) {
	bridgeKeysArrMemBytes, err := suite.bridge.GetBridgesPaginated(suite.claimer.CallOpts, big.NewInt(0))
	if err != nil {
		suite.t.Errorf("Error getting bridges paginated: '%+v'", err)
	} else {
		configsLength := 0
		for _, config := range bridgeKeysArrMemBytes.Configs {
			if config.IssuingChainDoor.Hex() != zeroAddress.Hex() {
				configsLength += 1
			}
		}
		if configsLength != expectedLength {
			suite.t.Errorf("Invalid bridge configs length value - expected '%+v' got '%+v'", expectedLength, configsLength)
		}

		paramsLength := 0
		for _, params := range bridgeKeysArrMemBytes.Params {
			if params.SignatureReward.Cmp(big.NewInt(0)) != 0 {
				paramsLength += 1
			}
		}
		if paramsLength != expectedLength {
			suite.t.Errorf("Invalid bridge params length value - expected '%+v' got '%+v'", expectedLength, paramsLength)
		}
	}
}

func (suite *BridgeTestSuite) runTokenRegisteredTest(tokenAddress string, expectedValue bool) {
	isRegistered, err := suite.bridge.IsTokenRegistered(suite.claimer.CallOpts, common.HexToAddress(tokenAddress))
	if err != nil {
		suite.t.Errorf("Error getting is token registered: '%+v'", err)
	} else if isRegistered != expectedValue {
		suite.t.Errorf("Invalid IsTokenRegistered (%+v) value - expected '%+v' got '%+v'", tokenAddress, expectedValue, isRegistered)
	}
}

func (suite *BridgeTestSuite) runWitnessesTest() {
	witnesses, err := suite.bridge.GetWitnesses(suite.claimer.CallOpts)
	if err != nil {
		suite.t.Errorf("Error getting witnesses: '%+v'", err)
	} else {
		if len(witnesses) != len(suite.witnesses) {
			suite.t.Errorf("Invalid witnesses length - expected '%+v' got '%+v'", len(suite.witnesses), len(witnesses))
		}
		for _, expWitness := range suite.witnesses {
			found := false
			for _, witness := range witnesses {
				if witness.Hex() == expWitness.EvmAddress.Hex() {
					found = true
				}
			}
			if !found {
				suite.t.Errorf("Witness not found - expected '%+v'", expWitness.EvmAddress.Hex())
			}
		}
	}
}

func (suite *BridgeTestSuite) runOwnerTest() {
	owner, err := suite.bridge.Owner(suite.claimer.CallOpts)
	if err != nil {
		suite.t.Errorf("Error getting owner: '%+v'", err)
	} else if owner.Hex() != suite.safeAddress.Hex() {
		suite.t.Errorf("Invalid owner - expected '%+v' got '%+v'", suite.safeAddress.Hex(), owner.Hex())
	}
}

func (suite *BridgeTestSuite) runPausedTest(expect bool) {
	paused, err := suite.bridge.Paused(suite.claimer.CallOpts)
	if err != nil {
		suite.t.Errorf("Error getting paused: '%+v'", err)
	} else if paused != expect {
		suite.t.Errorf("Invalid paused - expected '%+v' got '%+v'", expect, paused)
	}
}

func (suite *BridgeTestSuite) runSafeTest() {
	safe, err := suite.bridge.Safe(suite.claimer.CallOpts)
	if err != nil {
		suite.t.Errorf("Error getting safe: '%+v'", err)
	} else if safe.Hex() != suite.safeAddress.Hex() {
		suite.t.Errorf("Invalid safe - expected '%+v' got '%+v'", suite.safeAddress.Hex(), safe.Hex())
	}
}
