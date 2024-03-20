// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bridge

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// XChainTypesBridgeChainIssue is an auto generated low-level Go binding around an user-defined struct.
type XChainTypesBridgeChainIssue struct {
	Issuer   common.Address
	Currency string
}

// XChainTypesBridgeConfig is an auto generated low-level Go binding around an user-defined struct.
type XChainTypesBridgeConfig struct {
	LockingChainDoor  common.Address
	LockingChainIssue XChainTypesBridgeChainIssue
	IssuingChainDoor  common.Address
	IssuingChainIssue XChainTypesBridgeChainIssue
}

// XChainTypesBridgeParams is an auto generated low-level Go binding around an user-defined struct.
type XChainTypesBridgeParams struct {
	MinCreateAmount *big.Int
	SignatureReward *big.Int
}

// BridgeMetaData contains all meta data concerning the Bridge contract.
var BridgeMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"AccountAlreadyCreated\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"BridgeAlreadyRegistered\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"BridgeNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CallerIsNotCreator\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CallerIsNotWitness\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ClaimNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InsufficientAmount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InsufficientReward\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidAmount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidOperation\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSender\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSentAmount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NoBridgeToken\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SendError\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TokenAlreadyRegistered\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"bridgeKey\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"claimId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"witness\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"}],\"name\":\"AddClaimAttestation\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"bridgeKey\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"witness\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"AddCreateAccountAttestation\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"previousAdmin\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"AdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"beacon\",\"type\":\"address\"}],\"name\":\"BeaconUpgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"bridgeKey\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"claimId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"}],\"name\":\"Claim\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"bridgeKey\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"claimId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"}],\"name\":\"Commit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"bridgeKey\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"claimId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"CommitWithoutAddress\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"CreateAccount\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"bridgeKey\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"creator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"signatureReward\",\"type\":\"uint256\"}],\"name\":\"CreateAccountCommit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"bridgeKey\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"lockingChainDoor\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"lockingChainIssueIssuer\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"lockingChainIssueCurrency\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"issuingChainDoor\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"issuingChainIssueIssuer\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"issuingChainIssueCurrency\",\"type\":\"string\"}],\"name\":\"CreateBridge\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"}],\"name\":\"CreateBridgeRequest\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"bridgeKey\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"claimId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"creator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"CreateClaim\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"bridgeKey\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"claimId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Credit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"Paused\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"Unpaused\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"MIN_CREATE_BRIDGE_REWARD\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractGnosisSafeL2\",\"name\":\"safe\",\"type\":\"address\"}],\"name\":\"__BridgeDoorCommon_init\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"__Manageable_init\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"_safe\",\"outputs\":[{\"internalType\":\"contractGnosisSafeL2\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"lockingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"lockingChainIssue\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"issuingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"issuingChainIssue\",\"type\":\"tuple\"}],\"internalType\":\"structXChainTypes.BridgeConfig\",\"name\":\"bridgeConfig\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"claimId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"}],\"name\":\"addClaimAttestation\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"lockingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"lockingChainIssue\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"issuingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"issuingChainIssue\",\"type\":\"tuple\"}],\"internalType\":\"structXChainTypes.BridgeConfig\",\"name\":\"bridgeConfig\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"signatureReward\",\"type\":\"uint256\"}],\"name\":\"addCreateAccountAttestation\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"lockingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"lockingChainIssue\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"issuingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"issuingChainIssue\",\"type\":\"tuple\"}],\"internalType\":\"structXChainTypes.BridgeConfig\",\"name\":\"bridgeConfig\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"claimId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"}],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"lockingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"lockingChainIssue\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"issuingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"issuingChainIssue\",\"type\":\"tuple\"}],\"internalType\":\"structXChainTypes.BridgeConfig\",\"name\":\"bridgeConfig\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"claimId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"commit\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"lockingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"lockingChainIssue\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"issuingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"issuingChainIssue\",\"type\":\"tuple\"}],\"internalType\":\"structXChainTypes.BridgeConfig\",\"name\":\"bridgeConfig\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"claimId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"commitWithoutAddress\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"lockingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"lockingChainIssue\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"issuingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"issuingChainIssue\",\"type\":\"tuple\"}],\"internalType\":\"structXChainTypes.BridgeConfig\",\"name\":\"bridgeConfig\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"signatureReward\",\"type\":\"uint256\"}],\"name\":\"createAccountCommit\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"lockingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"lockingChainIssue\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"issuingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"issuingChainIssue\",\"type\":\"tuple\"}],\"internalType\":\"structXChainTypes.BridgeConfig\",\"name\":\"config\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"minCreateAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"signatureReward\",\"type\":\"uint256\"}],\"internalType\":\"structXChainTypes.BridgeParams\",\"name\":\"params\",\"type\":\"tuple\"}],\"name\":\"createBridge\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"}],\"name\":\"createBridgeRequest\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"lockingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"lockingChainIssue\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"issuingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"issuingChainIssue\",\"type\":\"tuple\"}],\"internalType\":\"structXChainTypes.BridgeConfig\",\"name\":\"bridgeConfig\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"createClaimId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"enumEnum.Operation\",\"name\":\"operation\",\"type\":\"uint8\"}],\"name\":\"execute\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"lockingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"lockingChainIssue\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"issuingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"issuingChainIssue\",\"type\":\"tuple\"}],\"internalType\":\"structXChainTypes.BridgeConfig\",\"name\":\"bridgeConfig\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"claimId\",\"type\":\"uint256\"}],\"name\":\"getBridgeClaim\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"lockingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"lockingChainIssue\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"issuingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"issuingChainIssue\",\"type\":\"tuple\"}],\"internalType\":\"structXChainTypes.BridgeConfig\",\"name\":\"bridgeConfig\",\"type\":\"tuple\"}],\"name\":\"getBridgeConfig\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"lockingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"lockingChainIssue\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"issuingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"issuingChainIssue\",\"type\":\"tuple\"}],\"internalType\":\"structXChainTypes.BridgeConfig\",\"name\":\"bridgeConfig\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"getBridgeCreateAccount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"lockingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"lockingChainIssue\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"issuingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"issuingChainIssue\",\"type\":\"tuple\"}],\"internalType\":\"structXChainTypes.BridgeConfig\",\"name\":\"bridgeConfig\",\"type\":\"tuple\"}],\"name\":\"getBridgeKey\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"lockingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"lockingChainIssue\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"issuingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"issuingChainIssue\",\"type\":\"tuple\"}],\"internalType\":\"structXChainTypes.BridgeConfig\",\"name\":\"bridgeConfig\",\"type\":\"tuple\"}],\"name\":\"getBridgeParams\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"lockingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"lockingChainIssue\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"issuingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"issuingChainIssue\",\"type\":\"tuple\"}],\"internalType\":\"structXChainTypes.BridgeConfig\",\"name\":\"bridgeConfig\",\"type\":\"tuple\"}],\"name\":\"getBridgeToken\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"page\",\"type\":\"uint256\"}],\"name\":\"getBridgesPaginated\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"lockingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"lockingChainIssue\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"issuingChainDoor\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"currency\",\"type\":\"string\"}],\"internalType\":\"structXChainTypes.BridgeChainIssue\",\"name\":\"issuingChainIssue\",\"type\":\"tuple\"}],\"internalType\":\"structXChainTypes.BridgeConfig[]\",\"name\":\"configs\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"minCreateAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"signatureReward\",\"type\":\"uint256\"}],\"internalType\":\"structXChainTypes.BridgeParams[]\",\"name\":\"params\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getWitnesses\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractGnosisSafeL2\",\"name\":\"safe\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"isTokenRegistered\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pause\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"paused\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proxiableUUID\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"unpause\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"}],\"name\":\"upgradeTo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"upgradeToAndCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
}

// BridgeABI is the input ABI used to generate the binding from.
// Deprecated: Use BridgeMetaData.ABI instead.
var BridgeABI = BridgeMetaData.ABI

// Bridge is an auto generated Go binding around an Ethereum contract.
type Bridge struct {
	BridgeCaller     // Read-only binding to the contract
	BridgeTransactor // Write-only binding to the contract
	BridgeFilterer   // Log filterer for contract events
}

