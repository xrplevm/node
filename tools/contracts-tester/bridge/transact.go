package bridge

import (
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func (suite *BridgeTestSuite) runAddClaimAttestationTest() {
	sender := common.HexToAddress("0x11")
	destination := common.HexToAddress("0x22")
	amount := new(big.Int).Mul(big.NewInt(10), valueMultiplier)

	// Create claim to attest
	transactOpts := bind.TransactOpts{Signer: suite.claimer.Signer, From: suite.claimer.EvmAddress, GasLimit: 0, Value: suite.bridgeParams.SignatureReward}
	tx, err := suite.bridge.CreateClaimId(&transactOpts, *suite.bridgeConfig, sender)
	if err != nil {
		suite.t.Errorf("Error creating create claim id transaction (add claim att): '%+v'", err)
		return
	}
	suite.waitForTransaction(tx.Hash())
	suite.claims++
	suite.runCreateClaimEventTest(suite.claims)
	event := suite.getLatestCreateClaimEvent()

	// Should revert when not witness
	transactOpts.Value = nil
	tx, err = suite.bridge.AddClaimAttestation(&transactOpts, *suite.bridgeConfig, event.ClaimId, amount, sender, destination)
	if err == nil {
		suite.t.Errorf("Error creating add claim attestation transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runAddClaimAttestationEventTest(0)
	}

	// Should revert when claim does not exist
	transactOpts = bind.TransactOpts{Signer: suite.witnesses[0].Signer, From: suite.witnesses[0].EvmAddress, GasLimit: 0}
	tx, err = suite.bridge.AddClaimAttestation(&transactOpts, *suite.bridgeConfig, big.NewInt(10000), amount, sender, destination)
	if err == nil {
		suite.t.Errorf("Error creating add claim attestation transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runAddClaimAttestationEventTest(0)
	}

	// Should revert when invalid claim sender
	tx, err = suite.bridge.AddClaimAttestation(&transactOpts, *suite.bridgeConfig, event.ClaimId, amount, event.Creator, destination)
	if err == nil {
		suite.t.Errorf("Error creating add claim attestation transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runAddClaimAttestationEventTest(0)
	}

	// Should work and emit addClaimAttestation event
	for i := 0; i < suite.threshold; i += 1 {
		transactOpts = bind.TransactOpts{Signer: suite.witnesses[i].Signer, From: suite.witnesses[i].EvmAddress, GasLimit: 0}

		tx, err = suite.bridge.AddClaimAttestation(&transactOpts, *suite.bridgeConfig, event.ClaimId, amount, sender, destination)
		if err != nil {
			suite.t.Errorf("Error creating add claim attestation transaction: '%+v'", err)
		} else {
			suite.waitForTransaction(tx.Hash())
			suite.runAddClaimAttestationEventTest(i + 1)
		}
	}

	// When threshold reached should emit credit event
	suite.runCreditEventTest(1)
}

func (suite *BridgeTestSuite) runAddCreateAccountAttestationTest() {
	destination := common.HexToAddress("0x20")
	amount := suite.bridgeParams.MinCreateAmount
	sigReward := suite.bridgeParams.SignatureReward

	// Should revert when not witness
	transactOpts := bind.TransactOpts{Signer: suite.claimer.Signer, From: suite.claimer.EvmAddress, GasLimit: 0}
	tx, err := suite.bridge.AddCreateAccountAttestation(&transactOpts, *suite.bridgeConfig, destination, amount, sigReward)
	if err == nil {
		suite.t.Errorf("Error creating add create account attestation transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runAddCreateAccountAttestationEventTest(0)
		suite.runGetBridgeCreateAccountTest(*suite.bridgeConfig, destination, big.NewInt(0), false, false)
	}

	// Should revert when less than min create amount
	transactOpts = bind.TransactOpts{Signer: suite.witnesses[0].Signer, From: suite.witnesses[0].EvmAddress, GasLimit: 0}
	tx, err = suite.bridge.AddCreateAccountAttestation(&transactOpts, *suite.bridgeConfig, destination, big.NewInt(0), sigReward)
	if err == nil {
		suite.t.Errorf("Error creating add create account attestation transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runAddCreateAccountAttestationEventTest(0)
		suite.runGetBridgeCreateAccountTest(*suite.bridgeConfig, destination, big.NewInt(0), false, false)
	}

	// Should work and emit addCreateAccountAttestation event
	for i := 0; i < suite.threshold; i += 1 {
		transactOpts = bind.TransactOpts{Signer: suite.witnesses[i].Signer, From: suite.witnesses[i].EvmAddress, GasLimit: 0}
		tx, err = suite.bridge.AddCreateAccountAttestation(&transactOpts, *suite.bridgeConfig, destination, amount, sigReward)
		if err != nil {
			suite.t.Errorf("Error creating add create account attestation transaction: '%+v'", err)
		} else {
			suite.waitForTransaction(tx.Hash())
			suite.runAddCreateAccountAttestationEventTest(i + 1)
		}
	}

	// When threshold reached should emit create account event
	suite.runCreateAccountEventTest(1)
	suite.runAddCreateAccountAttestationEventTest(suite.threshold)
	suite.runGetBridgeCreateAccountTest(*suite.bridgeConfig, destination, sigReward, true, true)

	// Should revert when account already created
	transactOpts = bind.TransactOpts{Signer: suite.witnesses[0].Signer, From: suite.witnesses[0].EvmAddress, GasLimit: 0}
	tx, err = suite.bridge.AddCreateAccountAttestation(&transactOpts, *suite.bridgeConfig, destination, amount, sigReward)
	if err == nil {
		suite.t.Errorf("Error creating add create account attestation transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runAddCreateAccountAttestationEventTest(suite.threshold)
	}
}

func (suite *BridgeTestSuite) runClaimTest() {
	sender := common.HexToAddress("0x15")
	destination := common.HexToAddress("0x25")
	amount := new(big.Int).Mul(big.NewInt(15), valueMultiplier)

	// Create claim to attest
	transactOpts := bind.TransactOpts{Signer: suite.claimer.Signer, From: suite.claimer.EvmAddress, GasLimit: 0, Value: suite.bridgeParams.SignatureReward}
	tx, err := suite.bridge.CreateClaimId(&transactOpts, *suite.bridgeConfig, sender)
	if err != nil {
		suite.t.Errorf("Error creating create claim id transaction (claim): '%+v'", err)
		return
	}
	suite.waitForTransaction(tx.Hash())
	suite.claims++
	suite.runCreateClaimEventTest(suite.claims)
	event := suite.getLatestCreateClaimEvent()

	// Should revert when claim does not exist
	transactOpts.Value = nil
	tx, err = suite.bridge.Claim(&transactOpts, *suite.bridgeConfig, big.NewInt(100000), amount, destination)
	if err == nil {
		suite.t.Errorf("Error creating claim transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runClaimEventTest(0)
	}

	// Should revert when claimer is not creator
	transactOpts = bind.TransactOpts{Signer: suite.witnesses[0].Signer, From: suite.witnesses[0].EvmAddress, GasLimit: 0}
	tx, err = suite.bridge.Claim(&transactOpts, *suite.bridgeConfig, event.ClaimId, amount, destination)
	if err == nil {
		suite.t.Errorf("Error creating claim transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runClaimEventTest(0)
	}

	// Attestate claim
	for i := 0; i < suite.threshold; i += 1 {
		if i == 1 {
			// Should revert when there is no consensus
			transactOpts = bind.TransactOpts{Signer: suite.claimer.Signer, From: suite.claimer.EvmAddress, GasLimit: 0}
			tx, err = suite.bridge.Claim(&transactOpts, *suite.bridgeConfig, event.ClaimId, amount, destination)
			if err == nil {
				suite.t.Errorf("Error creating claim transaction, should be reverted but success: '%+v'", tx)
			} else {
				if !strings.Contains(err.Error(), executionRevertedStr) {
					suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
				}
				suite.runClaimEventTest(0)
			}
		}

		transactOpts = bind.TransactOpts{Signer: suite.witnesses[i].Signer, From: suite.witnesses[i].EvmAddress, GasLimit: 0}
		tx, err = suite.bridge.AddClaimAttestation(&transactOpts, *suite.bridgeConfig, event.ClaimId, amount, sender, zeroAddress)
		if err != nil {
			suite.t.Errorf("Error creating add claim attestation transaction (claim): '%+v'", err)
		} else {
			suite.waitForTransaction(tx.Hash())
			// suite.runAddClaimAttestationEventTest(i + 1)
		}
	}

	// Should revert when amount is not attested amount
	transactOpts = bind.TransactOpts{Signer: suite.claimer.Signer, From: suite.claimer.EvmAddress, GasLimit: 0}
	tx, err = suite.bridge.Claim(&transactOpts, *suite.bridgeConfig, event.ClaimId, big.NewInt(100), destination)
	if err == nil {
		suite.t.Errorf("Error creating claim transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runClaimEventTest(0)
	}

	// Should work correctly and emit claim event
	tx, err = suite.bridge.Claim(&transactOpts, *suite.bridgeConfig, event.ClaimId, amount, destination)
	if err != nil {
		suite.t.Errorf("Error creating claim transaction: '%+v'", err)
	} else {
		suite.waitForTransaction(tx.Hash())
		suite.runClaimEventTest(1)
	}
}

func (suite *BridgeTestSuite) runCommitTest() {
	// Should work but no event is created
	transactOpts := bind.TransactOpts{Signer: suite.claimer.Signer, From: suite.claimer.EvmAddress, GasLimit: 0}
	tx, err := suite.bridge.Commit(&transactOpts, *suite.bridgeConfig, common.HexToAddress("0x2"), big.NewInt(0), big.NewInt(0))
	if err != nil {
		suite.t.Errorf("Error creating commit transaction: '%+v'", err)
	} else {
		suite.waitForTransaction(tx.Hash())
		suite.runCommitEventTest(0)
	}

	// Should revert if value is not enough
	value := new(big.Int).Mul(big.NewInt(20), valueMultiplier)
	amount := new(big.Int).Mul(big.NewInt(25), valueMultiplier)
	transactOpts.Value = value
	tx, err = suite.bridge.Commit(&transactOpts, *suite.bridgeConfig, common.HexToAddress("0x2"), big.NewInt(1), amount)
	if err == nil {
		suite.t.Errorf("Error creating commit transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runCommitEventTest(0)
	}

	// Should work and emit new event
	tx, err = suite.bridge.Commit(&transactOpts, *suite.bridgeConfig, common.HexToAddress("0x2"), big.NewInt(2), value)
	if err != nil {
		suite.t.Errorf("Error creating commit transaction: '%+v'", err)
	} else {
		suite.waitForTransaction(tx.Hash())
		suite.runCommitEventTest(1)
	}
}

func (suite *BridgeTestSuite) runCommitWithoutAddressTest() {
	// Should work but no event is created
	transactOpts := bind.TransactOpts{Signer: suite.claimer.Signer, From: suite.claimer.EvmAddress, GasLimit: 0}
	tx, err := suite.bridge.CommitWithoutAddress(&transactOpts, *suite.bridgeConfig, big.NewInt(0), big.NewInt(0))
	if err != nil {
		suite.t.Errorf("Error creating commit without address transaction: '%+v'", err)
	} else {
		suite.waitForTransaction(tx.Hash())
		suite.runCommitWithoutAddressEventTest(0)
	}

	// Should revert if value is not enough
	value := new(big.Int).Mul(big.NewInt(20), valueMultiplier)
	amount := new(big.Int).Mul(big.NewInt(25), valueMultiplier)
	transactOpts.Value = value
	tx, err = suite.bridge.CommitWithoutAddress(&transactOpts, *suite.bridgeConfig, big.NewInt(1), amount)
	if err == nil {
		suite.t.Errorf("Error creating commit without address transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runCommitWithoutAddressEventTest(0)
	}

	// Should work and emit new event
	tx, err = suite.bridge.CommitWithoutAddress(&transactOpts, *suite.bridgeConfig, big.NewInt(2), value)
	if err != nil {
		suite.t.Errorf("Error creating commit without address transaction: '%+v'", err)
	} else {
		suite.waitForTransaction(tx.Hash())
		suite.runCommitWithoutAddressEventTest(1)
	}
}

func (suite *BridgeTestSuite) runCreateAccountCommitTest() {
	destination := common.HexToAddress("0x1")
	amount := suite.bridgeParams.MinCreateAmount
	sigReward := suite.bridgeParams.SignatureReward

	// Should revert if signature reward is not enough
	transactOpts := bind.TransactOpts{Signer: suite.claimer.Signer, From: suite.claimer.EvmAddress, GasLimit: 0}
	tx, err := suite.bridge.CreateAccountCommit(&transactOpts, *suite.bridgeConfig, destination, amount, big.NewInt(0))
	if err == nil {
		suite.t.Errorf("Error creating create account commit transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runCreateAccountCommitEventTest(0)
	}

	// Should revert if amount is not enough
	tx, err = suite.bridge.CreateAccountCommit(&transactOpts, *suite.bridgeConfig, destination, big.NewInt(0), sigReward)
	if err == nil {
		suite.t.Errorf("Error creating create account commit transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runCreateAccountCommitEventTest(0)
	}

	// Should revert if sent amount is not enough
	tx, err = suite.bridge.CreateAccountCommit(&transactOpts, *suite.bridgeConfig, destination, amount, sigReward)
	if err == nil {
		suite.t.Errorf("Error creating create account commit transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runCreateAccountCommitEventTest(0)
	}

	// Should Work and emit new event
	transactOpts.Value = new(big.Int).Add(amount, sigReward)
	tx, err = suite.bridge.CreateAccountCommit(&transactOpts, *suite.bridgeConfig, destination, amount, sigReward)
	if err != nil {
		suite.t.Errorf("Error creating create account commit transaction: '%+v'", err)
	} else {
		suite.waitForTransaction(tx.Hash())
		suite.runCreateAccountCommitEventTest(1)
	}
}

func (suite *BridgeTestSuite) runCreateClaimIdTest() {
	sender := common.HexToAddress("0x1")

	// Should revert if signature reward is not enough
	transactOpts := bind.TransactOpts{Signer: suite.claimer.Signer, From: suite.claimer.EvmAddress, GasLimit: 0}
	tx, err := suite.bridge.CreateClaimId(&transactOpts, *suite.bridgeConfig, sender)
	if err == nil {
		suite.t.Errorf("Error creating create claim id transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runCreateClaimEventTest(suite.claims)
	}

	// Should Work and emit new event
	transactOpts.Value = suite.bridgeParams.SignatureReward
	tx, err = suite.bridge.CreateClaimId(&transactOpts, *suite.bridgeConfig, sender)
	if err != nil {
		suite.t.Errorf("Error creating create claim id transaction: '%+v'", err)
	} else {
		suite.waitForTransaction(tx.Hash())
		suite.claims++
		suite.runCreateClaimEventTest(suite.claims)
		event := suite.getLatestCreateClaimEvent()
		suite.runGetBridgeClaimTest(*suite.bridgeConfig, event.ClaimId, suite.claimer.EvmAddress, sender, true)
	}
}

func (suite *BridgeTestSuite) runCreateBridgeTest() {
	bridgeConfig := XChainTypesBridgeConfig{
		LockingChainDoor: common.HexToAddress("0x100"),
		LockingChainIssue: XChainTypesBridgeChainIssue{
			Issuer:   common.HexToAddress("0x101"),
			Currency: "MET",
		},
		IssuingChainDoor: common.HexToAddress("0x102"),
		IssuingChainIssue: XChainTypesBridgeChainIssue{
			Issuer:   common.HexToAddress("0x103"),
			Currency: "MET",
		},
	}
	sigReward, _ := new(big.Int).SetString("500000000000000000", 10)
	bridgeParams := XChainTypesBridgeParams{big.NewInt(0), sigReward}

	transactOpts := bind.TransactOpts{Signer: suite.claimer.Signer, From: suite.claimer.EvmAddress, GasLimit: 0}
	tx, err := suite.bridge.CreateBridge(&transactOpts, bridgeConfig, bridgeParams)
	if err == nil {
		suite.t.Errorf("Error creating bridge request transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runCreateBridgeEventTest(0)
	}

	transactOpts = bind.TransactOpts{Signer: suite.witnesses[0].Signer, From: suite.witnesses[0].EvmAddress, GasLimit: 0}
	tx, err = suite.bridge.CreateBridge(&transactOpts, bridgeConfig, bridgeParams)
	if err == nil {
		suite.t.Errorf("Error creating bridge request transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runCreateBridgeEventTest(0)
	}
}

func (suite *BridgeTestSuite) runCreateBridgeRequestTest() {
	// Should revert if insufficient value
	transactOpts := bind.TransactOpts{Signer: suite.claimer.Signer, From: suite.claimer.EvmAddress, GasLimit: 0}
	tx, err := suite.bridge.CreateBridgeRequest(&transactOpts, common.HexToAddress("0x1"))
	if err == nil {
		suite.t.Errorf("Error creating bridge request transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runCreateBridgeRequestEventTest(0)
	}

	// Should get MINCREATEBRIDGEREWARD correctly
	minReward, err := suite.bridge.MINCREATEBRIDGEREWARD(suite.claimer.CallOpts)
	if err != nil {
		suite.t.Errorf("Error retrieving MINCREATEBRIDGEREWARD: '%+v'", err)
		return
	}

	// Should revert if token already exists
	transactOpts.Value = minReward
	tx, err = suite.bridge.CreateBridgeRequest(&transactOpts, zeroAddress)
	if err == nil {
		suite.t.Errorf("Error creating bridge request transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runCreateBridgeRequestEventTest(0)
	}

	// Works correctly
	tx, err = suite.bridge.CreateBridgeRequest(&transactOpts, common.HexToAddress("0x1"))
	if err != nil {
		suite.t.Errorf("Error creating bridge request transaction: '%+v'", err)
	} else {
		suite.waitForTransaction(tx.Hash())
		suite.runCreateBridgeRequestEventTest(1)
	}
}

func (suite *BridgeTestSuite) runPauseTest() {
	// Should revert caller is not owner
	transactOpts := bind.TransactOpts{Signer: suite.claimer.Signer, From: suite.claimer.EvmAddress, GasLimit: 0}
	tx, err := suite.bridge.Pause(&transactOpts)
	if err == nil {
		suite.t.Errorf("Error creating pause transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runPausedEventTest(0)
	}

	// Should revert caller is not owner
	transactOpts = bind.TransactOpts{Signer: suite.witnesses[0].Signer, From: suite.witnesses[0].EvmAddress, GasLimit: 0}
	tx, err = suite.bridge.Pause(&transactOpts)
	if err == nil {
		suite.t.Errorf("Error creating pause transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runPausedEventTest(0)
	}
	suite.runPausedTest(false)
}

func (suite *BridgeTestSuite) runUnpauseTest() {
	// Should revert caller is not owner
	transactOpts := bind.TransactOpts{Signer: suite.claimer.Signer, From: suite.claimer.EvmAddress, GasLimit: 0}
	tx, err := suite.bridge.Pause(&transactOpts)
	if err == nil {
		suite.t.Errorf("Error creating unpause transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runUnpausedEventTest(0)
	}

	// Should revert caller is not owner
	transactOpts = bind.TransactOpts{Signer: suite.witnesses[0].Signer, From: suite.witnesses[0].EvmAddress, GasLimit: 0}
	tx, err = suite.bridge.Pause(&transactOpts)
	if err == nil {
		suite.t.Errorf("Error creating unpause transaction, should be reverted but success: '%+v'", tx)
	} else {
		if !strings.Contains(err.Error(), executionRevertedStr) {
			suite.t.Errorf("Invalid error value - expected to contain '%+v', got '%+v'", executionRevertedStr, err.Error())
		}
		suite.runUnpausedEventTest(0)
	}
	suite.runPausedTest(false)
}

func (suite *BridgeTestSuite) waitForTransaction(txHash common.Hash) {
	var err error
	isPending := true

	for isPending {
		_, isPending, err = suite.client.TransactionByHash(suite.ctx, txHash)
		if err != nil && err.Error() != "not found" {
			suite.t.Errorf("Error getting transaction by hash: '%+v'", err)
			break
		}

		time.Sleep(time.Second)
	}
}