// BridgeCaller is an auto generated read-only Go binding around an Ethereum contract.
type BridgeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BridgeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BridgeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BridgeSession struct {
	Contract     *Bridge           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BridgeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BridgeCallerSession struct {
	Contract *BridgeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// BridgeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BridgeTransactorSession struct {
	Contract     *BridgeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BridgeRaw is an auto generated low-level Go binding around an Ethereum contract.
type BridgeRaw struct {
	Contract *Bridge // Generic contract binding to access the raw methods on
}

// BridgeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BridgeCallerRaw struct {
	Contract *BridgeCaller // Generic read-only contract binding to access the raw methods on
}

// BridgeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BridgeTransactorRaw struct {
	Contract *BridgeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBridge creates a new instance of Bridge, bound to a specific deployed contract.
func NewBridge(address common.Address, backend bind.ContractBackend) (*Bridge, error) {
	contract, err := bindBridge(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Bridge{BridgeCaller: BridgeCaller{contract: contract}, BridgeTransactor: BridgeTransactor{contract: contract}, BridgeFilterer: BridgeFilterer{contract: contract}}, nil
}

// NewBridgeCaller creates a new read-only instance of Bridge, bound to a specific deployed contract.
func NewBridgeCaller(address common.Address, caller bind.ContractCaller) (*BridgeCaller, error) {
	contract, err := bindBridge(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BridgeCaller{contract: contract}, nil
}

// NewBridgeTransactor creates a new write-only instance of Bridge, bound to a specific deployed contract.
func NewBridgeTransactor(address common.Address, transactor bind.ContractTransactor) (*BridgeTransactor, error) {
	contract, err := bindBridge(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BridgeTransactor{contract: contract}, nil
}

// NewBridgeFilterer creates a new log filterer instance of Bridge, bound to a specific deployed contract.
func NewBridgeFilterer(address common.Address, filterer bind.ContractFilterer) (*BridgeFilterer, error) {
	contract, err := bindBridge(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BridgeFilterer{contract: contract}, nil
}

// bindBridge binds a generic wrapper to an already deployed contract.
func bindBridge(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BridgeABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bridge *BridgeRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bridge.Contract.BridgeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bridge *BridgeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.Contract.BridgeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bridge *BridgeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bridge.Contract.BridgeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bridge *BridgeCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bridge.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bridge *BridgeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bridge *BridgeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bridge.Contract.contract.Transact(opts, method, params...)
}

// MINCREATEBRIDGEREWARD is a free data retrieval call binding the contract method 0xd794559a.
//
// Solidity: function MIN_CREATE_BRIDGE_REWARD() view returns(uint256)
func (_Bridge *BridgeCaller) MINCREATEBRIDGEREWARD(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "MIN_CREATE_BRIDGE_REWARD")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MINCREATEBRIDGEREWARD is a free data retrieval call binding the contract method 0xd794559a.
//
// Solidity: function MIN_CREATE_BRIDGE_REWARD() view returns(uint256)
func (_Bridge *BridgeSession) MINCREATEBRIDGEREWARD() (*big.Int, error) {
	return _Bridge.Contract.MINCREATEBRIDGEREWARD(&_Bridge.CallOpts)
}

// MINCREATEBRIDGEREWARD is a free data retrieval call binding the contract method 0xd794559a.
//
// Solidity: function MIN_CREATE_BRIDGE_REWARD() view returns(uint256)
func (_Bridge *BridgeCallerSession) MINCREATEBRIDGEREWARD() (*big.Int, error) {
	return _Bridge.Contract.MINCREATEBRIDGEREWARD(&_Bridge.CallOpts)
}

// Safe is a free data retrieval call binding the contract method 0x2a983d3a.
//
// Solidity: function _safe() view returns(address)
func (_Bridge *BridgeCaller) Safe(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "_safe")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Safe is a free data retrieval call binding the contract method 0x2a983d3a.
//
// Solidity: function _safe() view returns(address)
func (_Bridge *BridgeSession) Safe() (common.Address, error) {
	return _Bridge.Contract.Safe(&_Bridge.CallOpts)
}

// Safe is a free data retrieval call binding the contract method 0x2a983d3a.
//
// Solidity: function _safe() view returns(address)
func (_Bridge *BridgeCallerSession) Safe() (common.Address, error) {
	return _Bridge.Contract.Safe(&_Bridge.CallOpts)
}

// GetBridgeClaim is a free data retrieval call binding the contract method 0xc45ddbc2.
//
// Solidity: function getBridgeClaim((address,(address,string),address,(address,string)) bridgeConfig, uint256 claimId) view returns(address, address, bool)
func (_Bridge *BridgeCaller) GetBridgeClaim(opts *bind.CallOpts, bridgeConfig XChainTypesBridgeConfig, claimId *big.Int) (common.Address, common.Address, bool, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "getBridgeClaim", bridgeConfig, claimId)

	if err != nil {
		return *new(common.Address), *new(common.Address), *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	out1 := *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	out2 := *abi.ConvertType(out[2], new(bool)).(*bool)

	return out0, out1, out2, err

}

// GetBridgeClaim is a free data retrieval call binding the contract method 0xc45ddbc2.
//
// Solidity: function getBridgeClaim((address,(address,string),address,(address,string)) bridgeConfig, uint256 claimId) view returns(address, address, bool)
func (_Bridge *BridgeSession) GetBridgeClaim(bridgeConfig XChainTypesBridgeConfig, claimId *big.Int) (common.Address, common.Address, bool, error) {
	return _Bridge.Contract.GetBridgeClaim(&_Bridge.CallOpts, bridgeConfig, claimId)
}

// GetBridgeClaim is a free data retrieval call binding the contract method 0xc45ddbc2.
//
// Solidity: function getBridgeClaim((address,(address,string),address,(address,string)) bridgeConfig, uint256 claimId) view returns(address, address, bool)
func (_Bridge *BridgeCallerSession) GetBridgeClaim(bridgeConfig XChainTypesBridgeConfig, claimId *big.Int) (common.Address, common.Address, bool, error) {
	return _Bridge.Contract.GetBridgeClaim(&_Bridge.CallOpts, bridgeConfig, claimId)
}

// GetBridgeConfig is a free data retrieval call binding the contract method 0x0b2c50d2.
//
// Solidity: function getBridgeConfig((address,(address,string),address,(address,string)) bridgeConfig) view returns(address, address, string, address, address, string)
func (_Bridge *BridgeCaller) GetBridgeConfig(opts *bind.CallOpts, bridgeConfig XChainTypesBridgeConfig) (common.Address, common.Address, string, common.Address, common.Address, string, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "getBridgeConfig", bridgeConfig)

	if err != nil {
		return *new(common.Address), *new(common.Address), *new(string), *new(common.Address), *new(common.Address), *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	out1 := *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	out2 := *abi.ConvertType(out[2], new(string)).(*string)
	out3 := *abi.ConvertType(out[3], new(common.Address)).(*common.Address)
	out4 := *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	out5 := *abi.ConvertType(out[5], new(string)).(*string)

	return out0, out1, out2, out3, out4, out5, err

}

// GetBridgeConfig is a free data retrieval call binding the contract method 0x0b2c50d2.
//
// Solidity: function getBridgeConfig((address,(address,string),address,(address,string)) bridgeConfig) view returns(address, address, string, address, address, string)
func (_Bridge *BridgeSession) GetBridgeConfig(bridgeConfig XChainTypesBridgeConfig) (common.Address, common.Address, string, common.Address, common.Address, string, error) {
	return _Bridge.Contract.GetBridgeConfig(&_Bridge.CallOpts, bridgeConfig)
}

// GetBridgeConfig is a free data retrieval call binding the contract method 0x0b2c50d2.
//
// Solidity: function getBridgeConfig((address,(address,string),address,(address,string)) bridgeConfig) view returns(address, address, string, address, address, string)
func (_Bridge *BridgeCallerSession) GetBridgeConfig(bridgeConfig XChainTypesBridgeConfig) (common.Address, common.Address, string, common.Address, common.Address, string, error) {
	return _Bridge.Contract.GetBridgeConfig(&_Bridge.CallOpts, bridgeConfig)
}

// GetBridgeCreateAccount is a free data retrieval call binding the contract method 0x2d92de74.
//
// Solidity: function getBridgeCreateAccount((address,(address,string),address,(address,string)) bridgeConfig, address account) view returns(uint256, bool, bool)
func (_Bridge *BridgeCaller) GetBridgeCreateAccount(opts *bind.CallOpts, bridgeConfig XChainTypesBridgeConfig, account common.Address) (*big.Int, bool, bool, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "getBridgeCreateAccount", bridgeConfig, account)

	if err != nil {
		return *new(*big.Int), *new(bool), *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	out1 := *abi.ConvertType(out[1], new(bool)).(*bool)
	out2 := *abi.ConvertType(out[2], new(bool)).(*bool)

	return out0, out1, out2, err

}

// GetBridgeCreateAccount is a free data retrieval call binding the contract method 0x2d92de74.
//
// Solidity: function getBridgeCreateAccount((address,(address,string),address,(address,string)) bridgeConfig, address account) view returns(uint256, bool, bool)
func (_Bridge *BridgeSession) GetBridgeCreateAccount(bridgeConfig XChainTypesBridgeConfig, account common.Address) (*big.Int, bool, bool, error) {
	return _Bridge.Contract.GetBridgeCreateAccount(&_Bridge.CallOpts, bridgeConfig, account)
}

// GetBridgeCreateAccount is a free data retrieval call binding the contract method 0x2d92de74.
//
// Solidity: function getBridgeCreateAccount((address,(address,string),address,(address,string)) bridgeConfig, address account) view returns(uint256, bool, bool)
func (_Bridge *BridgeCallerSession) GetBridgeCreateAccount(bridgeConfig XChainTypesBridgeConfig, account common.Address) (*big.Int, bool, bool, error) {
	return _Bridge.Contract.GetBridgeCreateAccount(&_Bridge.CallOpts, bridgeConfig, account)
}

// GetBridgeKey is a free data retrieval call binding the contract method 0x5bafd0c1.
//
// Solidity: function getBridgeKey((address,(address,string),address,(address,string)) bridgeConfig) pure returns(bytes32)
func (_Bridge *BridgeCaller) GetBridgeKey(opts *bind.CallOpts, bridgeConfig XChainTypesBridgeConfig) ([32]byte, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "getBridgeKey", bridgeConfig)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetBridgeKey is a free data retrieval call binding the contract method 0x5bafd0c1.
//
// Solidity: function getBridgeKey((address,(address,string),address,(address,string)) bridgeConfig) pure returns(bytes32)
func (_Bridge *BridgeSession) GetBridgeKey(bridgeConfig XChainTypesBridgeConfig) ([32]byte, error) {
	return _Bridge.Contract.GetBridgeKey(&_Bridge.CallOpts, bridgeConfig)
}

// GetBridgeKey is a free data retrieval call binding the contract method 0x5bafd0c1.
//
// Solidity: function getBridgeKey((address,(address,string),address,(address,string)) bridgeConfig) pure returns(bytes32)
func (_Bridge *BridgeCallerSession) GetBridgeKey(bridgeConfig XChainTypesBridgeConfig) ([32]byte, error) {
	return _Bridge.Contract.GetBridgeKey(&_Bridge.CallOpts, bridgeConfig)
}

// GetBridgeParams is a free data retrieval call binding the contract method 0x72e0376f.
//
// Solidity: function getBridgeParams((address,(address,string),address,(address,string)) bridgeConfig) view returns(uint256, uint256)
func (_Bridge *BridgeCaller) GetBridgeParams(opts *bind.CallOpts, bridgeConfig XChainTypesBridgeConfig) (*big.Int, *big.Int, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "getBridgeParams", bridgeConfig)

	if err != nil {
		return *new(*big.Int), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return out0, out1, err

}

// GetBridgeParams is a free data retrieval call binding the contract method 0x72e0376f.
//
// Solidity: function getBridgeParams((address,(address,string),address,(address,string)) bridgeConfig) view returns(uint256, uint256)
func (_Bridge *BridgeSession) GetBridgeParams(bridgeConfig XChainTypesBridgeConfig) (*big.Int, *big.Int, error) {
	return _Bridge.Contract.GetBridgeParams(&_Bridge.CallOpts, bridgeConfig)
}

// GetBridgeParams is a free data retrieval call binding the contract method 0x72e0376f.
//
// Solidity: function getBridgeParams((address,(address,string),address,(address,string)) bridgeConfig) view returns(uint256, uint256)
func (_Bridge *BridgeCallerSession) GetBridgeParams(bridgeConfig XChainTypesBridgeConfig) (*big.Int, *big.Int, error) {
	return _Bridge.Contract.GetBridgeParams(&_Bridge.CallOpts, bridgeConfig)
}

// GetBridgeToken is a free data retrieval call binding the contract method 0xa6c0873b.
//
// Solidity: function getBridgeToken((address,(address,string),address,(address,string)) bridgeConfig) view returns(address)
func (_Bridge *BridgeCaller) GetBridgeToken(opts *bind.CallOpts, bridgeConfig XChainTypesBridgeConfig) (common.Address, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "getBridgeToken", bridgeConfig)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetBridgeToken is a free data retrieval call binding the contract method 0xa6c0873b.
//
// Solidity: function getBridgeToken((address,(address,string),address,(address,string)) bridgeConfig) view returns(address)
func (_Bridge *BridgeSession) GetBridgeToken(bridgeConfig XChainTypesBridgeConfig) (common.Address, error) {
	return _Bridge.Contract.GetBridgeToken(&_Bridge.CallOpts, bridgeConfig)
}

// GetBridgeToken is a free data retrieval call binding the contract method 0xa6c0873b.
//
// Solidity: function getBridgeToken((address,(address,string),address,(address,string)) bridgeConfig) view returns(address)
func (_Bridge *BridgeCallerSession) GetBridgeToken(bridgeConfig XChainTypesBridgeConfig) (common.Address, error) {
	return _Bridge.Contract.GetBridgeToken(&_Bridge.CallOpts, bridgeConfig)
}

// GetBridgesPaginated is a free data retrieval call binding the contract method 0x34525b1b.
//
// Solidity: function getBridgesPaginated(uint256 page) view returns((address,(address,string),address,(address,string))[] configs, (uint256,uint256)[] params)
func (_Bridge *BridgeCaller) GetBridgesPaginated(opts *bind.CallOpts, page *big.Int) (struct {
	Configs []XChainTypesBridgeConfig
	Params  []XChainTypesBridgeParams
}, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "getBridgesPaginated", page)

	outstruct := new(struct {
		Configs []XChainTypesBridgeConfig
		Params  []XChainTypesBridgeParams
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Configs = *abi.ConvertType(out[0], new([]XChainTypesBridgeConfig)).(*[]XChainTypesBridgeConfig)
	outstruct.Params = *abi.ConvertType(out[1], new([]XChainTypesBridgeParams)).(*[]XChainTypesBridgeParams)

	return *outstruct, err

}

// GetBridgesPaginated is a free data retrieval call binding the contract method 0x34525b1b.
//
// Solidity: function getBridgesPaginated(uint256 page) view returns((address,(address,string),address,(address,string))[] configs, (uint256,uint256)[] params)
func (_Bridge *BridgeSession) GetBridgesPaginated(page *big.Int) (struct {
	Configs []XChainTypesBridgeConfig
	Params  []XChainTypesBridgeParams
}, error) {
	return _Bridge.Contract.GetBridgesPaginated(&_Bridge.CallOpts, page)
}

// GetBridgesPaginated is a free data retrieval call binding the contract method 0x34525b1b.
//
// Solidity: function getBridgesPaginated(uint256 page) view returns((address,(address,string),address,(address,string))[] configs, (uint256,uint256)[] params)
func (_Bridge *BridgeCallerSession) GetBridgesPaginated(page *big.Int) (struct {
	Configs []XChainTypesBridgeConfig
	Params  []XChainTypesBridgeParams
}, error) {
	return _Bridge.Contract.GetBridgesPaginated(&_Bridge.CallOpts, page)
}

// GetWitnesses is a free data retrieval call binding the contract method 0x96d195bd.
//
// Solidity: function getWitnesses() view returns(address[])
func (_Bridge *BridgeCaller) GetWitnesses(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "getWitnesses")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetWitnesses is a free data retrieval call binding the contract method 0x96d195bd.
//
// Solidity: function getWitnesses() view returns(address[])
func (_Bridge *BridgeSession) GetWitnesses() ([]common.Address, error) {
	return _Bridge.Contract.GetWitnesses(&_Bridge.CallOpts)
}

// GetWitnesses is a free data retrieval call binding the contract method 0x96d195bd.
//
// Solidity: function getWitnesses() view returns(address[])
func (_Bridge *BridgeCallerSession) GetWitnesses() ([]common.Address, error) {
	return _Bridge.Contract.GetWitnesses(&_Bridge.CallOpts)
}

// IsTokenRegistered is a free data retrieval call binding the contract method 0x26aa101f.
//
// Solidity: function isTokenRegistered(address token) view returns(bool)
func (_Bridge *BridgeCaller) IsTokenRegistered(opts *bind.CallOpts, token common.Address) (bool, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "isTokenRegistered", token)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsTokenRegistered is a free data retrieval call binding the contract method 0x26aa101f.
//
// Solidity: function isTokenRegistered(address token) view returns(bool)
func (_Bridge *BridgeSession) IsTokenRegistered(token common.Address) (bool, error) {
	return _Bridge.Contract.IsTokenRegistered(&_Bridge.CallOpts, token)
}

// IsTokenRegistered is a free data retrieval call binding the contract method 0x26aa101f.
//
// Solidity: function isTokenRegistered(address token) view returns(bool)
func (_Bridge *BridgeCallerSession) IsTokenRegistered(token common.Address) (bool, error) {
	return _Bridge.Contract.IsTokenRegistered(&_Bridge.CallOpts, token)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bridge *BridgeCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bridge *BridgeSession) Owner() (common.Address, error) {
	return _Bridge.Contract.Owner(&_Bridge.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bridge *BridgeCallerSession) Owner() (common.Address, error) {
	return _Bridge.Contract.Owner(&_Bridge.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Bridge *BridgeCaller) Paused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "paused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Bridge *BridgeSession) Paused() (bool, error) {
	return _Bridge.Contract.Paused(&_Bridge.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Bridge *BridgeCallerSession) Paused() (bool, error) {
	return _Bridge.Contract.Paused(&_Bridge.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_Bridge *BridgeCaller) ProxiableUUID(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "proxiableUUID")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_Bridge *BridgeSession) ProxiableUUID() ([32]byte, error) {
	return _Bridge.Contract.ProxiableUUID(&_Bridge.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_Bridge *BridgeCallerSession) ProxiableUUID() ([32]byte, error) {
	return _Bridge.Contract.ProxiableUUID(&_Bridge.CallOpts)
}

// BridgeDoorCommonInit is a paid mutator transaction binding the contract method 0xe684ab76.
//
// Solidity: function __BridgeDoorCommon_init(address safe) returns()
func (_Bridge *BridgeTransactor) BridgeDoorCommonInit(opts *bind.TransactOpts, safe common.Address) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "__BridgeDoorCommon_init", safe)
}

// BridgeDoorCommonInit is a paid mutator transaction binding the contract method 0xe684ab76.
//
// Solidity: function __BridgeDoorCommon_init(address safe) returns()
func (_Bridge *BridgeSession) BridgeDoorCommonInit(safe common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.BridgeDoorCommonInit(&_Bridge.TransactOpts, safe)
}

// BridgeDoorCommonInit is a paid mutator transaction binding the contract method 0xe684ab76.
//
// Solidity: function __BridgeDoorCommon_init(address safe) returns()
func (_Bridge *BridgeTransactorSession) BridgeDoorCommonInit(safe common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.BridgeDoorCommonInit(&_Bridge.TransactOpts, safe)
}

// ManageableInit is a paid mutator transaction binding the contract method 0x4072c3b3.
//
// Solidity: function __Manageable_init() returns()
func (_Bridge *BridgeTransactor) ManageableInit(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "__Manageable_init")
}

// ManageableInit is a paid mutator transaction binding the contract method 0x4072c3b3.
//
// Solidity: function __Manageable_init() returns()
func (_Bridge *BridgeSession) ManageableInit() (*types.Transaction, error) {
	return _Bridge.Contract.ManageableInit(&_Bridge.TransactOpts)
}

// ManageableInit is a paid mutator transaction binding the contract method 0x4072c3b3.
//
// Solidity: function __Manageable_init() returns()
func (_Bridge *BridgeTransactorSession) ManageableInit() (*types.Transaction, error) {
	return _Bridge.Contract.ManageableInit(&_Bridge.TransactOpts)
}

// AddClaimAttestation is a paid mutator transaction binding the contract method 0x192dd3cc.
//
// Solidity: function addClaimAttestation((address,(address,string),address,(address,string)) bridgeConfig, uint256 claimId, uint256 amount, address sender, address destination) returns()
func (_Bridge *BridgeTransactor) AddClaimAttestation(opts *bind.TransactOpts, bridgeConfig XChainTypesBridgeConfig, claimId *big.Int, amount *big.Int, sender common.Address, destination common.Address) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "addClaimAttestation", bridgeConfig, claimId, amount, sender, destination)
}

// AddClaimAttestation is a paid mutator transaction binding the contract method 0x192dd3cc.
//
// Solidity: function addClaimAttestation((address,(address,string),address,(address,string)) bridgeConfig, uint256 claimId, uint256 amount, address sender, address destination) returns()
func (_Bridge *BridgeSession) AddClaimAttestation(bridgeConfig XChainTypesBridgeConfig, claimId *big.Int, amount *big.Int, sender common.Address, destination common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.AddClaimAttestation(&_Bridge.TransactOpts, bridgeConfig, claimId, amount, sender, destination)
}

// AddClaimAttestation is a paid mutator transaction binding the contract method 0x192dd3cc.
//
// Solidity: function addClaimAttestation((address,(address,string),address,(address,string)) bridgeConfig, uint256 claimId, uint256 amount, address sender, address destination) returns()
func (_Bridge *BridgeTransactorSession) AddClaimAttestation(bridgeConfig XChainTypesBridgeConfig, claimId *big.Int, amount *big.Int, sender common.Address, destination common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.AddClaimAttestation(&_Bridge.TransactOpts, bridgeConfig, claimId, amount, sender, destination)
}

// AddCreateAccountAttestation is a paid mutator transaction binding the contract method 0x95fa18dd.
//
// Solidity: function addCreateAccountAttestation((address,(address,string),address,(address,string)) bridgeConfig, address destination, uint256 amount, uint256 signatureReward) returns()
func (_Bridge *BridgeTransactor) AddCreateAccountAttestation(opts *bind.TransactOpts, bridgeConfig XChainTypesBridgeConfig, destination common.Address, amount *big.Int, signatureReward *big.Int) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "addCreateAccountAttestation", bridgeConfig, destination, amount, signatureReward)
}

// AddCreateAccountAttestation is a paid mutator transaction binding the contract method 0x95fa18dd.
//
// Solidity: function addCreateAccountAttestation((address,(address,string),address,(address,string)) bridgeConfig, address destination, uint256 amount, uint256 signatureReward) returns()
func (_Bridge *BridgeSession) AddCreateAccountAttestation(bridgeConfig XChainTypesBridgeConfig, destination common.Address, amount *big.Int, signatureReward *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.AddCreateAccountAttestation(&_Bridge.TransactOpts, bridgeConfig, destination, amount, signatureReward)
}

// AddCreateAccountAttestation is a paid mutator transaction binding the contract method 0x95fa18dd.
//
// Solidity: function addCreateAccountAttestation((address,(address,string),address,(address,string)) bridgeConfig, address destination, uint256 amount, uint256 signatureReward) returns()
func (_Bridge *BridgeTransactorSession) AddCreateAccountAttestation(bridgeConfig XChainTypesBridgeConfig, destination common.Address, amount *big.Int, signatureReward *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.AddCreateAccountAttestation(&_Bridge.TransactOpts, bridgeConfig, destination, amount, signatureReward)
}

// Claim is a paid mutator transaction binding the contract method 0x8436c642.
//
// Solidity: function claim((address,(address,string),address,(address,string)) bridgeConfig, uint256 claimId, uint256 amount, address destination) returns()
func (_Bridge *BridgeTransactor) Claim(opts *bind.TransactOpts, bridgeConfig XChainTypesBridgeConfig, claimId *big.Int, amount *big.Int, destination common.Address) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "claim", bridgeConfig, claimId, amount, destination)
}

// Claim is a paid mutator transaction binding the contract method 0x8436c642.
//
// Solidity: function claim((address,(address,string),address,(address,string)) bridgeConfig, uint256 claimId, uint256 amount, address destination) returns()
func (_Bridge *BridgeSession) Claim(bridgeConfig XChainTypesBridgeConfig, claimId *big.Int, amount *big.Int, destination common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.Claim(&_Bridge.TransactOpts, bridgeConfig, claimId, amount, destination)
}

// Claim is a paid mutator transaction binding the contract method 0x8436c642.
//
// Solidity: function claim((address,(address,string),address,(address,string)) bridgeConfig, uint256 claimId, uint256 amount, address destination) returns()
func (_Bridge *BridgeTransactorSession) Claim(bridgeConfig XChainTypesBridgeConfig, claimId *big.Int, amount *big.Int, destination common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.Claim(&_Bridge.TransactOpts, bridgeConfig, claimId, amount, destination)
}

// Commit is a paid mutator transaction binding the contract method 0x75f1d504.
//
// Solidity: function commit((address,(address,string),address,(address,string)) bridgeConfig, address receiver, uint256 claimId, uint256 amount) payable returns()
func (_Bridge *BridgeTransactor) Commit(opts *bind.TransactOpts, bridgeConfig XChainTypesBridgeConfig, receiver common.Address, claimId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "commit", bridgeConfig, receiver, claimId, amount)
}

// Commit is a paid mutator transaction binding the contract method 0x75f1d504.
//
// Solidity: function commit((address,(address,string),address,(address,string)) bridgeConfig, address receiver, uint256 claimId, uint256 amount) payable returns()
func (_Bridge *BridgeSession) Commit(bridgeConfig XChainTypesBridgeConfig, receiver common.Address, claimId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.Commit(&_Bridge.TransactOpts, bridgeConfig, receiver, claimId, amount)
}

// Commit is a paid mutator transaction binding the contract method 0x75f1d504.
//
// Solidity: function commit((address,(address,string),address,(address,string)) bridgeConfig, address receiver, uint256 claimId, uint256 amount) payable returns()
func (_Bridge *BridgeTransactorSession) Commit(bridgeConfig XChainTypesBridgeConfig, receiver common.Address, claimId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.Commit(&_Bridge.TransactOpts, bridgeConfig, receiver, claimId, amount)
}

// CommitWithoutAddress is a paid mutator transaction binding the contract method 0x2ca4cb53.
//
// Solidity: function commitWithoutAddress((address,(address,string),address,(address,string)) bridgeConfig, uint256 claimId, uint256 amount) payable returns()
func (_Bridge *BridgeTransactor) CommitWithoutAddress(opts *bind.TransactOpts, bridgeConfig XChainTypesBridgeConfig, claimId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "commitWithoutAddress", bridgeConfig, claimId, amount)
}

// CommitWithoutAddress is a paid mutator transaction binding the contract method 0x2ca4cb53.
//
// Solidity: function commitWithoutAddress((address,(address,string),address,(address,string)) bridgeConfig, uint256 claimId, uint256 amount) payable returns()
func (_Bridge *BridgeSession) CommitWithoutAddress(bridgeConfig XChainTypesBridgeConfig, claimId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.CommitWithoutAddress(&_Bridge.TransactOpts, bridgeConfig, claimId, amount)
}

// CommitWithoutAddress is a paid mutator transaction binding the contract method 0x2ca4cb53.
//
// Solidity: function commitWithoutAddress((address,(address,string),address,(address,string)) bridgeConfig, uint256 claimId, uint256 amount) payable returns()
func (_Bridge *BridgeTransactorSession) CommitWithoutAddress(bridgeConfig XChainTypesBridgeConfig, claimId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.CommitWithoutAddress(&_Bridge.TransactOpts, bridgeConfig, claimId, amount)
}

// CreateAccountCommit is a paid mutator transaction binding the contract method 0x3f2702e4.
//
// Solidity: function createAccountCommit((address,(address,string),address,(address,string)) bridgeConfig, address destination, uint256 amount, uint256 signatureReward) payable returns()
func (_Bridge *BridgeTransactor) CreateAccountCommit(opts *bind.TransactOpts, bridgeConfig XChainTypesBridgeConfig, destination common.Address, amount *big.Int, signatureReward *big.Int) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "createAccountCommit", bridgeConfig, destination, amount, signatureReward)
}

// CreateAccountCommit is a paid mutator transaction binding the contract method 0x3f2702e4.
//
// Solidity: function createAccountCommit((address,(address,string),address,(address,string)) bridgeConfig, address destination, uint256 amount, uint256 signatureReward) payable returns()
func (_Bridge *BridgeSession) CreateAccountCommit(bridgeConfig XChainTypesBridgeConfig, destination common.Address, amount *big.Int, signatureReward *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.CreateAccountCommit(&_Bridge.TransactOpts, bridgeConfig, destination, amount, signatureReward)
}

// CreateAccountCommit is a paid mutator transaction binding the contract method 0x3f2702e4.
//
// Solidity: function createAccountCommit((address,(address,string),address,(address,string)) bridgeConfig, address destination, uint256 amount, uint256 signatureReward) payable returns()
func (_Bridge *BridgeTransactorSession) CreateAccountCommit(bridgeConfig XChainTypesBridgeConfig, destination common.Address, amount *big.Int, signatureReward *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.CreateAccountCommit(&_Bridge.TransactOpts, bridgeConfig, destination, amount, signatureReward)
}

// CreateBridge is a paid mutator transaction binding the contract method 0x4a07d673.
//
// Solidity: function createBridge((address,(address,string),address,(address,string)) config, (uint256,uint256) params) returns()
func (_Bridge *BridgeTransactor) CreateBridge(opts *bind.TransactOpts, config XChainTypesBridgeConfig, params XChainTypesBridgeParams) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "createBridge", config, params)
}

// CreateBridge is a paid mutator transaction binding the contract method 0x4a07d673.
//
// Solidity: function createBridge((address,(address,string),address,(address,string)) config, (uint256,uint256) params) returns()
func (_Bridge *BridgeSession) CreateBridge(config XChainTypesBridgeConfig, params XChainTypesBridgeParams) (*types.Transaction, error) {
	return _Bridge.Contract.CreateBridge(&_Bridge.TransactOpts, config, params)
}

// CreateBridge is a paid mutator transaction binding the contract method 0x4a07d673.
//
// Solidity: function createBridge((address,(address,string),address,(address,string)) config, (uint256,uint256) params) returns()
func (_Bridge *BridgeTransactorSession) CreateBridge(config XChainTypesBridgeConfig, params XChainTypesBridgeParams) (*types.Transaction, error) {
	return _Bridge.Contract.CreateBridge(&_Bridge.TransactOpts, config, params)
}

// CreateBridgeRequest is a paid mutator transaction binding the contract method 0x10231036.
//
// Solidity: function createBridgeRequest(address tokenAddress) payable returns()
func (_Bridge *BridgeTransactor) CreateBridgeRequest(opts *bind.TransactOpts, tokenAddress common.Address) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "createBridgeRequest", tokenAddress)
}

// CreateBridgeRequest is a paid mutator transaction binding the contract method 0x10231036.
//
// Solidity: function createBridgeRequest(address tokenAddress) payable returns()
func (_Bridge *BridgeSession) CreateBridgeRequest(tokenAddress common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.CreateBridgeRequest(&_Bridge.TransactOpts, tokenAddress)
}

// CreateBridgeRequest is a paid mutator transaction binding the contract method 0x10231036.
//
// Solidity: function createBridgeRequest(address tokenAddress) payable returns()
func (_Bridge *BridgeTransactorSession) CreateBridgeRequest(tokenAddress common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.CreateBridgeRequest(&_Bridge.TransactOpts, tokenAddress)
}

// CreateClaimId is a paid mutator transaction binding the contract method 0x8d5cd5bd.
//
// Solidity: function createClaimId((address,(address,string),address,(address,string)) bridgeConfig, address sender) payable returns(uint256)
func (_Bridge *BridgeTransactor) CreateClaimId(opts *bind.TransactOpts, bridgeConfig XChainTypesBridgeConfig, sender common.Address) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "createClaimId", bridgeConfig, sender)
}

// CreateClaimId is a paid mutator transaction binding the contract method 0x8d5cd5bd.
//
// Solidity: function createClaimId((address,(address,string),address,(address,string)) bridgeConfig, address sender) payable returns(uint256)
func (_Bridge *BridgeSession) CreateClaimId(bridgeConfig XChainTypesBridgeConfig, sender common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.CreateClaimId(&_Bridge.TransactOpts, bridgeConfig, sender)
}

// CreateClaimId is a paid mutator transaction binding the contract method 0x8d5cd5bd.
//
// Solidity: function createClaimId((address,(address,string),address,(address,string)) bridgeConfig, address sender) payable returns(uint256)
func (_Bridge *BridgeTransactorSession) CreateClaimId(bridgeConfig XChainTypesBridgeConfig, sender common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.CreateClaimId(&_Bridge.TransactOpts, bridgeConfig, sender)
}

// Execute is a paid mutator transaction binding the contract method 0x51945447.
//
// Solidity: function execute(address to, uint256 value, bytes data, uint8 operation) returns(bool success)
func (_Bridge *BridgeTransactor) Execute(opts *bind.TransactOpts, to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "execute", to, value, data, operation)
}

// Execute is a paid mutator transaction binding the contract method 0x51945447.
//
// Solidity: function execute(address to, uint256 value, bytes data, uint8 operation) returns(bool success)
func (_Bridge *BridgeSession) Execute(to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _Bridge.Contract.Execute(&_Bridge.TransactOpts, to, value, data, operation)
}

// Execute is a paid mutator transaction binding the contract method 0x51945447.
//
// Solidity: function execute(address to, uint256 value, bytes data, uint8 operation) returns(bool success)
func (_Bridge *BridgeTransactorSession) Execute(to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _Bridge.Contract.Execute(&_Bridge.TransactOpts, to, value, data, operation)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address safe) returns()
func (_Bridge *BridgeTransactor) Initialize(opts *bind.TransactOpts, safe common.Address) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "initialize", safe)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address safe) returns()
func (_Bridge *BridgeSession) Initialize(safe common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.Initialize(&_Bridge.TransactOpts, safe)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address safe) returns()
func (_Bridge *BridgeTransactorSession) Initialize(safe common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.Initialize(&_Bridge.TransactOpts, safe)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_Bridge *BridgeTransactor) Pause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "pause")
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_Bridge *BridgeSession) Pause() (*types.Transaction, error) {
	return _Bridge.Contract.Pause(&_Bridge.TransactOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_Bridge *BridgeTransactorSession) Pause() (*types.Transaction, error) {
	return _Bridge.Contract.Pause(&_Bridge.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bridge *BridgeTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bridge *BridgeSession) RenounceOwnership() (*types.Transaction, error) {
	return _Bridge.Contract.RenounceOwnership(&_Bridge.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bridge *BridgeTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Bridge.Contract.RenounceOwnership(&_Bridge.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bridge *BridgeTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bridge *BridgeSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.TransferOwnership(&_Bridge.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bridge *BridgeTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.TransferOwnership(&_Bridge.TransactOpts, newOwner)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_Bridge *BridgeTransactor) Unpause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "unpause")
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_Bridge *BridgeSession) Unpause() (*types.Transaction, error) {
	return _Bridge.Contract.Unpause(&_Bridge.TransactOpts)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_Bridge *BridgeTransactorSession) Unpause() (*types.Transaction, error) {
	return _Bridge.Contract.Unpause(&_Bridge.TransactOpts)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_Bridge *BridgeTransactor) UpgradeTo(opts *bind.TransactOpts, newImplementation common.Address) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "upgradeTo", newImplementation)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_Bridge *BridgeSession) UpgradeTo(newImplementation common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.UpgradeTo(&_Bridge.TransactOpts, newImplementation)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_Bridge *BridgeTransactorSession) UpgradeTo(newImplementation common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.UpgradeTo(&_Bridge.TransactOpts, newImplementation)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Bridge *BridgeTransactor) UpgradeToAndCall(opts *bind.TransactOpts, newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "upgradeToAndCall", newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Bridge *BridgeSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Bridge.Contract.UpgradeToAndCall(&_Bridge.TransactOpts, newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Bridge *BridgeTransactorSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Bridge.Contract.UpgradeToAndCall(&_Bridge.TransactOpts, newImplementation, data)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Bridge *BridgeTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Bridge *BridgeSession) Receive() (*types.Transaction, error) {
	return _Bridge.Contract.Receive(&_Bridge.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Bridge *BridgeTransactorSession) Receive() (*types.Transaction, error) {
	return _Bridge.Contract.Receive(&_Bridge.TransactOpts)
}

// BridgeAddClaimAttestationIterator is returned from FilterAddClaimAttestation and is used to iterate over the raw logs and unpacked data for AddClaimAttestation events raised by the Bridge contract.
type BridgeAddClaimAttestationIterator struct {
	Event *BridgeAddClaimAttestation // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeAddClaimAttestationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeAddClaimAttestation)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeAddClaimAttestation)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeAddClaimAttestationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeAddClaimAttestationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeAddClaimAttestation represents a AddClaimAttestation event raised by the Bridge contract.
type BridgeAddClaimAttestation struct {
	BridgeKey [32]byte
	ClaimId   *big.Int
	Witness   common.Address
	Value     *big.Int
	Receiver  common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterAddClaimAttestation is a free log retrieval operation binding the contract event 0x750bc14bd61ac34397f57970e1a0fd14fd27247cc77aa2b3c304e2e973c52a40.
//
// Solidity: event AddClaimAttestation(bytes32 indexed bridgeKey, uint256 indexed claimId, address indexed witness, uint256 value, address receiver)
func (_Bridge *BridgeFilterer) FilterAddClaimAttestation(opts *bind.FilterOpts, bridgeKey [][32]byte, claimId []*big.Int, witness []common.Address) (*BridgeAddClaimAttestationIterator, error) {

	var bridgeKeyRule []interface{}
	for _, bridgeKeyItem := range bridgeKey {
		bridgeKeyRule = append(bridgeKeyRule, bridgeKeyItem)
	}
	var claimIdRule []interface{}
	for _, claimIdItem := range claimId {
		claimIdRule = append(claimIdRule, claimIdItem)
	}
	var witnessRule []interface{}
	for _, witnessItem := range witness {
		witnessRule = append(witnessRule, witnessItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "AddClaimAttestation", bridgeKeyRule, claimIdRule, witnessRule)
	if err != nil {
		return nil, err
	}
	return &BridgeAddClaimAttestationIterator{contract: _Bridge.contract, event: "AddClaimAttestation", logs: logs, sub: sub}, nil
}

// WatchAddClaimAttestation is a free log subscription operation binding the contract event 0x750bc14bd61ac34397f57970e1a0fd14fd27247cc77aa2b3c304e2e973c52a40.
//
// Solidity: event AddClaimAttestation(bytes32 indexed bridgeKey, uint256 indexed claimId, address indexed witness, uint256 value, address receiver)
func (_Bridge *BridgeFilterer) WatchAddClaimAttestation(opts *bind.WatchOpts, sink chan<- *BridgeAddClaimAttestation, bridgeKey [][32]byte, claimId []*big.Int, witness []common.Address) (event.Subscription, error) {

	var bridgeKeyRule []interface{}
	for _, bridgeKeyItem := range bridgeKey {
		bridgeKeyRule = append(bridgeKeyRule, bridgeKeyItem)
	}
	var claimIdRule []interface{}
	for _, claimIdItem := range claimId {
		claimIdRule = append(claimIdRule, claimIdItem)
	}
	var witnessRule []interface{}
	for _, witnessItem := range witness {
		witnessRule = append(witnessRule, witnessItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "AddClaimAttestation", bridgeKeyRule, claimIdRule, witnessRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeAddClaimAttestation)
				if err := _Bridge.contract.UnpackLog(event, "AddClaimAttestation", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAddClaimAttestation is a log parse operation binding the contract event 0x750bc14bd61ac34397f57970e1a0fd14fd27247cc77aa2b3c304e2e973c52a40.
//
// Solidity: event AddClaimAttestation(bytes32 indexed bridgeKey, uint256 indexed claimId, address indexed witness, uint256 value, address receiver)
func (_Bridge *BridgeFilterer) ParseAddClaimAttestation(log types.Log) (*BridgeAddClaimAttestation, error) {
	event := new(BridgeAddClaimAttestation)
	if err := _Bridge.contract.UnpackLog(event, "AddClaimAttestation", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeAddCreateAccountAttestationIterator is returned from FilterAddCreateAccountAttestation and is used to iterate over the raw logs and unpacked data for AddCreateAccountAttestation events raised by the Bridge contract.
type BridgeAddCreateAccountAttestationIterator struct {
	Event *BridgeAddCreateAccountAttestation // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeAddCreateAccountAttestationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeAddCreateAccountAttestation)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeAddCreateAccountAttestation)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeAddCreateAccountAttestationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeAddCreateAccountAttestationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeAddCreateAccountAttestation represents a AddCreateAccountAttestation event raised by the Bridge contract.
type BridgeAddCreateAccountAttestation struct {
	BridgeKey [32]byte
	Witness   common.Address
	Receiver  common.Address
	Value     *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterAddCreateAccountAttestation is a free log retrieval operation binding the contract event 0x64f8fc141a8c2d310942bbc5236041fa0a0c6c39a9fb295be5c959224382815e.
//
// Solidity: event AddCreateAccountAttestation(bytes32 indexed bridgeKey, address indexed witness, address indexed receiver, uint256 value)
func (_Bridge *BridgeFilterer) FilterAddCreateAccountAttestation(opts *bind.FilterOpts, bridgeKey [][32]byte, witness []common.Address, receiver []common.Address) (*BridgeAddCreateAccountAttestationIterator, error) {

	var bridgeKeyRule []interface{}
	for _, bridgeKeyItem := range bridgeKey {
		bridgeKeyRule = append(bridgeKeyRule, bridgeKeyItem)
	}
	var witnessRule []interface{}
	for _, witnessItem := range witness {
		witnessRule = append(witnessRule, witnessItem)
	}
	var receiverRule []interface{}
	for _, receiverItem := range receiver {
		receiverRule = append(receiverRule, receiverItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "AddCreateAccountAttestation", bridgeKeyRule, witnessRule, receiverRule)
	if err != nil {
		return nil, err
	}
	return &BridgeAddCreateAccountAttestationIterator{contract: _Bridge.contract, event: "AddCreateAccountAttestation", logs: logs, sub: sub}, nil
}

// WatchAddCreateAccountAttestation is a free log subscription operation binding the contract event 0x64f8fc141a8c2d310942bbc5236041fa0a0c6c39a9fb295be5c959224382815e.
//
// Solidity: event AddCreateAccountAttestation(bytes32 indexed bridgeKey, address indexed witness, address indexed receiver, uint256 value)
func (_Bridge *BridgeFilterer) WatchAddCreateAccountAttestation(opts *bind.WatchOpts, sink chan<- *BridgeAddCreateAccountAttestation, bridgeKey [][32]byte, witness []common.Address, receiver []common.Address) (event.Subscription, error) {

	var bridgeKeyRule []interface{}
	for _, bridgeKeyItem := range bridgeKey {
		bridgeKeyRule = append(bridgeKeyRule, bridgeKeyItem)
	}
	var witnessRule []interface{}
	for _, witnessItem := range witness {
		witnessRule = append(witnessRule, witnessItem)
	}
	var receiverRule []interface{}
	for _, receiverItem := range receiver {
		receiverRule = append(receiverRule, receiverItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "AddCreateAccountAttestation", bridgeKeyRule, witnessRule, receiverRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeAddCreateAccountAttestation)
				if err := _Bridge.contract.UnpackLog(event, "AddCreateAccountAttestation", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAddCreateAccountAttestation is a log parse operation binding the contract event 0x64f8fc141a8c2d310942bbc5236041fa0a0c6c39a9fb295be5c959224382815e.
//
// Solidity: event AddCreateAccountAttestation(bytes32 indexed bridgeKey, address indexed witness, address indexed receiver, uint256 value)
func (_Bridge *BridgeFilterer) ParseAddCreateAccountAttestation(log types.Log) (*BridgeAddCreateAccountAttestation, error) {
	event := new(BridgeAddCreateAccountAttestation)
	if err := _Bridge.contract.UnpackLog(event, "AddCreateAccountAttestation", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeAdminChangedIterator is returned from FilterAdminChanged and is used to iterate over the raw logs and unpacked data for AdminChanged events raised by the Bridge contract.
type BridgeAdminChangedIterator struct {
	Event *BridgeAdminChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeAdminChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeAdminChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeAdminChanged represents a AdminChanged event raised by the Bridge contract.
type BridgeAdminChanged struct {
	PreviousAdmin common.Address
	NewAdmin      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterAdminChanged is a free log retrieval operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_Bridge *BridgeFilterer) FilterAdminChanged(opts *bind.FilterOpts) (*BridgeAdminChangedIterator, error) {

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "AdminChanged")
	if err != nil {
		return nil, err
	}
	return &BridgeAdminChangedIterator{contract: _Bridge.contract, event: "AdminChanged", logs: logs, sub: sub}, nil
}

// WatchAdminChanged is a free log subscription operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_Bridge *BridgeFilterer) WatchAdminChanged(opts *bind.WatchOpts, sink chan<- *BridgeAdminChanged) (event.Subscription, error) {

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "AdminChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeAdminChanged)
				if err := _Bridge.contract.UnpackLog(event, "AdminChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAdminChanged is a log parse operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_Bridge *BridgeFilterer) ParseAdminChanged(log types.Log) (*BridgeAdminChanged, error) {
	event := new(BridgeAdminChanged)
	if err := _Bridge.contract.UnpackLog(event, "AdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeBeaconUpgradedIterator is returned from FilterBeaconUpgraded and is used to iterate over the raw logs and unpacked data for BeaconUpgraded events raised by the Bridge contract.
type BridgeBeaconUpgradedIterator struct {
	Event *BridgeBeaconUpgraded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeBeaconUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeBeaconUpgraded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeBeaconUpgraded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeBeaconUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeBeaconUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeBeaconUpgraded represents a BeaconUpgraded event raised by the Bridge contract.
type BridgeBeaconUpgraded struct {
	Beacon common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterBeaconUpgraded is a free log retrieval operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_Bridge *BridgeFilterer) FilterBeaconUpgraded(opts *bind.FilterOpts, beacon []common.Address) (*BridgeBeaconUpgradedIterator, error) {

	var beaconRule []interface{}
	for _, beaconItem := range beacon {
		beaconRule = append(beaconRule, beaconItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "BeaconUpgraded", beaconRule)
	if err != nil {
		return nil, err
	}
	return &BridgeBeaconUpgradedIterator{contract: _Bridge.contract, event: "BeaconUpgraded", logs: logs, sub: sub}, nil
}

// WatchBeaconUpgraded is a free log subscription operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_Bridge *BridgeFilterer) WatchBeaconUpgraded(opts *bind.WatchOpts, sink chan<- *BridgeBeaconUpgraded, beacon []common.Address) (event.Subscription, error) {

	var beaconRule []interface{}
	for _, beaconItem := range beacon {
		beaconRule = append(beaconRule, beaconItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "BeaconUpgraded", beaconRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeBeaconUpgraded)
				if err := _Bridge.contract.UnpackLog(event, "BeaconUpgraded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseBeaconUpgraded is a log parse operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_Bridge *BridgeFilterer) ParseBeaconUpgraded(log types.Log) (*BridgeBeaconUpgraded, error) {
	event := new(BridgeBeaconUpgraded)
	if err := _Bridge.contract.UnpackLog(event, "BeaconUpgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeClaimIterator is returned from FilterClaim and is used to iterate over the raw logs and unpacked data for Claim events raised by the Bridge contract.
type BridgeClaimIterator struct {
	Event *BridgeClaim // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeClaimIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeClaim)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeClaim)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeClaimIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeClaimIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeClaim represents a Claim event raised by the Bridge contract.
type BridgeClaim struct {
	BridgeKey   [32]byte
	ClaimId     *big.Int
	Sender      common.Address
	Value       *big.Int
	Destination common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterClaim is a free log retrieval operation binding the contract event 0x436897f58db529a6e27c5b7aa31967d35b9b81540e4c797b9322c1740441bf54.
//
// Solidity: event Claim(bytes32 indexed bridgeKey, uint256 indexed claimId, address indexed sender, uint256 value, address destination)
func (_Bridge *BridgeFilterer) FilterClaim(opts *bind.FilterOpts, bridgeKey [][32]byte, claimId []*big.Int, sender []common.Address) (*BridgeClaimIterator, error) {

	var bridgeKeyRule []interface{}
	for _, bridgeKeyItem := range bridgeKey {
		bridgeKeyRule = append(bridgeKeyRule, bridgeKeyItem)
	}
	var claimIdRule []interface{}
	for _, claimIdItem := range claimId {
		claimIdRule = append(claimIdRule, claimIdItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "Claim", bridgeKeyRule, claimIdRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &BridgeClaimIterator{contract: _Bridge.contract, event: "Claim", logs: logs, sub: sub}, nil
}

// WatchClaim is a free log subscription operation binding the contract event 0x436897f58db529a6e27c5b7aa31967d35b9b81540e4c797b9322c1740441bf54.
//
// Solidity: event Claim(bytes32 indexed bridgeKey, uint256 indexed claimId, address indexed sender, uint256 value, address destination)
func (_Bridge *BridgeFilterer) WatchClaim(opts *bind.WatchOpts, sink chan<- *BridgeClaim, bridgeKey [][32]byte, claimId []*big.Int, sender []common.Address) (event.Subscription, error) {

	var bridgeKeyRule []interface{}
	for _, bridgeKeyItem := range bridgeKey {
		bridgeKeyRule = append(bridgeKeyRule, bridgeKeyItem)
	}
	var claimIdRule []interface{}
	for _, claimIdItem := range claimId {
		claimIdRule = append(claimIdRule, claimIdItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "Claim", bridgeKeyRule, claimIdRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeClaim)
				if err := _Bridge.contract.UnpackLog(event, "Claim", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseClaim is a log parse operation binding the contract event 0x436897f58db529a6e27c5b7aa31967d35b9b81540e4c797b9322c1740441bf54.
//
// Solidity: event Claim(bytes32 indexed bridgeKey, uint256 indexed claimId, address indexed sender, uint256 value, address destination)
func (_Bridge *BridgeFilterer) ParseClaim(log types.Log) (*BridgeClaim, error) {
	event := new(BridgeClaim)
	if err := _Bridge.contract.UnpackLog(event, "Claim", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeCommitIterator is returned from FilterCommit and is used to iterate over the raw logs and unpacked data for Commit events raised by the Bridge contract.
type BridgeCommitIterator struct {
	Event *BridgeCommit // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeCommitIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeCommit)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeCommit)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeCommitIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeCommitIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeCommit represents a Commit event raised by the Bridge contract.
type BridgeCommit struct {
	BridgeKey [32]byte
	ClaimId   *big.Int
	Sender    common.Address
	Value     *big.Int
	Receiver  common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterCommit is a free log retrieval operation binding the contract event 0x290bb2c4e47aea59589d24c5b64f7033109290d4636d646112f2d4b442b32a11.
//
// Solidity: event Commit(bytes32 indexed bridgeKey, uint256 indexed claimId, address indexed sender, uint256 value, address receiver)
func (_Bridge *BridgeFilterer) FilterCommit(opts *bind.FilterOpts, bridgeKey [][32]byte, claimId []*big.Int, sender []common.Address) (*BridgeCommitIterator, error) {

	var bridgeKeyRule []interface{}
	for _, bridgeKeyItem := range bridgeKey {
		bridgeKeyRule = append(bridgeKeyRule, bridgeKeyItem)
	}
	var claimIdRule []interface{}
	for _, claimIdItem := range claimId {
		claimIdRule = append(claimIdRule, claimIdItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "Commit", bridgeKeyRule, claimIdRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &BridgeCommitIterator{contract: _Bridge.contract, event: "Commit", logs: logs, sub: sub}, nil
}

// WatchCommit is a free log subscription operation binding the contract event 0x290bb2c4e47aea59589d24c5b64f7033109290d4636d646112f2d4b442b32a11.
//
// Solidity: event Commit(bytes32 indexed bridgeKey, uint256 indexed claimId, address indexed sender, uint256 value, address receiver)
func (_Bridge *BridgeFilterer) WatchCommit(opts *bind.WatchOpts, sink chan<- *BridgeCommit, bridgeKey [][32]byte, claimId []*big.Int, sender []common.Address) (event.Subscription, error) {

	var bridgeKeyRule []interface{}
	for _, bridgeKeyItem := range bridgeKey {
		bridgeKeyRule = append(bridgeKeyRule, bridgeKeyItem)
	}
	var claimIdRule []interface{}
	for _, claimIdItem := range claimId {
		claimIdRule = append(claimIdRule, claimIdItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "Commit", bridgeKeyRule, claimIdRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeCommit)
				if err := _Bridge.contract.UnpackLog(event, "Commit", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCommit is a log parse operation binding the contract event 0x290bb2c4e47aea59589d24c5b64f7033109290d4636d646112f2d4b442b32a11.
//
// Solidity: event Commit(bytes32 indexed bridgeKey, uint256 indexed claimId, address indexed sender, uint256 value, address receiver)
func (_Bridge *BridgeFilterer) ParseCommit(log types.Log) (*BridgeCommit, error) {
	event := new(BridgeCommit)
	if err := _Bridge.contract.UnpackLog(event, "Commit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeCommitWithoutAddressIterator is returned from FilterCommitWithoutAddress and is used to iterate over the raw logs and unpacked data for CommitWithoutAddress events raised by the Bridge contract.
type BridgeCommitWithoutAddressIterator struct {
	Event *BridgeCommitWithoutAddress // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeCommitWithoutAddressIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeCommitWithoutAddress)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeCommitWithoutAddress)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeCommitWithoutAddressIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeCommitWithoutAddressIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeCommitWithoutAddress represents a CommitWithoutAddress event raised by the Bridge contract.
type BridgeCommitWithoutAddress struct {
	BridgeKey [32]byte
	ClaimId   *big.Int
	Sender    common.Address
	Value     *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterCommitWithoutAddress is a free log retrieval operation binding the contract event 0x32783b18313608dabbcd9856301a7fa07369fd1c09a56fbef10659aa5f699fa1.
//
// Solidity: event CommitWithoutAddress(bytes32 indexed bridgeKey, uint256 indexed claimId, address indexed sender, uint256 value)
func (_Bridge *BridgeFilterer) FilterCommitWithoutAddress(opts *bind.FilterOpts, bridgeKey [][32]byte, claimId []*big.Int, sender []common.Address) (*BridgeCommitWithoutAddressIterator, error) {

	var bridgeKeyRule []interface{}
	for _, bridgeKeyItem := range bridgeKey {
		bridgeKeyRule = append(bridgeKeyRule, bridgeKeyItem)
	}
	var claimIdRule []interface{}
	for _, claimIdItem := range claimId {
		claimIdRule = append(claimIdRule, claimIdItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "CommitWithoutAddress", bridgeKeyRule, claimIdRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &BridgeCommitWithoutAddressIterator{contract: _Bridge.contract, event: "CommitWithoutAddress", logs: logs, sub: sub}, nil
}

// WatchCommitWithoutAddress is a free log subscription operation binding the contract event 0x32783b18313608dabbcd9856301a7fa07369fd1c09a56fbef10659aa5f699fa1.
//
// Solidity: event CommitWithoutAddress(bytes32 indexed bridgeKey, uint256 indexed claimId, address indexed sender, uint256 value)
func (_Bridge *BridgeFilterer) WatchCommitWithoutAddress(opts *bind.WatchOpts, sink chan<- *BridgeCommitWithoutAddress, bridgeKey [][32]byte, claimId []*big.Int, sender []common.Address) (event.Subscription, error) {

	var bridgeKeyRule []interface{}
	for _, bridgeKeyItem := range bridgeKey {
		bridgeKeyRule = append(bridgeKeyRule, bridgeKeyItem)
	}
	var claimIdRule []interface{}
	for _, claimIdItem := range claimId {
		claimIdRule = append(claimIdRule, claimIdItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "CommitWithoutAddress", bridgeKeyRule, claimIdRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeCommitWithoutAddress)
				if err := _Bridge.contract.UnpackLog(event, "CommitWithoutAddress", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCommitWithoutAddress is a log parse operation binding the contract event 0x32783b18313608dabbcd9856301a7fa07369fd1c09a56fbef10659aa5f699fa1.
//
// Solidity: event CommitWithoutAddress(bytes32 indexed bridgeKey, uint256 indexed claimId, address indexed sender, uint256 value)
func (_Bridge *BridgeFilterer) ParseCommitWithoutAddress(log types.Log) (*BridgeCommitWithoutAddress, error) {
	event := new(BridgeCommitWithoutAddress)
	if err := _Bridge.contract.UnpackLog(event, "CommitWithoutAddress", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeCreateAccountIterator is returned from FilterCreateAccount and is used to iterate over the raw logs and unpacked data for CreateAccount events raised by the Bridge contract.
type BridgeCreateAccountIterator struct {
	Event *BridgeCreateAccount // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeCreateAccountIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeCreateAccount)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeCreateAccount)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeCreateAccountIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeCreateAccountIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeCreateAccount represents a CreateAccount event raised by the Bridge contract.
type BridgeCreateAccount struct {
	Receiver common.Address
	Value    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterCreateAccount is a free log retrieval operation binding the contract event 0x85841522199c696c3d4a549fea06732153559ded5db5cf6dfa3bb099827f2c84.
//
// Solidity: event CreateAccount(address indexed receiver, uint256 value)
func (_Bridge *BridgeFilterer) FilterCreateAccount(opts *bind.FilterOpts, receiver []common.Address) (*BridgeCreateAccountIterator, error) {

	var receiverRule []interface{}
	for _, receiverItem := range receiver {
		receiverRule = append(receiverRule, receiverItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "CreateAccount", receiverRule)
	if err != nil {
		return nil, err
	}
	return &BridgeCreateAccountIterator{contract: _Bridge.contract, event: "CreateAccount", logs: logs, sub: sub}, nil
}

// WatchCreateAccount is a free log subscription operation binding the contract event 0x85841522199c696c3d4a549fea06732153559ded5db5cf6dfa3bb099827f2c84.
//
// Solidity: event CreateAccount(address indexed receiver, uint256 value)
func (_Bridge *BridgeFilterer) WatchCreateAccount(opts *bind.WatchOpts, sink chan<- *BridgeCreateAccount, receiver []common.Address) (event.Subscription, error) {

	var receiverRule []interface{}
	for _, receiverItem := range receiver {
		receiverRule = append(receiverRule, receiverItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "CreateAccount", receiverRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeCreateAccount)
				if err := _Bridge.contract.UnpackLog(event, "CreateAccount", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCreateAccount is a log parse operation binding the contract event 0x85841522199c696c3d4a549fea06732153559ded5db5cf6dfa3bb099827f2c84.
//
// Solidity: event CreateAccount(address indexed receiver, uint256 value)
func (_Bridge *BridgeFilterer) ParseCreateAccount(log types.Log) (*BridgeCreateAccount, error) {
	event := new(BridgeCreateAccount)
	if err := _Bridge.contract.UnpackLog(event, "CreateAccount", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeCreateAccountCommitIterator is returned from FilterCreateAccountCommit and is used to iterate over the raw logs and unpacked data for CreateAccountCommit events raised by the Bridge contract.
type BridgeCreateAccountCommitIterator struct {
	Event *BridgeCreateAccountCommit // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeCreateAccountCommitIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeCreateAccountCommit)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeCreateAccountCommit)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeCreateAccountCommitIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeCreateAccountCommitIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeCreateAccountCommit represents a CreateAccountCommit event raised by the Bridge contract.
type BridgeCreateAccountCommit struct {
	BridgeKey       [32]byte
	Creator         common.Address
	Destination     common.Address
	Value           *big.Int
	SignatureReward *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterCreateAccountCommit is a free log retrieval operation binding the contract event 0x32ebca0d0dd8fc03a488dcaab22112132af86cbc4c2df086bc7f328d751f5d7e.
//
// Solidity: event CreateAccountCommit(bytes32 indexed bridgeKey, address indexed creator, address indexed destination, uint256 value, uint256 signatureReward)
func (_Bridge *BridgeFilterer) FilterCreateAccountCommit(opts *bind.FilterOpts, bridgeKey [][32]byte, creator []common.Address, destination []common.Address) (*BridgeCreateAccountCommitIterator, error) {

	var bridgeKeyRule []interface{}
	for _, bridgeKeyItem := range bridgeKey {
		bridgeKeyRule = append(bridgeKeyRule, bridgeKeyItem)
	}
	var creatorRule []interface{}
	for _, creatorItem := range creator {
		creatorRule = append(creatorRule, creatorItem)
	}
	var destinationRule []interface{}
	for _, destinationItem := range destination {
		destinationRule = append(destinationRule, destinationItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "CreateAccountCommit", bridgeKeyRule, creatorRule, destinationRule)
	if err != nil {
		return nil, err
	}
	return &BridgeCreateAccountCommitIterator{contract: _Bridge.contract, event: "CreateAccountCommit", logs: logs, sub: sub}, nil
}

// WatchCreateAccountCommit is a free log subscription operation binding the contract event 0x32ebca0d0dd8fc03a488dcaab22112132af86cbc4c2df086bc7f328d751f5d7e.
//
// Solidity: event CreateAccountCommit(bytes32 indexed bridgeKey, address indexed creator, address indexed destination, uint256 value, uint256 signatureReward)
func (_Bridge *BridgeFilterer) WatchCreateAccountCommit(opts *bind.WatchOpts, sink chan<- *BridgeCreateAccountCommit, bridgeKey [][32]byte, creator []common.Address, destination []common.Address) (event.Subscription, error) {

	var bridgeKeyRule []interface{}
	for _, bridgeKeyItem := range bridgeKey {
		bridgeKeyRule = append(bridgeKeyRule, bridgeKeyItem)
	}
	var creatorRule []interface{}
	for _, creatorItem := range creator {
		creatorRule = append(creatorRule, creatorItem)
	}
	var destinationRule []interface{}
	for _, destinationItem := range destination {
		destinationRule = append(destinationRule, destinationItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "CreateAccountCommit", bridgeKeyRule, creatorRule, destinationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeCreateAccountCommit)
				if err := _Bridge.contract.UnpackLog(event, "CreateAccountCommit", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCreateAccountCommit is a log parse operation binding the contract event 0x32ebca0d0dd8fc03a488dcaab22112132af86cbc4c2df086bc7f328d751f5d7e.
//
// Solidity: event CreateAccountCommit(bytes32 indexed bridgeKey, address indexed creator, address indexed destination, uint256 value, uint256 signatureReward)
func (_Bridge *BridgeFilterer) ParseCreateAccountCommit(log types.Log) (*BridgeCreateAccountCommit, error) {
	event := new(BridgeCreateAccountCommit)
	if err := _Bridge.contract.UnpackLog(event, "CreateAccountCommit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeCreateBridgeIterator is returned from FilterCreateBridge and is used to iterate over the raw logs and unpacked data for CreateBridge events raised by the Bridge contract.
type BridgeCreateBridgeIterator struct {
	Event *BridgeCreateBridge // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeCreateBridgeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeCreateBridge)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeCreateBridge)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeCreateBridgeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeCreateBridgeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeCreateBridge represents a CreateBridge event raised by the Bridge contract.
type BridgeCreateBridge struct {
	BridgeKey                 [32]byte
	LockingChainDoor          common.Address
	LockingChainIssueIssuer   common.Address
	LockingChainIssueCurrency string
	IssuingChainDoor          common.Address
	IssuingChainIssueIssuer   common.Address
	IssuingChainIssueCurrency string
	Raw                       types.Log // Blockchain specific contextual infos
}

// FilterCreateBridge is a free log retrieval operation binding the contract event 0xe8155fae5a2ff8b2ff714dc22bb65489bdf14794bdc7e5802d7ba8e62f0c3ccf.
//
// Solidity: event CreateBridge(bytes32 indexed bridgeKey, address lockingChainDoor, address lockingChainIssueIssuer, string lockingChainIssueCurrency, address issuingChainDoor, address issuingChainIssueIssuer, string issuingChainIssueCurrency)
func (_Bridge *BridgeFilterer) FilterCreateBridge(opts *bind.FilterOpts, bridgeKey [][32]byte) (*BridgeCreateBridgeIterator, error) {

	var bridgeKeyRule []interface{}
	for _, bridgeKeyItem := range bridgeKey {
		bridgeKeyRule = append(bridgeKeyRule, bridgeKeyItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "CreateBridge", bridgeKeyRule)
	if err != nil {
		return nil, err
	}
	return &BridgeCreateBridgeIterator{contract: _Bridge.contract, event: "CreateBridge", logs: logs, sub: sub}, nil
}

// WatchCreateBridge is a free log subscription operation binding the contract event 0xe8155fae5a2ff8b2ff714dc22bb65489bdf14794bdc7e5802d7ba8e62f0c3ccf.
//
// Solidity: event CreateBridge(bytes32 indexed bridgeKey, address lockingChainDoor, address lockingChainIssueIssuer, string lockingChainIssueCurrency, address issuingChainDoor, address issuingChainIssueIssuer, string issuingChainIssueCurrency)
func (_Bridge *BridgeFilterer) WatchCreateBridge(opts *bind.WatchOpts, sink chan<- *BridgeCreateBridge, bridgeKey [][32]byte) (event.Subscription, error) {

	var bridgeKeyRule []interface{}
	for _, bridgeKeyItem := range bridgeKey {
		bridgeKeyRule = append(bridgeKeyRule, bridgeKeyItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "CreateBridge", bridgeKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeCreateBridge)
				if err := _Bridge.contract.UnpackLog(event, "CreateBridge", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCreateBridge is a log parse operation binding the contract event 0xe8155fae5a2ff8b2ff714dc22bb65489bdf14794bdc7e5802d7ba8e62f0c3ccf.
//
// Solidity: event CreateBridge(bytes32 indexed bridgeKey, address lockingChainDoor, address lockingChainIssueIssuer, string lockingChainIssueCurrency, address issuingChainDoor, address issuingChainIssueIssuer, string issuingChainIssueCurrency)
func (_Bridge *BridgeFilterer) ParseCreateBridge(log types.Log) (*BridgeCreateBridge, error) {
	event := new(BridgeCreateBridge)
	if err := _Bridge.contract.UnpackLog(event, "CreateBridge", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeCreateBridgeRequestIterator is returned from FilterCreateBridgeRequest and is used to iterate over the raw logs and unpacked data for CreateBridgeRequest events raised by the Bridge contract.
type BridgeCreateBridgeRequestIterator struct {
	Event *BridgeCreateBridgeRequest // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeCreateBridgeRequestIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeCreateBridgeRequest)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeCreateBridgeRequest)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeCreateBridgeRequestIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeCreateBridgeRequestIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeCreateBridgeRequest represents a CreateBridgeRequest event raised by the Bridge contract.
type BridgeCreateBridgeRequest struct {
	TokenAddress common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterCreateBridgeRequest is a free log retrieval operation binding the contract event 0x54942c22b4f7613321d895cad0749836eeb5c8b282c630c4478e07913f814dbc.
//
// Solidity: event CreateBridgeRequest(address tokenAddress)
func (_Bridge *BridgeFilterer) FilterCreateBridgeRequest(opts *bind.FilterOpts) (*BridgeCreateBridgeRequestIterator, error) {

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "CreateBridgeRequest")
	if err != nil {
		return nil, err
	}
	return &BridgeCreateBridgeRequestIterator{contract: _Bridge.contract, event: "CreateBridgeRequest", logs: logs, sub: sub}, nil
}

// WatchCreateBridgeRequest is a free log subscription operation binding the contract event 0x54942c22b4f7613321d895cad0749836eeb5c8b282c630c4478e07913f814dbc.
//
// Solidity: event CreateBridgeRequest(address tokenAddress)
func (_Bridge *BridgeFilterer) WatchCreateBridgeRequest(opts *bind.WatchOpts, sink chan<- *BridgeCreateBridgeRequest) (event.Subscription, error) {

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "CreateBridgeRequest")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeCreateBridgeRequest)
				if err := _Bridge.contract.UnpackLog(event, "CreateBridgeRequest", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCreateBridgeRequest is a log parse operation binding the contract event 0x54942c22b4f7613321d895cad0749836eeb5c8b282c630c4478e07913f814dbc.
//
// Solidity: event CreateBridgeRequest(address tokenAddress)
func (_Bridge *BridgeFilterer) ParseCreateBridgeRequest(log types.Log) (*BridgeCreateBridgeRequest, error) {
	event := new(BridgeCreateBridgeRequest)
	if err := _Bridge.contract.UnpackLog(event, "CreateBridgeRequest", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeCreateClaimIterator is returned from FilterCreateClaim and is used to iterate over the raw logs and unpacked data for CreateClaim events raised by the Bridge contract.
type BridgeCreateClaimIterator struct {
	Event *BridgeCreateClaim // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeCreateClaimIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeCreateClaim)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeCreateClaim)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeCreateClaimIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeCreateClaimIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeCreateClaim represents a CreateClaim event raised by the Bridge contract.
type BridgeCreateClaim struct {
	BridgeKey [32]byte
	ClaimId   *big.Int
	Creator   common.Address
	Sender    common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterCreateClaim is a free log retrieval operation binding the contract event 0xc7ecca132ed5d1d6c462587819023eee197ef7fb00b399bbfc2ce032587f0c6d.
//
// Solidity: event CreateClaim(bytes32 indexed bridgeKey, uint256 indexed claimId, address indexed creator, address sender)
func (_Bridge *BridgeFilterer) FilterCreateClaim(opts *bind.FilterOpts, bridgeKey [][32]byte, claimId []*big.Int, creator []common.Address) (*BridgeCreateClaimIterator, error) {

	var bridgeKeyRule []interface{}
	for _, bridgeKeyItem := range bridgeKey {
		bridgeKeyRule = append(bridgeKeyRule, bridgeKeyItem)
	}
	var claimIdRule []interface{}
	for _, claimIdItem := range claimId {
		claimIdRule = append(claimIdRule, claimIdItem)
	}
	var creatorRule []interface{}
	for _, creatorItem := range creator {
		creatorRule = append(creatorRule, creatorItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "CreateClaim", bridgeKeyRule, claimIdRule, creatorRule)
	if err != nil {
		return nil, err
	}
	return &BridgeCreateClaimIterator{contract: _Bridge.contract, event: "CreateClaim", logs: logs, sub: sub}, nil
}

// WatchCreateClaim is a free log subscription operation binding the contract event 0xc7ecca132ed5d1d6c462587819023eee197ef7fb00b399bbfc2ce032587f0c6d.
//
// Solidity: event CreateClaim(bytes32 indexed bridgeKey, uint256 indexed claimId, address indexed creator, address sender)
func (_Bridge *BridgeFilterer) WatchCreateClaim(opts *bind.WatchOpts, sink chan<- *BridgeCreateClaim, bridgeKey [][32]byte, claimId []*big.Int, creator []common.Address) (event.Subscription, error) {

	var bridgeKeyRule []interface{}
	for _, bridgeKeyItem := range bridgeKey {
		bridgeKeyRule = append(bridgeKeyRule, bridgeKeyItem)
	}
	var claimIdRule []interface{}
	for _, claimIdItem := range claimId {
		claimIdRule = append(claimIdRule, claimIdItem)
	}
	var creatorRule []interface{}
	for _, creatorItem := range creator {
		creatorRule = append(creatorRule, creatorItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "CreateClaim", bridgeKeyRule, claimIdRule, creatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeCreateClaim)
				if err := _Bridge.contract.UnpackLog(event, "CreateClaim", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCreateClaim is a log parse operation binding the contract event 0xc7ecca132ed5d1d6c462587819023eee197ef7fb00b399bbfc2ce032587f0c6d.
//
// Solidity: event CreateClaim(bytes32 indexed bridgeKey, uint256 indexed claimId, address indexed creator, address sender)
func (_Bridge *BridgeFilterer) ParseCreateClaim(log types.Log) (*BridgeCreateClaim, error) {
	event := new(BridgeCreateClaim)
	if err := _Bridge.contract.UnpackLog(event, "CreateClaim", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeCreditIterator is returned from FilterCredit and is used to iterate over the raw logs and unpacked data for Credit events raised by the Bridge contract.
type BridgeCreditIterator struct {
	Event *BridgeCredit // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeCreditIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeCredit)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeCredit)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeCreditIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeCreditIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeCredit represents a Credit event raised by the Bridge contract.
type BridgeCredit struct {
	BridgeKey [32]byte
	ClaimId   *big.Int
	Receiver  common.Address
	Value     *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterCredit is a free log retrieval operation binding the contract event 0x0087db26e45ef9d7e62d7966c0bc6310075c3e120cae4af40d1791e1b01f7e71.
//
// Solidity: event Credit(bytes32 indexed bridgeKey, uint256 indexed claimId, address indexed receiver, uint256 value)
func (_Bridge *BridgeFilterer) FilterCredit(opts *bind.FilterOpts, bridgeKey [][32]byte, claimId []*big.Int, receiver []common.Address) (*BridgeCreditIterator, error) {

	var bridgeKeyRule []interface{}
	for _, bridgeKeyItem := range bridgeKey {
		bridgeKeyRule = append(bridgeKeyRule, bridgeKeyItem)
	}
	var claimIdRule []interface{}
	for _, claimIdItem := range claimId {
		claimIdRule = append(claimIdRule, claimIdItem)
	}
	var receiverRule []interface{}
	for _, receiverItem := range receiver {
		receiverRule = append(receiverRule, receiverItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "Credit", bridgeKeyRule, claimIdRule, receiverRule)
	if err != nil {
		return nil, err
	}
	return &BridgeCreditIterator{contract: _Bridge.contract, event: "Credit", logs: logs, sub: sub}, nil
}

// WatchCredit is a free log subscription operation binding the contract event 0x0087db26e45ef9d7e62d7966c0bc6310075c3e120cae4af40d1791e1b01f7e71.
//
// Solidity: event Credit(bytes32 indexed bridgeKey, uint256 indexed claimId, address indexed receiver, uint256 value)
func (_Bridge *BridgeFilterer) WatchCredit(opts *bind.WatchOpts, sink chan<- *BridgeCredit, bridgeKey [][32]byte, claimId []*big.Int, receiver []common.Address) (event.Subscription, error) {

	var bridgeKeyRule []interface{}
	for _, bridgeKeyItem := range bridgeKey {
		bridgeKeyRule = append(bridgeKeyRule, bridgeKeyItem)
	}
	var claimIdRule []interface{}
	for _, claimIdItem := range claimId {
		claimIdRule = append(claimIdRule, claimIdItem)
	}
	var receiverRule []interface{}
	for _, receiverItem := range receiver {
		receiverRule = append(receiverRule, receiverItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "Credit", bridgeKeyRule, claimIdRule, receiverRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeCredit)
				if err := _Bridge.contract.UnpackLog(event, "Credit", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCredit is a log parse operation binding the contract event 0x0087db26e45ef9d7e62d7966c0bc6310075c3e120cae4af40d1791e1b01f7e71.
//
// Solidity: event Credit(bytes32 indexed bridgeKey, uint256 indexed claimId, address indexed receiver, uint256 value)
func (_Bridge *BridgeFilterer) ParseCredit(log types.Log) (*BridgeCredit, error) {
	event := new(BridgeCredit)
	if err := _Bridge.contract.UnpackLog(event, "Credit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Bridge contract.
type BridgeInitializedIterator struct {
	Event *BridgeInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeInitialized represents a Initialized event raised by the Bridge contract.
type BridgeInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Bridge *BridgeFilterer) FilterInitialized(opts *bind.FilterOpts) (*BridgeInitializedIterator, error) {

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &BridgeInitializedIterator{contract: _Bridge.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Bridge *BridgeFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *BridgeInitialized) (event.Subscription, error) {

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeInitialized)
				if err := _Bridge.contract.UnpackLog(event, "Initialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Bridge *BridgeFilterer) ParseInitialized(log types.Log) (*BridgeInitialized, error) {
	event := new(BridgeInitialized)
	if err := _Bridge.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Bridge contract.
type BridgeOwnershipTransferredIterator struct {
	Event *BridgeOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeOwnershipTransferred represents a OwnershipTransferred event raised by the Bridge contract.
type BridgeOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Bridge *BridgeFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BridgeOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BridgeOwnershipTransferredIterator{contract: _Bridge.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Bridge *BridgeFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BridgeOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeOwnershipTransferred)
				if err := _Bridge.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Bridge *BridgeFilterer) ParseOwnershipTransferred(log types.Log) (*BridgeOwnershipTransferred, error) {
	event := new(BridgeOwnershipTransferred)
	if err := _Bridge.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgePausedIterator is returned from FilterPaused and is used to iterate over the raw logs and unpacked data for Paused events raised by the Bridge contract.
type BridgePausedIterator struct {
	Event *BridgePaused // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgePausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgePaused)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgePaused)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgePausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgePausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgePaused represents a Paused event raised by the Bridge contract.
type BridgePaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPaused is a free log retrieval operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_Bridge *BridgeFilterer) FilterPaused(opts *bind.FilterOpts) (*BridgePausedIterator, error) {

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return &BridgePausedIterator{contract: _Bridge.contract, event: "Paused", logs: logs, sub: sub}, nil
}

// WatchPaused is a free log subscription operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_Bridge *BridgeFilterer) WatchPaused(opts *bind.WatchOpts, sink chan<- *BridgePaused) (event.Subscription, error) {

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgePaused)
				if err := _Bridge.contract.UnpackLog(event, "Paused", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePaused is a log parse operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_Bridge *BridgeFilterer) ParsePaused(log types.Log) (*BridgePaused, error) {
	event := new(BridgePaused)
	if err := _Bridge.contract.UnpackLog(event, "Paused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeUnpausedIterator is returned from FilterUnpaused and is used to iterate over the raw logs and unpacked data for Unpaused events raised by the Bridge contract.
type BridgeUnpausedIterator struct {
	Event *BridgeUnpaused // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeUnpausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeUnpaused)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeUnpaused)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeUnpausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeUnpausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeUnpaused represents a Unpaused event raised by the Bridge contract.
type BridgeUnpaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterUnpaused is a free log retrieval operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_Bridge *BridgeFilterer) FilterUnpaused(opts *bind.FilterOpts) (*BridgeUnpausedIterator, error) {

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return &BridgeUnpausedIterator{contract: _Bridge.contract, event: "Unpaused", logs: logs, sub: sub}, nil
}

// WatchUnpaused is a free log subscription operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_Bridge *BridgeFilterer) WatchUnpaused(opts *bind.WatchOpts, sink chan<- *BridgeUnpaused) (event.Subscription, error) {

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeUnpaused)
				if err := _Bridge.contract.UnpackLog(event, "Unpaused", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUnpaused is a log parse operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_Bridge *BridgeFilterer) ParseUnpaused(log types.Log) (*BridgeUnpaused, error) {
	event := new(BridgeUnpaused)
	if err := _Bridge.contract.UnpackLog(event, "Unpaused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeUpgradedIterator is returned from FilterUpgraded and is used to iterate over the raw logs and unpacked data for Upgraded events raised by the Bridge contract.
type BridgeUpgradedIterator struct {
	Event *BridgeUpgraded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeUpgraded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeUpgraded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeUpgraded represents a Upgraded event raised by the Bridge contract.
type BridgeUpgraded struct {
	Implementation common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpgraded is a free log retrieval operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_Bridge *BridgeFilterer) FilterUpgraded(opts *bind.FilterOpts, implementation []common.Address) (*BridgeUpgradedIterator, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return &BridgeUpgradedIterator{contract: _Bridge.contract, event: "Upgraded", logs: logs, sub: sub}, nil
}

// WatchUpgraded is a free log subscription operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_Bridge *BridgeFilterer) WatchUpgraded(opts *bind.WatchOpts, sink chan<- *BridgeUpgraded, implementation []common.Address) (event.Subscription, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeUpgraded)
				if err := _Bridge.contract.UnpackLog(event, "Upgraded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUpgraded is a log parse operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_Bridge *BridgeFilterer) ParseUpgraded(log types.Log) (*BridgeUpgraded, error) {
	event := new(BridgeUpgraded)
	if err := _Bridge.contract.UnpackLog(event, "Upgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
