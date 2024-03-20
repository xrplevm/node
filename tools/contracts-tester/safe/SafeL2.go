// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package safe

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

// SafeMetaData contains all meta data concerning the Safe contract.
var SafeMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"AddedOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"approvedHash\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"ApproveHash\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"handler\",\"type\":\"address\"}],\"name\":\"ChangedFallbackHandler\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"guard\",\"type\":\"address\"}],\"name\":\"ChangedGuard\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"threshold\",\"type\":\"uint256\"}],\"name\":\"ChangedThreshold\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"DisabledModule\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"EnabledModule\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"txHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"payment\",\"type\":\"uint256\"}],\"name\":\"ExecutionFailure\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"ExecutionFromModuleFailure\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"ExecutionFromModuleSuccess\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"txHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"payment\",\"type\":\"uint256\"}],\"name\":\"ExecutionSuccess\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"RemovedOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"enumEnum.Operation\",\"name\":\"operation\",\"type\":\"uint8\"}],\"name\":\"SafeModuleTransaction\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"enumEnum.Operation\",\"name\":\"operation\",\"type\":\"uint8\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"safeTxGas\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"baseGas\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"gasPrice\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"gasToken\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"addresspayable\",\"name\":\"refundReceiver\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"additionalInfo\",\"type\":\"bytes\"}],\"name\":\"SafeMultiSigTransaction\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"SafeReceived\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"initiator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address[]\",\"name\":\"owners\",\"type\":\"address[]\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"threshold\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"initializer\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"fallbackHandler\",\"type\":\"address\"}],\"name\":\"SafeSetup\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"msgHash\",\"type\":\"bytes32\"}],\"name\":\"SignMsg\",\"type\":\"event\"},{\"stateMutability\":\"nonpayable\",\"type\":\"fallback\"},{\"inputs\":[],\"name\":\"VERSION\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_threshold\",\"type\":\"uint256\"}],\"name\":\"addOwnerWithThreshold\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"hashToApprove\",\"type\":\"bytes32\"}],\"name\":\"approveHash\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"approvedHashes\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_threshold\",\"type\":\"uint256\"}],\"name\":\"changeThreshold\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"executor\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"dataHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"requiredSignatures\",\"type\":\"uint256\"}],\"name\":\"checkNSignatures\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"dataHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"}],\"name\":\"checkSignatures\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"prevModule\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"disableModule\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"domainSeparator\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"enableModule\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"enumEnum.Operation\",\"name\":\"operation\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"safeTxGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"baseGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasPrice\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"gasToken\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"refundReceiver\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"}],\"name\":\"execTransaction\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"enumEnum.Operation\",\"name\":\"operation\",\"type\":\"uint8\"}],\"name\":\"execTransactionFromModule\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"enumEnum.Operation\",\"name\":\"operation\",\"type\":\"uint8\"}],\"name\":\"execTransactionFromModuleReturnData\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes\",\"name\":\"returnData\",\"type\":\"bytes\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"start\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"pageSize\",\"type\":\"uint256\"}],\"name\":\"getModulesPaginated\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"array\",\"type\":\"address[]\"},{\"internalType\":\"address\",\"name\":\"next\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwners\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"length\",\"type\":\"uint256\"}],\"name\":\"getStorageAt\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getThreshold\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"enumEnum.Operation\",\"name\":\"operation\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"safeTxGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"baseGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasPrice\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"gasToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"refundReceiver\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"getTransactionHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"}],\"name\":\"isModuleEnabled\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"isOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"nonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"prevOwner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_threshold\",\"type\":\"uint256\"}],\"name\":\"removeOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"handler\",\"type\":\"address\"}],\"name\":\"setFallbackHandler\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"guard\",\"type\":\"address\"}],\"name\":\"setGuard\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_owners\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"_threshold\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"fallbackHandler\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"paymentToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"payment\",\"type\":\"uint256\"},{\"internalType\":\"addresspayable\",\"name\":\"paymentReceiver\",\"type\":\"address\"}],\"name\":\"setup\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"signedMessages\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"targetContract\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"calldataPayload\",\"type\":\"bytes\"}],\"name\":\"simulateAndRevert\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"prevOwner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"oldOwner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"swapOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
}

// SafeABI is the input ABI used to generate the binding from.
// Deprecated: Use SafeMetaData.ABI instead.
var SafeABI = SafeMetaData.ABI

// Safe is an auto generated Go binding around an Ethereum contract.
type Safe struct {
	SafeCaller     // Read-only binding to the contract
	SafeTransactor // Write-only binding to the contract
	SafeFilterer   // Log filterer for contract events
}

// SafeCaller is an auto generated read-only Go binding around an Ethereum contract.
type SafeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SafeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SafeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SafeSession struct {
	Contract     *Safe             // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SafeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SafeCallerSession struct {
	Contract *SafeCaller   // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// SafeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SafeTransactorSession struct {
	Contract     *SafeTransactor   // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SafeRaw is an auto generated low-level Go binding around an Ethereum contract.
type SafeRaw struct {
	Contract *Safe // Generic contract binding to access the raw methods on
}

// SafeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SafeCallerRaw struct {
	Contract *SafeCaller // Generic read-only contract binding to access the raw methods on
}

// SafeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SafeTransactorRaw struct {
	Contract *SafeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSafe creates a new instance of Safe, bound to a specific deployed contract.
func NewSafe(address common.Address, backend bind.ContractBackend) (*Safe, error) {
	contract, err := bindSafe(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Safe{SafeCaller: SafeCaller{contract: contract}, SafeTransactor: SafeTransactor{contract: contract}, SafeFilterer: SafeFilterer{contract: contract}}, nil
}

// NewSafeCaller creates a new read-only instance of Safe, bound to a specific deployed contract.
func NewSafeCaller(address common.Address, caller bind.ContractCaller) (*SafeCaller, error) {
	contract, err := bindSafe(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SafeCaller{contract: contract}, nil
}

// NewSafeTransactor creates a new write-only instance of Safe, bound to a specific deployed contract.
func NewSafeTransactor(address common.Address, transactor bind.ContractTransactor) (*SafeTransactor, error) {
	contract, err := bindSafe(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SafeTransactor{contract: contract}, nil
}

// NewSafeFilterer creates a new log filterer instance of Safe, bound to a specific deployed contract.
func NewSafeFilterer(address common.Address, filterer bind.ContractFilterer) (*SafeFilterer, error) {
	contract, err := bindSafe(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SafeFilterer{contract: contract}, nil
}

// bindSafe binds a generic wrapper to an already deployed contract.
func bindSafe(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SafeABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Safe *SafeRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Safe.Contract.SafeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Safe *SafeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Safe.Contract.SafeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Safe *SafeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Safe.Contract.SafeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Safe *SafeCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Safe.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Safe *SafeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Safe.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Safe *SafeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Safe.Contract.contract.Transact(opts, method, params...)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_Safe *SafeCaller) VERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Safe.contract.Call(opts, &out, "VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_Safe *SafeSession) VERSION() (string, error) {
	return _Safe.Contract.VERSION(&_Safe.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_Safe *SafeCallerSession) VERSION() (string, error) {
	return _Safe.Contract.VERSION(&_Safe.CallOpts)
}

// ApprovedHashes is a free data retrieval call binding the contract method 0x7d832974.
//
// Solidity: function approvedHashes(address , bytes32 ) view returns(uint256)
func (_Safe *SafeCaller) ApprovedHashes(opts *bind.CallOpts, arg0 common.Address, arg1 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Safe.contract.Call(opts, &out, "approvedHashes", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ApprovedHashes is a free data retrieval call binding the contract method 0x7d832974.
//
// Solidity: function approvedHashes(address , bytes32 ) view returns(uint256)
func (_Safe *SafeSession) ApprovedHashes(arg0 common.Address, arg1 [32]byte) (*big.Int, error) {
	return _Safe.Contract.ApprovedHashes(&_Safe.CallOpts, arg0, arg1)
}

// ApprovedHashes is a free data retrieval call binding the contract method 0x7d832974.
//
// Solidity: function approvedHashes(address , bytes32 ) view returns(uint256)
func (_Safe *SafeCallerSession) ApprovedHashes(arg0 common.Address, arg1 [32]byte) (*big.Int, error) {
	return _Safe.Contract.ApprovedHashes(&_Safe.CallOpts, arg0, arg1)
}

// CheckNSignatures is a free data retrieval call binding the contract method 0xcd5d1f77.
//
// Solidity: function checkNSignatures(address executor, bytes32 dataHash, bytes , bytes signatures, uint256 requiredSignatures) view returns()
func (_Safe *SafeCaller) CheckNSignatures(opts *bind.CallOpts, executor common.Address, dataHash [32]byte, arg2 []byte, signatures []byte, requiredSignatures *big.Int) error {
	var out []interface{}
	err := _Safe.contract.Call(opts, &out, "checkNSignatures", executor, dataHash, arg2, signatures, requiredSignatures)

	if err != nil {
		return err
	}

	return err

}

// CheckNSignatures is a free data retrieval call binding the contract method 0xcd5d1f77.
//
// Solidity: function checkNSignatures(address executor, bytes32 dataHash, bytes , bytes signatures, uint256 requiredSignatures) view returns()
func (_Safe *SafeSession) CheckNSignatures(executor common.Address, dataHash [32]byte, arg2 []byte, signatures []byte, requiredSignatures *big.Int) error {
	return _Safe.Contract.CheckNSignatures(&_Safe.CallOpts, executor, dataHash, arg2, signatures, requiredSignatures)
}

// CheckNSignatures is a free data retrieval call binding the contract method 0xcd5d1f77.
//
// Solidity: function checkNSignatures(address executor, bytes32 dataHash, bytes , bytes signatures, uint256 requiredSignatures) view returns()
func (_Safe *SafeCallerSession) CheckNSignatures(executor common.Address, dataHash [32]byte, arg2 []byte, signatures []byte, requiredSignatures *big.Int) error {
	return _Safe.Contract.CheckNSignatures(&_Safe.CallOpts, executor, dataHash, arg2, signatures, requiredSignatures)
}

// CheckSignatures is a free data retrieval call binding the contract method 0x934f3a11.
//
// Solidity: function checkSignatures(bytes32 dataHash, bytes data, bytes signatures) view returns()
func (_Safe *SafeCaller) CheckSignatures(opts *bind.CallOpts, dataHash [32]byte, data []byte, signatures []byte) error {
	var out []interface{}
	err := _Safe.contract.Call(opts, &out, "checkSignatures", dataHash, data, signatures)

	if err != nil {
		return err
	}

	return err

}

// CheckSignatures is a free data retrieval call binding the contract method 0x934f3a11.
//
// Solidity: function checkSignatures(bytes32 dataHash, bytes data, bytes signatures) view returns()
func (_Safe *SafeSession) CheckSignatures(dataHash [32]byte, data []byte, signatures []byte) error {
	return _Safe.Contract.CheckSignatures(&_Safe.CallOpts, dataHash, data, signatures)
}

// CheckSignatures is a free data retrieval call binding the contract method 0x934f3a11.
//
// Solidity: function checkSignatures(bytes32 dataHash, bytes data, bytes signatures) view returns()
func (_Safe *SafeCallerSession) CheckSignatures(dataHash [32]byte, data []byte, signatures []byte) error {
	return _Safe.Contract.CheckSignatures(&_Safe.CallOpts, dataHash, data, signatures)
}

// DomainSeparator is a free data retrieval call binding the contract method 0xf698da25.
//
// Solidity: function domainSeparator() view returns(bytes32)
func (_Safe *SafeCaller) DomainSeparator(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Safe.contract.Call(opts, &out, "domainSeparator")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DomainSeparator is a free data retrieval call binding the contract method 0xf698da25.
//
// Solidity: function domainSeparator() view returns(bytes32)
func (_Safe *SafeSession) DomainSeparator() ([32]byte, error) {
	return _Safe.Contract.DomainSeparator(&_Safe.CallOpts)
}

// DomainSeparator is a free data retrieval call binding the contract method 0xf698da25.
//
// Solidity: function domainSeparator() view returns(bytes32)
func (_Safe *SafeCallerSession) DomainSeparator() ([32]byte, error) {
	return _Safe.Contract.DomainSeparator(&_Safe.CallOpts)
}

// GetModulesPaginated is a free data retrieval call binding the contract method 0xcc2f8452.
//
// Solidity: function getModulesPaginated(address start, uint256 pageSize) view returns(address[] array, address next)
func (_Safe *SafeCaller) GetModulesPaginated(opts *bind.CallOpts, start common.Address, pageSize *big.Int) (struct {
	Array []common.Address
	Next  common.Address
}, error) {
	var out []interface{}
	err := _Safe.contract.Call(opts, &out, "getModulesPaginated", start, pageSize)

	outstruct := new(struct {
		Array []common.Address
		Next  common.Address
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Array = *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)
	outstruct.Next = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)

	return *outstruct, err

}

// GetModulesPaginated is a free data retrieval call binding the contract method 0xcc2f8452.
//
// Solidity: function getModulesPaginated(address start, uint256 pageSize) view returns(address[] array, address next)
func (_Safe *SafeSession) GetModulesPaginated(start common.Address, pageSize *big.Int) (struct {
	Array []common.Address
	Next  common.Address
}, error) {
	return _Safe.Contract.GetModulesPaginated(&_Safe.CallOpts, start, pageSize)
}

// GetModulesPaginated is a free data retrieval call binding the contract method 0xcc2f8452.
//
// Solidity: function getModulesPaginated(address start, uint256 pageSize) view returns(address[] array, address next)
func (_Safe *SafeCallerSession) GetModulesPaginated(start common.Address, pageSize *big.Int) (struct {
	Array []common.Address
	Next  common.Address
}, error) {
	return _Safe.Contract.GetModulesPaginated(&_Safe.CallOpts, start, pageSize)
}

// GetOwners is a free data retrieval call binding the contract method 0xa0e67e2b.
//
// Solidity: function getOwners() view returns(address[])
func (_Safe *SafeCaller) GetOwners(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _Safe.contract.Call(opts, &out, "getOwners")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetOwners is a free data retrieval call binding the contract method 0xa0e67e2b.
//
// Solidity: function getOwners() view returns(address[])
func (_Safe *SafeSession) GetOwners() ([]common.Address, error) {
	return _Safe.Contract.GetOwners(&_Safe.CallOpts)
}

// GetOwners is a free data retrieval call binding the contract method 0xa0e67e2b.
//
// Solidity: function getOwners() view returns(address[])
func (_Safe *SafeCallerSession) GetOwners() ([]common.Address, error) {
	return _Safe.Contract.GetOwners(&_Safe.CallOpts)
}

// GetStorageAt is a free data retrieval call binding the contract method 0x5624b25b.
//
// Solidity: function getStorageAt(uint256 offset, uint256 length) view returns(bytes)
func (_Safe *SafeCaller) GetStorageAt(opts *bind.CallOpts, offset *big.Int, length *big.Int) ([]byte, error) {
	var out []interface{}
	err := _Safe.contract.Call(opts, &out, "getStorageAt", offset, length)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetStorageAt is a free data retrieval call binding the contract method 0x5624b25b.
//
// Solidity: function getStorageAt(uint256 offset, uint256 length) view returns(bytes)
func (_Safe *SafeSession) GetStorageAt(offset *big.Int, length *big.Int) ([]byte, error) {
	return _Safe.Contract.GetStorageAt(&_Safe.CallOpts, offset, length)
}

// GetStorageAt is a free data retrieval call binding the contract method 0x5624b25b.
//
// Solidity: function getStorageAt(uint256 offset, uint256 length) view returns(bytes)
func (_Safe *SafeCallerSession) GetStorageAt(offset *big.Int, length *big.Int) ([]byte, error) {
	return _Safe.Contract.GetStorageAt(&_Safe.CallOpts, offset, length)
}

// GetThreshold is a free data retrieval call binding the contract method 0xe75235b8.
//
// Solidity: function getThreshold() view returns(uint256)
func (_Safe *SafeCaller) GetThreshold(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Safe.contract.Call(opts, &out, "getThreshold")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetThreshold is a free data retrieval call binding the contract method 0xe75235b8.
//
// Solidity: function getThreshold() view returns(uint256)
func (_Safe *SafeSession) GetThreshold() (*big.Int, error) {
	return _Safe.Contract.GetThreshold(&_Safe.CallOpts)
}

// GetThreshold is a free data retrieval call binding the contract method 0xe75235b8.
//
// Solidity: function getThreshold() view returns(uint256)
func (_Safe *SafeCallerSession) GetThreshold() (*big.Int, error) {
	return _Safe.Contract.GetThreshold(&_Safe.CallOpts)
}

// GetTransactionHash is a free data retrieval call binding the contract method 0xd8d11f78.
//
// Solidity: function getTransactionHash(address to, uint256 value, bytes data, uint8 operation, uint256 safeTxGas, uint256 baseGas, uint256 gasPrice, address gasToken, address refundReceiver, uint256 _nonce) view returns(bytes32)
func (_Safe *SafeCaller) GetTransactionHash(opts *bind.CallOpts, to common.Address, value *big.Int, data []byte, operation uint8, safeTxGas *big.Int, baseGas *big.Int, gasPrice *big.Int, gasToken common.Address, refundReceiver common.Address, _nonce *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _Safe.contract.Call(opts, &out, "getTransactionHash", to, value, data, operation, safeTxGas, baseGas, gasPrice, gasToken, refundReceiver, _nonce)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetTransactionHash is a free data retrieval call binding the contract method 0xd8d11f78.
//
// Solidity: function getTransactionHash(address to, uint256 value, bytes data, uint8 operation, uint256 safeTxGas, uint256 baseGas, uint256 gasPrice, address gasToken, address refundReceiver, uint256 _nonce) view returns(bytes32)
func (_Safe *SafeSession) GetTransactionHash(to common.Address, value *big.Int, data []byte, operation uint8, safeTxGas *big.Int, baseGas *big.Int, gasPrice *big.Int, gasToken common.Address, refundReceiver common.Address, _nonce *big.Int) ([32]byte, error) {
	return _Safe.Contract.GetTransactionHash(&_Safe.CallOpts, to, value, data, operation, safeTxGas, baseGas, gasPrice, gasToken, refundReceiver, _nonce)
}

// GetTransactionHash is a free data retrieval call binding the contract method 0xd8d11f78.
//
// Solidity: function getTransactionHash(address to, uint256 value, bytes data, uint8 operation, uint256 safeTxGas, uint256 baseGas, uint256 gasPrice, address gasToken, address refundReceiver, uint256 _nonce) view returns(bytes32)
func (_Safe *SafeCallerSession) GetTransactionHash(to common.Address, value *big.Int, data []byte, operation uint8, safeTxGas *big.Int, baseGas *big.Int, gasPrice *big.Int, gasToken common.Address, refundReceiver common.Address, _nonce *big.Int) ([32]byte, error) {
	return _Safe.Contract.GetTransactionHash(&_Safe.CallOpts, to, value, data, operation, safeTxGas, baseGas, gasPrice, gasToken, refundReceiver, _nonce)
}

// IsModuleEnabled is a free data retrieval call binding the contract method 0x2d9ad53d.
//
// Solidity: function isModuleEnabled(address module) view returns(bool)
func (_Safe *SafeCaller) IsModuleEnabled(opts *bind.CallOpts, module common.Address) (bool, error) {
	var out []interface{}
	err := _Safe.contract.Call(opts, &out, "isModuleEnabled", module)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsModuleEnabled is a free data retrieval call binding the contract method 0x2d9ad53d.
//
// Solidity: function isModuleEnabled(address module) view returns(bool)
func (_Safe *SafeSession) IsModuleEnabled(module common.Address) (bool, error) {
	return _Safe.Contract.IsModuleEnabled(&_Safe.CallOpts, module)
}

// IsModuleEnabled is a free data retrieval call binding the contract method 0x2d9ad53d.
//
// Solidity: function isModuleEnabled(address module) view returns(bool)
func (_Safe *SafeCallerSession) IsModuleEnabled(module common.Address) (bool, error) {
	return _Safe.Contract.IsModuleEnabled(&_Safe.CallOpts, module)
}

// IsOwner is a free data retrieval call binding the contract method 0x2f54bf6e.
//
// Solidity: function isOwner(address owner) view returns(bool)
func (_Safe *SafeCaller) IsOwner(opts *bind.CallOpts, owner common.Address) (bool, error) {
	var out []interface{}
	err := _Safe.contract.Call(opts, &out, "isOwner", owner)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOwner is a free data retrieval call binding the contract method 0x2f54bf6e.
//
// Solidity: function isOwner(address owner) view returns(bool)
func (_Safe *SafeSession) IsOwner(owner common.Address) (bool, error) {
	return _Safe.Contract.IsOwner(&_Safe.CallOpts, owner)
}

// IsOwner is a free data retrieval call binding the contract method 0x2f54bf6e.
//
// Solidity: function isOwner(address owner) view returns(bool)
func (_Safe *SafeCallerSession) IsOwner(owner common.Address) (bool, error) {
	return _Safe.Contract.IsOwner(&_Safe.CallOpts, owner)
}

// Nonce is a free data retrieval call binding the contract method 0xaffed0e0.
//
// Solidity: function nonce() view returns(uint256)
func (_Safe *SafeCaller) Nonce(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Safe.contract.Call(opts, &out, "nonce")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Nonce is a free data retrieval call binding the contract method 0xaffed0e0.
//
// Solidity: function nonce() view returns(uint256)
func (_Safe *SafeSession) Nonce() (*big.Int, error) {
	return _Safe.Contract.Nonce(&_Safe.CallOpts)
}

// Nonce is a free data retrieval call binding the contract method 0xaffed0e0.
//
// Solidity: function nonce() view returns(uint256)
func (_Safe *SafeCallerSession) Nonce() (*big.Int, error) {
	return _Safe.Contract.Nonce(&_Safe.CallOpts)
}

// SignedMessages is a free data retrieval call binding the contract method 0x5ae6bd37.
//
// Solidity: function signedMessages(bytes32 ) view returns(uint256)
func (_Safe *SafeCaller) SignedMessages(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Safe.contract.Call(opts, &out, "signedMessages", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SignedMessages is a free data retrieval call binding the contract method 0x5ae6bd37.
//
// Solidity: function signedMessages(bytes32 ) view returns(uint256)
func (_Safe *SafeSession) SignedMessages(arg0 [32]byte) (*big.Int, error) {
	return _Safe.Contract.SignedMessages(&_Safe.CallOpts, arg0)
}

// SignedMessages is a free data retrieval call binding the contract method 0x5ae6bd37.
//
// Solidity: function signedMessages(bytes32 ) view returns(uint256)
func (_Safe *SafeCallerSession) SignedMessages(arg0 [32]byte) (*big.Int, error) {
	return _Safe.Contract.SignedMessages(&_Safe.CallOpts, arg0)
}

// AddOwnerWithThreshold is a paid mutator transaction binding the contract method 0x0d582f13.
//
// Solidity: function addOwnerWithThreshold(address owner, uint256 _threshold) returns()
func (_Safe *SafeTransactor) AddOwnerWithThreshold(opts *bind.TransactOpts, owner common.Address, _threshold *big.Int) (*types.Transaction, error) {
	return _Safe.contract.Transact(opts, "addOwnerWithThreshold", owner, _threshold)
}

// AddOwnerWithThreshold is a paid mutator transaction binding the contract method 0x0d582f13.
//
// Solidity: function addOwnerWithThreshold(address owner, uint256 _threshold) returns()
func (_Safe *SafeSession) AddOwnerWithThreshold(owner common.Address, _threshold *big.Int) (*types.Transaction, error) {
	return _Safe.Contract.AddOwnerWithThreshold(&_Safe.TransactOpts, owner, _threshold)
}

// AddOwnerWithThreshold is a paid mutator transaction binding the contract method 0x0d582f13.
//
// Solidity: function addOwnerWithThreshold(address owner, uint256 _threshold) returns()
func (_Safe *SafeTransactorSession) AddOwnerWithThreshold(owner common.Address, _threshold *big.Int) (*types.Transaction, error) {
	return _Safe.Contract.AddOwnerWithThreshold(&_Safe.TransactOpts, owner, _threshold)
}

// ApproveHash is a paid mutator transaction binding the contract method 0xd4d9bdcd.
//
// Solidity: function approveHash(bytes32 hashToApprove) returns()
func (_Safe *SafeTransactor) ApproveHash(opts *bind.TransactOpts, hashToApprove [32]byte) (*types.Transaction, error) {
	return _Safe.contract.Transact(opts, "approveHash", hashToApprove)
}

// ApproveHash is a paid mutator transaction binding the contract method 0xd4d9bdcd.
//
// Solidity: function approveHash(bytes32 hashToApprove) returns()
func (_Safe *SafeSession) ApproveHash(hashToApprove [32]byte) (*types.Transaction, error) {
	return _Safe.Contract.ApproveHash(&_Safe.TransactOpts, hashToApprove)
}

// ApproveHash is a paid mutator transaction binding the contract method 0xd4d9bdcd.
//
// Solidity: function approveHash(bytes32 hashToApprove) returns()
func (_Safe *SafeTransactorSession) ApproveHash(hashToApprove [32]byte) (*types.Transaction, error) {
	return _Safe.Contract.ApproveHash(&_Safe.TransactOpts, hashToApprove)
}

// ChangeThreshold is a paid mutator transaction binding the contract method 0x694e80c3.
//
// Solidity: function changeThreshold(uint256 _threshold) returns()
func (_Safe *SafeTransactor) ChangeThreshold(opts *bind.TransactOpts, _threshold *big.Int) (*types.Transaction, error) {
	return _Safe.contract.Transact(opts, "changeThreshold", _threshold)
}

// ChangeThreshold is a paid mutator transaction binding the contract method 0x694e80c3.
//
// Solidity: function changeThreshold(uint256 _threshold) returns()
func (_Safe *SafeSession) ChangeThreshold(_threshold *big.Int) (*types.Transaction, error) {
	return _Safe.Contract.ChangeThreshold(&_Safe.TransactOpts, _threshold)
}

// ChangeThreshold is a paid mutator transaction binding the contract method 0x694e80c3.
//
// Solidity: function changeThreshold(uint256 _threshold) returns()
func (_Safe *SafeTransactorSession) ChangeThreshold(_threshold *big.Int) (*types.Transaction, error) {
	return _Safe.Contract.ChangeThreshold(&_Safe.TransactOpts, _threshold)
}

// DisableModule is a paid mutator transaction binding the contract method 0xe009cfde.
//
// Solidity: function disableModule(address prevModule, address module) returns()
func (_Safe *SafeTransactor) DisableModule(opts *bind.TransactOpts, prevModule common.Address, module common.Address) (*types.Transaction, error) {
	return _Safe.contract.Transact(opts, "disableModule", prevModule, module)
}

// DisableModule is a paid mutator transaction binding the contract method 0xe009cfde.
//
// Solidity: function disableModule(address prevModule, address module) returns()
func (_Safe *SafeSession) DisableModule(prevModule common.Address, module common.Address) (*types.Transaction, error) {
	return _Safe.Contract.DisableModule(&_Safe.TransactOpts, prevModule, module)
}

// DisableModule is a paid mutator transaction binding the contract method 0xe009cfde.
//
// Solidity: function disableModule(address prevModule, address module) returns()
func (_Safe *SafeTransactorSession) DisableModule(prevModule common.Address, module common.Address) (*types.Transaction, error) {
	return _Safe.Contract.DisableModule(&_Safe.TransactOpts, prevModule, module)
}

// EnableModule is a paid mutator transaction binding the contract method 0x610b5925.
//
// Solidity: function enableModule(address module) returns()
func (_Safe *SafeTransactor) EnableModule(opts *bind.TransactOpts, module common.Address) (*types.Transaction, error) {
	return _Safe.contract.Transact(opts, "enableModule", module)
}

// EnableModule is a paid mutator transaction binding the contract method 0x610b5925.
//
// Solidity: function enableModule(address module) returns()
func (_Safe *SafeSession) EnableModule(module common.Address) (*types.Transaction, error) {
	return _Safe.Contract.EnableModule(&_Safe.TransactOpts, module)
}

// EnableModule is a paid mutator transaction binding the contract method 0x610b5925.
//
// Solidity: function enableModule(address module) returns()
func (_Safe *SafeTransactorSession) EnableModule(module common.Address) (*types.Transaction, error) {
	return _Safe.Contract.EnableModule(&_Safe.TransactOpts, module)
}

// ExecTransaction is a paid mutator transaction binding the contract method 0x6a761202.
//
// Solidity: function execTransaction(address to, uint256 value, bytes data, uint8 operation, uint256 safeTxGas, uint256 baseGas, uint256 gasPrice, address gasToken, address refundReceiver, bytes signatures) payable returns(bool)
func (_Safe *SafeTransactor) ExecTransaction(opts *bind.TransactOpts, to common.Address, value *big.Int, data []byte, operation uint8, safeTxGas *big.Int, baseGas *big.Int, gasPrice *big.Int, gasToken common.Address, refundReceiver common.Address, signatures []byte) (*types.Transaction, error) {
	return _Safe.contract.Transact(opts, "execTransaction", to, value, data, operation, safeTxGas, baseGas, gasPrice, gasToken, refundReceiver, signatures)
}

// ExecTransaction is a paid mutator transaction binding the contract method 0x6a761202.
//
// Solidity: function execTransaction(address to, uint256 value, bytes data, uint8 operation, uint256 safeTxGas, uint256 baseGas, uint256 gasPrice, address gasToken, address refundReceiver, bytes signatures) payable returns(bool)
func (_Safe *SafeSession) ExecTransaction(to common.Address, value *big.Int, data []byte, operation uint8, safeTxGas *big.Int, baseGas *big.Int, gasPrice *big.Int, gasToken common.Address, refundReceiver common.Address, signatures []byte) (*types.Transaction, error) {
	return _Safe.Contract.ExecTransaction(&_Safe.TransactOpts, to, value, data, operation, safeTxGas, baseGas, gasPrice, gasToken, refundReceiver, signatures)
}

// ExecTransaction is a paid mutator transaction binding the contract method 0x6a761202.
//
// Solidity: function execTransaction(address to, uint256 value, bytes data, uint8 operation, uint256 safeTxGas, uint256 baseGas, uint256 gasPrice, address gasToken, address refundReceiver, bytes signatures) payable returns(bool)
func (_Safe *SafeTransactorSession) ExecTransaction(to common.Address, value *big.Int, data []byte, operation uint8, safeTxGas *big.Int, baseGas *big.Int, gasPrice *big.Int, gasToken common.Address, refundReceiver common.Address, signatures []byte) (*types.Transaction, error) {
	return _Safe.Contract.ExecTransaction(&_Safe.TransactOpts, to, value, data, operation, safeTxGas, baseGas, gasPrice, gasToken, refundReceiver, signatures)
}

// ExecTransactionFromModule is a paid mutator transaction binding the contract method 0x468721a7.
//
// Solidity: function execTransactionFromModule(address to, uint256 value, bytes data, uint8 operation) returns(bool success)
func (_Safe *SafeTransactor) ExecTransactionFromModule(opts *bind.TransactOpts, to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _Safe.contract.Transact(opts, "execTransactionFromModule", to, value, data, operation)
}

// ExecTransactionFromModule is a paid mutator transaction binding the contract method 0x468721a7.
//
// Solidity: function execTransactionFromModule(address to, uint256 value, bytes data, uint8 operation) returns(bool success)
func (_Safe *SafeSession) ExecTransactionFromModule(to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _Safe.Contract.ExecTransactionFromModule(&_Safe.TransactOpts, to, value, data, operation)
}

// ExecTransactionFromModule is a paid mutator transaction binding the contract method 0x468721a7.
//
// Solidity: function execTransactionFromModule(address to, uint256 value, bytes data, uint8 operation) returns(bool success)
func (_Safe *SafeTransactorSession) ExecTransactionFromModule(to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _Safe.Contract.ExecTransactionFromModule(&_Safe.TransactOpts, to, value, data, operation)
}

// ExecTransactionFromModuleReturnData is a paid mutator transaction binding the contract method 0x5229073f.
//
// Solidity: function execTransactionFromModuleReturnData(address to, uint256 value, bytes data, uint8 operation) returns(bool success, bytes returnData)
func (_Safe *SafeTransactor) ExecTransactionFromModuleReturnData(opts *bind.TransactOpts, to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _Safe.contract.Transact(opts, "execTransactionFromModuleReturnData", to, value, data, operation)
}

// ExecTransactionFromModuleReturnData is a paid mutator transaction binding the contract method 0x5229073f.
//
// Solidity: function execTransactionFromModuleReturnData(address to, uint256 value, bytes data, uint8 operation) returns(bool success, bytes returnData)
func (_Safe *SafeSession) ExecTransactionFromModuleReturnData(to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _Safe.Contract.ExecTransactionFromModuleReturnData(&_Safe.TransactOpts, to, value, data, operation)
}

// ExecTransactionFromModuleReturnData is a paid mutator transaction binding the contract method 0x5229073f.
//
// Solidity: function execTransactionFromModuleReturnData(address to, uint256 value, bytes data, uint8 operation) returns(bool success, bytes returnData)
func (_Safe *SafeTransactorSession) ExecTransactionFromModuleReturnData(to common.Address, value *big.Int, data []byte, operation uint8) (*types.Transaction, error) {
	return _Safe.Contract.ExecTransactionFromModuleReturnData(&_Safe.TransactOpts, to, value, data, operation)
}

// RemoveOwner is a paid mutator transaction binding the contract method 0xf8dc5dd9.
//
// Solidity: function removeOwner(address prevOwner, address owner, uint256 _threshold) returns()
func (_Safe *SafeTransactor) RemoveOwner(opts *bind.TransactOpts, prevOwner common.Address, owner common.Address, _threshold *big.Int) (*types.Transaction, error) {
	return _Safe.contract.Transact(opts, "removeOwner", prevOwner, owner, _threshold)
}

// RemoveOwner is a paid mutator transaction binding the contract method 0xf8dc5dd9.
//
// Solidity: function removeOwner(address prevOwner, address owner, uint256 _threshold) returns()
func (_Safe *SafeSession) RemoveOwner(prevOwner common.Address, owner common.Address, _threshold *big.Int) (*types.Transaction, error) {
	return _Safe.Contract.RemoveOwner(&_Safe.TransactOpts, prevOwner, owner, _threshold)
}

// RemoveOwner is a paid mutator transaction binding the contract method 0xf8dc5dd9.
//
// Solidity: function removeOwner(address prevOwner, address owner, uint256 _threshold) returns()
func (_Safe *SafeTransactorSession) RemoveOwner(prevOwner common.Address, owner common.Address, _threshold *big.Int) (*types.Transaction, error) {
	return _Safe.Contract.RemoveOwner(&_Safe.TransactOpts, prevOwner, owner, _threshold)
}

// SetFallbackHandler is a paid mutator transaction binding the contract method 0xf08a0323.
//
// Solidity: function setFallbackHandler(address handler) returns()
func (_Safe *SafeTransactor) SetFallbackHandler(opts *bind.TransactOpts, handler common.Address) (*types.Transaction, error) {
	return _Safe.contract.Transact(opts, "setFallbackHandler", handler)
}

// SetFallbackHandler is a paid mutator transaction binding the contract method 0xf08a0323.
//
// Solidity: function setFallbackHandler(address handler) returns()
func (_Safe *SafeSession) SetFallbackHandler(handler common.Address) (*types.Transaction, error) {
	return _Safe.Contract.SetFallbackHandler(&_Safe.TransactOpts, handler)
}

// SetFallbackHandler is a paid mutator transaction binding the contract method 0xf08a0323.
//
// Solidity: function setFallbackHandler(address handler) returns()
func (_Safe *SafeTransactorSession) SetFallbackHandler(handler common.Address) (*types.Transaction, error) {
	return _Safe.Contract.SetFallbackHandler(&_Safe.TransactOpts, handler)
}

// SetGuard is a paid mutator transaction binding the contract method 0xe19a9dd9.
//
// Solidity: function setGuard(address guard) returns()
func (_Safe *SafeTransactor) SetGuard(opts *bind.TransactOpts, guard common.Address) (*types.Transaction, error) {
	return _Safe.contract.Transact(opts, "setGuard", guard)
}

// SetGuard is a paid mutator transaction binding the contract method 0xe19a9dd9.
//
// Solidity: function setGuard(address guard) returns()
func (_Safe *SafeSession) SetGuard(guard common.Address) (*types.Transaction, error) {
	return _Safe.Contract.SetGuard(&_Safe.TransactOpts, guard)
}

// SetGuard is a paid mutator transaction binding the contract method 0xe19a9dd9.
//
// Solidity: function setGuard(address guard) returns()
func (_Safe *SafeTransactorSession) SetGuard(guard common.Address) (*types.Transaction, error) {
	return _Safe.Contract.SetGuard(&_Safe.TransactOpts, guard)
}

// Setup is a paid mutator transaction binding the contract method 0xb63e800d.
//
// Solidity: function setup(address[] _owners, uint256 _threshold, address to, bytes data, address fallbackHandler, address paymentToken, uint256 payment, address paymentReceiver) returns()
func (_Safe *SafeTransactor) Setup(opts *bind.TransactOpts, _owners []common.Address, _threshold *big.Int, to common.Address, data []byte, fallbackHandler common.Address, paymentToken common.Address, payment *big.Int, paymentReceiver common.Address) (*types.Transaction, error) {
	return _Safe.contract.Transact(opts, "setup", _owners, _threshold, to, data, fallbackHandler, paymentToken, payment, paymentReceiver)
}

// Setup is a paid mutator transaction binding the contract method 0xb63e800d.
//
// Solidity: function setup(address[] _owners, uint256 _threshold, address to, bytes data, address fallbackHandler, address paymentToken, uint256 payment, address paymentReceiver) returns()
func (_Safe *SafeSession) Setup(_owners []common.Address, _threshold *big.Int, to common.Address, data []byte, fallbackHandler common.Address, paymentToken common.Address, payment *big.Int, paymentReceiver common.Address) (*types.Transaction, error) {
	return _Safe.Contract.Setup(&_Safe.TransactOpts, _owners, _threshold, to, data, fallbackHandler, paymentToken, payment, paymentReceiver)
}

// Setup is a paid mutator transaction binding the contract method 0xb63e800d.
//
// Solidity: function setup(address[] _owners, uint256 _threshold, address to, bytes data, address fallbackHandler, address paymentToken, uint256 payment, address paymentReceiver) returns()
func (_Safe *SafeTransactorSession) Setup(_owners []common.Address, _threshold *big.Int, to common.Address, data []byte, fallbackHandler common.Address, paymentToken common.Address, payment *big.Int, paymentReceiver common.Address) (*types.Transaction, error) {
	return _Safe.Contract.Setup(&_Safe.TransactOpts, _owners, _threshold, to, data, fallbackHandler, paymentToken, payment, paymentReceiver)
}

// SimulateAndRevert is a paid mutator transaction binding the contract method 0xb4faba09.
//
// Solidity: function simulateAndRevert(address targetContract, bytes calldataPayload) returns()
func (_Safe *SafeTransactor) SimulateAndRevert(opts *bind.TransactOpts, targetContract common.Address, calldataPayload []byte) (*types.Transaction, error) {
	return _Safe.contract.Transact(opts, "simulateAndRevert", targetContract, calldataPayload)
}

// SimulateAndRevert is a paid mutator transaction binding the contract method 0xb4faba09.
//
// Solidity: function simulateAndRevert(address targetContract, bytes calldataPayload) returns()
func (_Safe *SafeSession) SimulateAndRevert(targetContract common.Address, calldataPayload []byte) (*types.Transaction, error) {
	return _Safe.Contract.SimulateAndRevert(&_Safe.TransactOpts, targetContract, calldataPayload)
}

// SimulateAndRevert is a paid mutator transaction binding the contract method 0xb4faba09.
//
// Solidity: function simulateAndRevert(address targetContract, bytes calldataPayload) returns()
func (_Safe *SafeTransactorSession) SimulateAndRevert(targetContract common.Address, calldataPayload []byte) (*types.Transaction, error) {
	return _Safe.Contract.SimulateAndRevert(&_Safe.TransactOpts, targetContract, calldataPayload)
}

// SwapOwner is a paid mutator transaction binding the contract method 0xe318b52b.
//
// Solidity: function swapOwner(address prevOwner, address oldOwner, address newOwner) returns()
func (_Safe *SafeTransactor) SwapOwner(opts *bind.TransactOpts, prevOwner common.Address, oldOwner common.Address, newOwner common.Address) (*types.Transaction, error) {
	return _Safe.contract.Transact(opts, "swapOwner", prevOwner, oldOwner, newOwner)
}

// SwapOwner is a paid mutator transaction binding the contract method 0xe318b52b.
//
// Solidity: function swapOwner(address prevOwner, address oldOwner, address newOwner) returns()
func (_Safe *SafeSession) SwapOwner(prevOwner common.Address, oldOwner common.Address, newOwner common.Address) (*types.Transaction, error) {
	return _Safe.Contract.SwapOwner(&_Safe.TransactOpts, prevOwner, oldOwner, newOwner)
}

// SwapOwner is a paid mutator transaction binding the contract method 0xe318b52b.
//
// Solidity: function swapOwner(address prevOwner, address oldOwner, address newOwner) returns()
func (_Safe *SafeTransactorSession) SwapOwner(prevOwner common.Address, oldOwner common.Address, newOwner common.Address) (*types.Transaction, error) {
	return _Safe.Contract.SwapOwner(&_Safe.TransactOpts, prevOwner, oldOwner, newOwner)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_Safe *SafeTransactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _Safe.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_Safe *SafeSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Safe.Contract.Fallback(&_Safe.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_Safe *SafeTransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Safe.Contract.Fallback(&_Safe.TransactOpts, calldata)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Safe *SafeTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Safe.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Safe *SafeSession) Receive() (*types.Transaction, error) {
	return _Safe.Contract.Receive(&_Safe.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Safe *SafeTransactorSession) Receive() (*types.Transaction, error) {
	return _Safe.Contract.Receive(&_Safe.TransactOpts)
}

// SafeAddedOwnerIterator is returned from FilterAddedOwner and is used to iterate over the raw logs and unpacked data for AddedOwner events raised by the Safe contract.
type SafeAddedOwnerIterator struct {
	Event *SafeAddedOwner // Event containing the contract specifics and raw log

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
func (it *SafeAddedOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SafeAddedOwner)
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
		it.Event = new(SafeAddedOwner)
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
func (it *SafeAddedOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SafeAddedOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SafeAddedOwner represents a AddedOwner event raised by the Safe contract.
type SafeAddedOwner struct {
	Owner common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterAddedOwner is a free log retrieval operation binding the contract event 0x9465fa0c962cc76958e6373a993326400c1c94f8be2fe3a952adfa7f60b2ea26.
//
// Solidity: event AddedOwner(address indexed owner)
func (_Safe *SafeFilterer) FilterAddedOwner(opts *bind.FilterOpts, owner []common.Address) (*SafeAddedOwnerIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Safe.contract.FilterLogs(opts, "AddedOwner", ownerRule)
	if err != nil {
		return nil, err
	}
	return &SafeAddedOwnerIterator{contract: _Safe.contract, event: "AddedOwner", logs: logs, sub: sub}, nil
}

// WatchAddedOwner is a free log subscription operation binding the contract event 0x9465fa0c962cc76958e6373a993326400c1c94f8be2fe3a952adfa7f60b2ea26.
//
// Solidity: event AddedOwner(address indexed owner)
func (_Safe *SafeFilterer) WatchAddedOwner(opts *bind.WatchOpts, sink chan<- *SafeAddedOwner, owner []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Safe.contract.WatchLogs(opts, "AddedOwner", ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SafeAddedOwner)
				if err := _Safe.contract.UnpackLog(event, "AddedOwner", log); err != nil {
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

// ParseAddedOwner is a log parse operation binding the contract event 0x9465fa0c962cc76958e6373a993326400c1c94f8be2fe3a952adfa7f60b2ea26.
//
// Solidity: event AddedOwner(address indexed owner)
func (_Safe *SafeFilterer) ParseAddedOwner(log types.Log) (*SafeAddedOwner, error) {
	event := new(SafeAddedOwner)
	if err := _Safe.contract.UnpackLog(event, "AddedOwner", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SafeApproveHashIterator is returned from FilterApproveHash and is used to iterate over the raw logs and unpacked data for ApproveHash events raised by the Safe contract.
type SafeApproveHashIterator struct {
	Event *SafeApproveHash // Event containing the contract specifics and raw log

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
func (it *SafeApproveHashIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SafeApproveHash)
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
		it.Event = new(SafeApproveHash)
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
func (it *SafeApproveHashIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SafeApproveHashIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SafeApproveHash represents a ApproveHash event raised by the Safe contract.
type SafeApproveHash struct {
	ApprovedHash [32]byte
	Owner        common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterApproveHash is a free log retrieval operation binding the contract event 0xf2a0eb156472d1440255b0d7c1e19cc07115d1051fe605b0dce69acfec884d9c.
//
// Solidity: event ApproveHash(bytes32 indexed approvedHash, address indexed owner)
func (_Safe *SafeFilterer) FilterApproveHash(opts *bind.FilterOpts, approvedHash [][32]byte, owner []common.Address) (*SafeApproveHashIterator, error) {

	var approvedHashRule []interface{}
	for _, approvedHashItem := range approvedHash {
		approvedHashRule = append(approvedHashRule, approvedHashItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Safe.contract.FilterLogs(opts, "ApproveHash", approvedHashRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &SafeApproveHashIterator{contract: _Safe.contract, event: "ApproveHash", logs: logs, sub: sub}, nil
}

// WatchApproveHash is a free log subscription operation binding the contract event 0xf2a0eb156472d1440255b0d7c1e19cc07115d1051fe605b0dce69acfec884d9c.
//
// Solidity: event ApproveHash(bytes32 indexed approvedHash, address indexed owner)
func (_Safe *SafeFilterer) WatchApproveHash(opts *bind.WatchOpts, sink chan<- *SafeApproveHash, approvedHash [][32]byte, owner []common.Address) (event.Subscription, error) {

	var approvedHashRule []interface{}
	for _, approvedHashItem := range approvedHash {
		approvedHashRule = append(approvedHashRule, approvedHashItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Safe.contract.WatchLogs(opts, "ApproveHash", approvedHashRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SafeApproveHash)
				if err := _Safe.contract.UnpackLog(event, "ApproveHash", log); err != nil {
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

// ParseApproveHash is a log parse operation binding the contract event 0xf2a0eb156472d1440255b0d7c1e19cc07115d1051fe605b0dce69acfec884d9c.
//
// Solidity: event ApproveHash(bytes32 indexed approvedHash, address indexed owner)
func (_Safe *SafeFilterer) ParseApproveHash(log types.Log) (*SafeApproveHash, error) {
	event := new(SafeApproveHash)
	if err := _Safe.contract.UnpackLog(event, "ApproveHash", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SafeChangedFallbackHandlerIterator is returned from FilterChangedFallbackHandler and is used to iterate over the raw logs and unpacked data for ChangedFallbackHandler events raised by the Safe contract.
type SafeChangedFallbackHandlerIterator struct {
	Event *SafeChangedFallbackHandler // Event containing the contract specifics and raw log

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
func (it *SafeChangedFallbackHandlerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SafeChangedFallbackHandler)
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
		it.Event = new(SafeChangedFallbackHandler)
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
func (it *SafeChangedFallbackHandlerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SafeChangedFallbackHandlerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SafeChangedFallbackHandler represents a ChangedFallbackHandler event raised by the Safe contract.
type SafeChangedFallbackHandler struct {
	Handler common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterChangedFallbackHandler is a free log retrieval operation binding the contract event 0x5ac6c46c93c8d0e53714ba3b53db3e7c046da994313d7ed0d192028bc7c228b0.
//
// Solidity: event ChangedFallbackHandler(address indexed handler)
func (_Safe *SafeFilterer) FilterChangedFallbackHandler(opts *bind.FilterOpts, handler []common.Address) (*SafeChangedFallbackHandlerIterator, error) {

	var handlerRule []interface{}
	for _, handlerItem := range handler {
		handlerRule = append(handlerRule, handlerItem)
	}

	logs, sub, err := _Safe.contract.FilterLogs(opts, "ChangedFallbackHandler", handlerRule)
	if err != nil {
		return nil, err
	}
	return &SafeChangedFallbackHandlerIterator{contract: _Safe.contract, event: "ChangedFallbackHandler", logs: logs, sub: sub}, nil
}

// WatchChangedFallbackHandler is a free log subscription operation binding the contract event 0x5ac6c46c93c8d0e53714ba3b53db3e7c046da994313d7ed0d192028bc7c228b0.
//
// Solidity: event ChangedFallbackHandler(address indexed handler)
func (_Safe *SafeFilterer) WatchChangedFallbackHandler(opts *bind.WatchOpts, sink chan<- *SafeChangedFallbackHandler, handler []common.Address) (event.Subscription, error) {

	var handlerRule []interface{}
	for _, handlerItem := range handler {
		handlerRule = append(handlerRule, handlerItem)
	}

	logs, sub, err := _Safe.contract.WatchLogs(opts, "ChangedFallbackHandler", handlerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SafeChangedFallbackHandler)
				if err := _Safe.contract.UnpackLog(event, "ChangedFallbackHandler", log); err != nil {
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

// ParseChangedFallbackHandler is a log parse operation binding the contract event 0x5ac6c46c93c8d0e53714ba3b53db3e7c046da994313d7ed0d192028bc7c228b0.
//
// Solidity: event ChangedFallbackHandler(address indexed handler)
func (_Safe *SafeFilterer) ParseChangedFallbackHandler(log types.Log) (*SafeChangedFallbackHandler, error) {
	event := new(SafeChangedFallbackHandler)
	if err := _Safe.contract.UnpackLog(event, "ChangedFallbackHandler", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SafeChangedGuardIterator is returned from FilterChangedGuard and is used to iterate over the raw logs and unpacked data for ChangedGuard events raised by the Safe contract.
type SafeChangedGuardIterator struct {
	Event *SafeChangedGuard // Event containing the contract specifics and raw log

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
func (it *SafeChangedGuardIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SafeChangedGuard)
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
		it.Event = new(SafeChangedGuard)
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
func (it *SafeChangedGuardIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SafeChangedGuardIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SafeChangedGuard represents a ChangedGuard event raised by the Safe contract.
type SafeChangedGuard struct {
	Guard common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterChangedGuard is a free log retrieval operation binding the contract event 0x1151116914515bc0891ff9047a6cb32cf902546f83066499bcf8ba33d2353fa2.
//
// Solidity: event ChangedGuard(address indexed guard)
func (_Safe *SafeFilterer) FilterChangedGuard(opts *bind.FilterOpts, guard []common.Address) (*SafeChangedGuardIterator, error) {

	var guardRule []interface{}
	for _, guardItem := range guard {
		guardRule = append(guardRule, guardItem)
	}

	logs, sub, err := _Safe.contract.FilterLogs(opts, "ChangedGuard", guardRule)
	if err != nil {
		return nil, err
	}
	return &SafeChangedGuardIterator{contract: _Safe.contract, event: "ChangedGuard", logs: logs, sub: sub}, nil
}

// WatchChangedGuard is a free log subscription operation binding the contract event 0x1151116914515bc0891ff9047a6cb32cf902546f83066499bcf8ba33d2353fa2.
//
// Solidity: event ChangedGuard(address indexed guard)
func (_Safe *SafeFilterer) WatchChangedGuard(opts *bind.WatchOpts, sink chan<- *SafeChangedGuard, guard []common.Address) (event.Subscription, error) {

	var guardRule []interface{}
	for _, guardItem := range guard {
		guardRule = append(guardRule, guardItem)
	}

	logs, sub, err := _Safe.contract.WatchLogs(opts, "ChangedGuard", guardRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SafeChangedGuard)
				if err := _Safe.contract.UnpackLog(event, "ChangedGuard", log); err != nil {
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

// ParseChangedGuard is a log parse operation binding the contract event 0x1151116914515bc0891ff9047a6cb32cf902546f83066499bcf8ba33d2353fa2.
//
// Solidity: event ChangedGuard(address indexed guard)
func (_Safe *SafeFilterer) ParseChangedGuard(log types.Log) (*SafeChangedGuard, error) {
	event := new(SafeChangedGuard)
	if err := _Safe.contract.UnpackLog(event, "ChangedGuard", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SafeChangedThresholdIterator is returned from FilterChangedThreshold and is used to iterate over the raw logs and unpacked data for ChangedThreshold events raised by the Safe contract.
type SafeChangedThresholdIterator struct {
	Event *SafeChangedThreshold // Event containing the contract specifics and raw log

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
func (it *SafeChangedThresholdIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SafeChangedThreshold)
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
		it.Event = new(SafeChangedThreshold)
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
func (it *SafeChangedThresholdIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SafeChangedThresholdIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SafeChangedThreshold represents a ChangedThreshold event raised by the Safe contract.
type SafeChangedThreshold struct {
	Threshold *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterChangedThreshold is a free log retrieval operation binding the contract event 0x610f7ff2b304ae8903c3de74c60c6ab1f7d6226b3f52c5161905bb5ad4039c93.
//
// Solidity: event ChangedThreshold(uint256 threshold)
func (_Safe *SafeFilterer) FilterChangedThreshold(opts *bind.FilterOpts) (*SafeChangedThresholdIterator, error) {

	logs, sub, err := _Safe.contract.FilterLogs(opts, "ChangedThreshold")
	if err != nil {
		return nil, err
	}
	return &SafeChangedThresholdIterator{contract: _Safe.contract, event: "ChangedThreshold", logs: logs, sub: sub}, nil
}

// WatchChangedThreshold is a free log subscription operation binding the contract event 0x610f7ff2b304ae8903c3de74c60c6ab1f7d6226b3f52c5161905bb5ad4039c93.
//
// Solidity: event ChangedThreshold(uint256 threshold)
func (_Safe *SafeFilterer) WatchChangedThreshold(opts *bind.WatchOpts, sink chan<- *SafeChangedThreshold) (event.Subscription, error) {

	logs, sub, err := _Safe.contract.WatchLogs(opts, "ChangedThreshold")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SafeChangedThreshold)
				if err := _Safe.contract.UnpackLog(event, "ChangedThreshold", log); err != nil {
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

// ParseChangedThreshold is a log parse operation binding the contract event 0x610f7ff2b304ae8903c3de74c60c6ab1f7d6226b3f52c5161905bb5ad4039c93.
//
// Solidity: event ChangedThreshold(uint256 threshold)
func (_Safe *SafeFilterer) ParseChangedThreshold(log types.Log) (*SafeChangedThreshold, error) {
	event := new(SafeChangedThreshold)
	if err := _Safe.contract.UnpackLog(event, "ChangedThreshold", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SafeDisabledModuleIterator is returned from FilterDisabledModule and is used to iterate over the raw logs and unpacked data for DisabledModule events raised by the Safe contract.
type SafeDisabledModuleIterator struct {
	Event *SafeDisabledModule // Event containing the contract specifics and raw log

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
func (it *SafeDisabledModuleIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SafeDisabledModule)
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
		it.Event = new(SafeDisabledModule)
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
func (it *SafeDisabledModuleIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SafeDisabledModuleIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SafeDisabledModule represents a DisabledModule event raised by the Safe contract.
type SafeDisabledModule struct {
	Module common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDisabledModule is a free log retrieval operation binding the contract event 0xaab4fa2b463f581b2b32cb3b7e3b704b9ce37cc209b5fb4d77e593ace4054276.
//
// Solidity: event DisabledModule(address indexed module)
func (_Safe *SafeFilterer) FilterDisabledModule(opts *bind.FilterOpts, module []common.Address) (*SafeDisabledModuleIterator, error) {

	var moduleRule []interface{}
	for _, moduleItem := range module {
		moduleRule = append(moduleRule, moduleItem)
	}

	logs, sub, err := _Safe.contract.FilterLogs(opts, "DisabledModule", moduleRule)
	if err != nil {
		return nil, err
	}
	return &SafeDisabledModuleIterator{contract: _Safe.contract, event: "DisabledModule", logs: logs, sub: sub}, nil
}

// WatchDisabledModule is a free log subscription operation binding the contract event 0xaab4fa2b463f581b2b32cb3b7e3b704b9ce37cc209b5fb4d77e593ace4054276.
//
// Solidity: event DisabledModule(address indexed module)
func (_Safe *SafeFilterer) WatchDisabledModule(opts *bind.WatchOpts, sink chan<- *SafeDisabledModule, module []common.Address) (event.Subscription, error) {

	var moduleRule []interface{}
	for _, moduleItem := range module {
		moduleRule = append(moduleRule, moduleItem)
	}

	logs, sub, err := _Safe.contract.WatchLogs(opts, "DisabledModule", moduleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SafeDisabledModule)
				if err := _Safe.contract.UnpackLog(event, "DisabledModule", log); err != nil {
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

// ParseDisabledModule is a log parse operation binding the contract event 0xaab4fa2b463f581b2b32cb3b7e3b704b9ce37cc209b5fb4d77e593ace4054276.
//
// Solidity: event DisabledModule(address indexed module)
func (_Safe *SafeFilterer) ParseDisabledModule(log types.Log) (*SafeDisabledModule, error) {
	event := new(SafeDisabledModule)
	if err := _Safe.contract.UnpackLog(event, "DisabledModule", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SafeEnabledModuleIterator is returned from FilterEnabledModule and is used to iterate over the raw logs and unpacked data for EnabledModule events raised by the Safe contract.
type SafeEnabledModuleIterator struct {
	Event *SafeEnabledModule // Event containing the contract specifics and raw log

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
func (it *SafeEnabledModuleIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SafeEnabledModule)
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
		it.Event = new(SafeEnabledModule)
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
func (it *SafeEnabledModuleIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SafeEnabledModuleIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SafeEnabledModule represents a EnabledModule event raised by the Safe contract.
type SafeEnabledModule struct {
	Module common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterEnabledModule is a free log retrieval operation binding the contract event 0xecdf3a3effea5783a3c4c2140e677577666428d44ed9d474a0b3a4c9943f8440.
//
// Solidity: event EnabledModule(address indexed module)
func (_Safe *SafeFilterer) FilterEnabledModule(opts *bind.FilterOpts, module []common.Address) (*SafeEnabledModuleIterator, error) {

	var moduleRule []interface{}
	for _, moduleItem := range module {
		moduleRule = append(moduleRule, moduleItem)
	}

	logs, sub, err := _Safe.contract.FilterLogs(opts, "EnabledModule", moduleRule)
	if err != nil {
		return nil, err
	}
	return &SafeEnabledModuleIterator{contract: _Safe.contract, event: "EnabledModule", logs: logs, sub: sub}, nil
}

// WatchEnabledModule is a free log subscription operation binding the contract event 0xecdf3a3effea5783a3c4c2140e677577666428d44ed9d474a0b3a4c9943f8440.
//
// Solidity: event EnabledModule(address indexed module)
func (_Safe *SafeFilterer) WatchEnabledModule(opts *bind.WatchOpts, sink chan<- *SafeEnabledModule, module []common.Address) (event.Subscription, error) {

	var moduleRule []interface{}
	for _, moduleItem := range module {
		moduleRule = append(moduleRule, moduleItem)
	}

	logs, sub, err := _Safe.contract.WatchLogs(opts, "EnabledModule", moduleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SafeEnabledModule)
				if err := _Safe.contract.UnpackLog(event, "EnabledModule", log); err != nil {
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

// ParseEnabledModule is a log parse operation binding the contract event 0xecdf3a3effea5783a3c4c2140e677577666428d44ed9d474a0b3a4c9943f8440.
//
// Solidity: event EnabledModule(address indexed module)
func (_Safe *SafeFilterer) ParseEnabledModule(log types.Log) (*SafeEnabledModule, error) {
	event := new(SafeEnabledModule)
	if err := _Safe.contract.UnpackLog(event, "EnabledModule", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SafeExecutionFailureIterator is returned from FilterExecutionFailure and is used to iterate over the raw logs and unpacked data for ExecutionFailure events raised by the Safe contract.
type SafeExecutionFailureIterator struct {
	Event *SafeExecutionFailure // Event containing the contract specifics and raw log

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
func (it *SafeExecutionFailureIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SafeExecutionFailure)
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
		it.Event = new(SafeExecutionFailure)
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
func (it *SafeExecutionFailureIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SafeExecutionFailureIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SafeExecutionFailure represents a ExecutionFailure event raised by the Safe contract.
type SafeExecutionFailure struct {
	TxHash  [32]byte
	Payment *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterExecutionFailure is a free log retrieval operation binding the contract event 0x23428b18acfb3ea64b08dc0c1d296ea9c09702c09083ca5272e64d115b687d23.
//
// Solidity: event ExecutionFailure(bytes32 indexed txHash, uint256 payment)
func (_Safe *SafeFilterer) FilterExecutionFailure(opts *bind.FilterOpts, txHash [][32]byte) (*SafeExecutionFailureIterator, error) {

	var txHashRule []interface{}
	for _, txHashItem := range txHash {
		txHashRule = append(txHashRule, txHashItem)
	}

	logs, sub, err := _Safe.contract.FilterLogs(opts, "ExecutionFailure", txHashRule)
	if err != nil {
		return nil, err
	}
	return &SafeExecutionFailureIterator{contract: _Safe.contract, event: "ExecutionFailure", logs: logs, sub: sub}, nil
}

// WatchExecutionFailure is a free log subscription operation binding the contract event 0x23428b18acfb3ea64b08dc0c1d296ea9c09702c09083ca5272e64d115b687d23.
//
// Solidity: event ExecutionFailure(bytes32 indexed txHash, uint256 payment)
func (_Safe *SafeFilterer) WatchExecutionFailure(opts *bind.WatchOpts, sink chan<- *SafeExecutionFailure, txHash [][32]byte) (event.Subscription, error) {

	var txHashRule []interface{}
	for _, txHashItem := range txHash {
		txHashRule = append(txHashRule, txHashItem)
	}

	logs, sub, err := _Safe.contract.WatchLogs(opts, "ExecutionFailure", txHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SafeExecutionFailure)
				if err := _Safe.contract.UnpackLog(event, "ExecutionFailure", log); err != nil {
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

// ParseExecutionFailure is a log parse operation binding the contract event 0x23428b18acfb3ea64b08dc0c1d296ea9c09702c09083ca5272e64d115b687d23.
//
// Solidity: event ExecutionFailure(bytes32 indexed txHash, uint256 payment)
func (_Safe *SafeFilterer) ParseExecutionFailure(log types.Log) (*SafeExecutionFailure, error) {
	event := new(SafeExecutionFailure)
	if err := _Safe.contract.UnpackLog(event, "ExecutionFailure", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SafeExecutionFromModuleFailureIterator is returned from FilterExecutionFromModuleFailure and is used to iterate over the raw logs and unpacked data for ExecutionFromModuleFailure events raised by the Safe contract.
type SafeExecutionFromModuleFailureIterator struct {
	Event *SafeExecutionFromModuleFailure // Event containing the contract specifics and raw log

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
func (it *SafeExecutionFromModuleFailureIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SafeExecutionFromModuleFailure)
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
		it.Event = new(SafeExecutionFromModuleFailure)
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
func (it *SafeExecutionFromModuleFailureIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SafeExecutionFromModuleFailureIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SafeExecutionFromModuleFailure represents a ExecutionFromModuleFailure event raised by the Safe contract.
type SafeExecutionFromModuleFailure struct {
	Module common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterExecutionFromModuleFailure is a free log retrieval operation binding the contract event 0xacd2c8702804128fdb0db2bb49f6d127dd0181c13fd45dbfe16de0930e2bd375.
//
// Solidity: event ExecutionFromModuleFailure(address indexed module)
func (_Safe *SafeFilterer) FilterExecutionFromModuleFailure(opts *bind.FilterOpts, module []common.Address) (*SafeExecutionFromModuleFailureIterator, error) {

	var moduleRule []interface{}
	for _, moduleItem := range module {
		moduleRule = append(moduleRule, moduleItem)
	}

	logs, sub, err := _Safe.contract.FilterLogs(opts, "ExecutionFromModuleFailure", moduleRule)
	if err != nil {
		return nil, err
	}
	return &SafeExecutionFromModuleFailureIterator{contract: _Safe.contract, event: "ExecutionFromModuleFailure", logs: logs, sub: sub}, nil
}

// WatchExecutionFromModuleFailure is a free log subscription operation binding the contract event 0xacd2c8702804128fdb0db2bb49f6d127dd0181c13fd45dbfe16de0930e2bd375.
//
// Solidity: event ExecutionFromModuleFailure(address indexed module)
func (_Safe *SafeFilterer) WatchExecutionFromModuleFailure(opts *bind.WatchOpts, sink chan<- *SafeExecutionFromModuleFailure, module []common.Address) (event.Subscription, error) {

	var moduleRule []interface{}
	for _, moduleItem := range module {
		moduleRule = append(moduleRule, moduleItem)
	}

	logs, sub, err := _Safe.contract.WatchLogs(opts, "ExecutionFromModuleFailure", moduleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SafeExecutionFromModuleFailure)
				if err := _Safe.contract.UnpackLog(event, "ExecutionFromModuleFailure", log); err != nil {
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

// ParseExecutionFromModuleFailure is a log parse operation binding the contract event 0xacd2c8702804128fdb0db2bb49f6d127dd0181c13fd45dbfe16de0930e2bd375.
//
// Solidity: event ExecutionFromModuleFailure(address indexed module)
func (_Safe *SafeFilterer) ParseExecutionFromModuleFailure(log types.Log) (*SafeExecutionFromModuleFailure, error) {
	event := new(SafeExecutionFromModuleFailure)
	if err := _Safe.contract.UnpackLog(event, "ExecutionFromModuleFailure", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SafeExecutionFromModuleSuccessIterator is returned from FilterExecutionFromModuleSuccess and is used to iterate over the raw logs and unpacked data for ExecutionFromModuleSuccess events raised by the Safe contract.
type SafeExecutionFromModuleSuccessIterator struct {
	Event *SafeExecutionFromModuleSuccess // Event containing the contract specifics and raw log

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
func (it *SafeExecutionFromModuleSuccessIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SafeExecutionFromModuleSuccess)
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
		it.Event = new(SafeExecutionFromModuleSuccess)
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
func (it *SafeExecutionFromModuleSuccessIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SafeExecutionFromModuleSuccessIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SafeExecutionFromModuleSuccess represents a ExecutionFromModuleSuccess event raised by the Safe contract.
type SafeExecutionFromModuleSuccess struct {
	Module common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterExecutionFromModuleSuccess is a free log retrieval operation binding the contract event 0x6895c13664aa4f67288b25d7a21d7aaa34916e355fb9b6fae0a139a9085becb8.
//
// Solidity: event ExecutionFromModuleSuccess(address indexed module)
func (_Safe *SafeFilterer) FilterExecutionFromModuleSuccess(opts *bind.FilterOpts, module []common.Address) (*SafeExecutionFromModuleSuccessIterator, error) {

	var moduleRule []interface{}
	for _, moduleItem := range module {
		moduleRule = append(moduleRule, moduleItem)
	}

	logs, sub, err := _Safe.contract.FilterLogs(opts, "ExecutionFromModuleSuccess", moduleRule)
	if err != nil {
		return nil, err
	}
	return &SafeExecutionFromModuleSuccessIterator{contract: _Safe.contract, event: "ExecutionFromModuleSuccess", logs: logs, sub: sub}, nil
}

// WatchExecutionFromModuleSuccess is a free log subscription operation binding the contract event 0x6895c13664aa4f67288b25d7a21d7aaa34916e355fb9b6fae0a139a9085becb8.
//
// Solidity: event ExecutionFromModuleSuccess(address indexed module)
func (_Safe *SafeFilterer) WatchExecutionFromModuleSuccess(opts *bind.WatchOpts, sink chan<- *SafeExecutionFromModuleSuccess, module []common.Address) (event.Subscription, error) {

	var moduleRule []interface{}
	for _, moduleItem := range module {
		moduleRule = append(moduleRule, moduleItem)
	}

	logs, sub, err := _Safe.contract.WatchLogs(opts, "ExecutionFromModuleSuccess", moduleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SafeExecutionFromModuleSuccess)
				if err := _Safe.contract.UnpackLog(event, "ExecutionFromModuleSuccess", log); err != nil {
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

// ParseExecutionFromModuleSuccess is a log parse operation binding the contract event 0x6895c13664aa4f67288b25d7a21d7aaa34916e355fb9b6fae0a139a9085becb8.
//
// Solidity: event ExecutionFromModuleSuccess(address indexed module)
func (_Safe *SafeFilterer) ParseExecutionFromModuleSuccess(log types.Log) (*SafeExecutionFromModuleSuccess, error) {
	event := new(SafeExecutionFromModuleSuccess)
	if err := _Safe.contract.UnpackLog(event, "ExecutionFromModuleSuccess", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SafeExecutionSuccessIterator is returned from FilterExecutionSuccess and is used to iterate over the raw logs and unpacked data for ExecutionSuccess events raised by the Safe contract.
type SafeExecutionSuccessIterator struct {
	Event *SafeExecutionSuccess // Event containing the contract specifics and raw log

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
func (it *SafeExecutionSuccessIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SafeExecutionSuccess)
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
		it.Event = new(SafeExecutionSuccess)
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
func (it *SafeExecutionSuccessIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SafeExecutionSuccessIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SafeExecutionSuccess represents a ExecutionSuccess event raised by the Safe contract.
type SafeExecutionSuccess struct {
	TxHash  [32]byte
	Payment *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterExecutionSuccess is a free log retrieval operation binding the contract event 0x442e715f626346e8c54381002da614f62bee8d27386535b2521ec8540898556e.
//
// Solidity: event ExecutionSuccess(bytes32 indexed txHash, uint256 payment)
func (_Safe *SafeFilterer) FilterExecutionSuccess(opts *bind.FilterOpts, txHash [][32]byte) (*SafeExecutionSuccessIterator, error) {

	var txHashRule []interface{}
	for _, txHashItem := range txHash {
		txHashRule = append(txHashRule, txHashItem)
	}

	logs, sub, err := _Safe.contract.FilterLogs(opts, "ExecutionSuccess", txHashRule)
	if err != nil {
		return nil, err
	}
	return &SafeExecutionSuccessIterator{contract: _Safe.contract, event: "ExecutionSuccess", logs: logs, sub: sub}, nil
}

// WatchExecutionSuccess is a free log subscription operation binding the contract event 0x442e715f626346e8c54381002da614f62bee8d27386535b2521ec8540898556e.
//
// Solidity: event ExecutionSuccess(bytes32 indexed txHash, uint256 payment)
func (_Safe *SafeFilterer) WatchExecutionSuccess(opts *bind.WatchOpts, sink chan<- *SafeExecutionSuccess, txHash [][32]byte) (event.Subscription, error) {

	var txHashRule []interface{}
	for _, txHashItem := range txHash {
		txHashRule = append(txHashRule, txHashItem)
	}

	logs, sub, err := _Safe.contract.WatchLogs(opts, "ExecutionSuccess", txHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SafeExecutionSuccess)
				if err := _Safe.contract.UnpackLog(event, "ExecutionSuccess", log); err != nil {
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

// ParseExecutionSuccess is a log parse operation binding the contract event 0x442e715f626346e8c54381002da614f62bee8d27386535b2521ec8540898556e.
//
// Solidity: event ExecutionSuccess(bytes32 indexed txHash, uint256 payment)
func (_Safe *SafeFilterer) ParseExecutionSuccess(log types.Log) (*SafeExecutionSuccess, error) {
	event := new(SafeExecutionSuccess)
	if err := _Safe.contract.UnpackLog(event, "ExecutionSuccess", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SafeRemovedOwnerIterator is returned from FilterRemovedOwner and is used to iterate over the raw logs and unpacked data for RemovedOwner events raised by the Safe contract.
type SafeRemovedOwnerIterator struct {
	Event *SafeRemovedOwner // Event containing the contract specifics and raw log

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
func (it *SafeRemovedOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SafeRemovedOwner)
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
		it.Event = new(SafeRemovedOwner)
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
func (it *SafeRemovedOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SafeRemovedOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SafeRemovedOwner represents a RemovedOwner event raised by the Safe contract.
type SafeRemovedOwner struct {
	Owner common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterRemovedOwner is a free log retrieval operation binding the contract event 0xf8d49fc529812e9a7c5c50e69c20f0dccc0db8fa95c98bc58cc9a4f1c1299eaf.
//
// Solidity: event RemovedOwner(address indexed owner)
func (_Safe *SafeFilterer) FilterRemovedOwner(opts *bind.FilterOpts, owner []common.Address) (*SafeRemovedOwnerIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Safe.contract.FilterLogs(opts, "RemovedOwner", ownerRule)
	if err != nil {
		return nil, err
	}
	return &SafeRemovedOwnerIterator{contract: _Safe.contract, event: "RemovedOwner", logs: logs, sub: sub}, nil
}

// WatchRemovedOwner is a free log subscription operation binding the contract event 0xf8d49fc529812e9a7c5c50e69c20f0dccc0db8fa95c98bc58cc9a4f1c1299eaf.
//
// Solidity: event RemovedOwner(address indexed owner)
func (_Safe *SafeFilterer) WatchRemovedOwner(opts *bind.WatchOpts, sink chan<- *SafeRemovedOwner, owner []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Safe.contract.WatchLogs(opts, "RemovedOwner", ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SafeRemovedOwner)
				if err := _Safe.contract.UnpackLog(event, "RemovedOwner", log); err != nil {
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

// ParseRemovedOwner is a log parse operation binding the contract event 0xf8d49fc529812e9a7c5c50e69c20f0dccc0db8fa95c98bc58cc9a4f1c1299eaf.
//
// Solidity: event RemovedOwner(address indexed owner)
func (_Safe *SafeFilterer) ParseRemovedOwner(log types.Log) (*SafeRemovedOwner, error) {
	event := new(SafeRemovedOwner)
	if err := _Safe.contract.UnpackLog(event, "RemovedOwner", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SafeSafeModuleTransactionIterator is returned from FilterSafeModuleTransaction and is used to iterate over the raw logs and unpacked data for SafeModuleTransaction events raised by the Safe contract.
type SafeSafeModuleTransactionIterator struct {
	Event *SafeSafeModuleTransaction // Event containing the contract specifics and raw log

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
func (it *SafeSafeModuleTransactionIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SafeSafeModuleTransaction)
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
		it.Event = new(SafeSafeModuleTransaction)
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
func (it *SafeSafeModuleTransactionIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SafeSafeModuleTransactionIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SafeSafeModuleTransaction represents a SafeModuleTransaction event raised by the Safe contract.
type SafeSafeModuleTransaction struct {
	Module    common.Address
	To        common.Address
	Value     *big.Int
	Data      []byte
	Operation uint8
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterSafeModuleTransaction is a free log retrieval operation binding the contract event 0xb648d3644f584ed1c2232d53c46d87e693586486ad0d1175f8656013110b714e.
//
// Solidity: event SafeModuleTransaction(address module, address to, uint256 value, bytes data, uint8 operation)
func (_Safe *SafeFilterer) FilterSafeModuleTransaction(opts *bind.FilterOpts) (*SafeSafeModuleTransactionIterator, error) {

	logs, sub, err := _Safe.contract.FilterLogs(opts, "SafeModuleTransaction")
	if err != nil {
		return nil, err
	}
	return &SafeSafeModuleTransactionIterator{contract: _Safe.contract, event: "SafeModuleTransaction", logs: logs, sub: sub}, nil
}

// WatchSafeModuleTransaction is a free log subscription operation binding the contract event 0xb648d3644f584ed1c2232d53c46d87e693586486ad0d1175f8656013110b714e.
//
// Solidity: event SafeModuleTransaction(address module, address to, uint256 value, bytes data, uint8 operation)
func (_Safe *SafeFilterer) WatchSafeModuleTransaction(opts *bind.WatchOpts, sink chan<- *SafeSafeModuleTransaction) (event.Subscription, error) {

	logs, sub, err := _Safe.contract.WatchLogs(opts, "SafeModuleTransaction")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SafeSafeModuleTransaction)
				if err := _Safe.contract.UnpackLog(event, "SafeModuleTransaction", log); err != nil {
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

// ParseSafeModuleTransaction is a log parse operation binding the contract event 0xb648d3644f584ed1c2232d53c46d87e693586486ad0d1175f8656013110b714e.
//
// Solidity: event SafeModuleTransaction(address module, address to, uint256 value, bytes data, uint8 operation)
func (_Safe *SafeFilterer) ParseSafeModuleTransaction(log types.Log) (*SafeSafeModuleTransaction, error) {
	event := new(SafeSafeModuleTransaction)
	if err := _Safe.contract.UnpackLog(event, "SafeModuleTransaction", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SafeSafeMultiSigTransactionIterator is returned from FilterSafeMultiSigTransaction and is used to iterate over the raw logs and unpacked data for SafeMultiSigTransaction events raised by the Safe contract.
type SafeSafeMultiSigTransactionIterator struct {
	Event *SafeSafeMultiSigTransaction // Event containing the contract specifics and raw log

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
func (it *SafeSafeMultiSigTransactionIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SafeSafeMultiSigTransaction)
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
		it.Event = new(SafeSafeMultiSigTransaction)
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
func (it *SafeSafeMultiSigTransactionIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SafeSafeMultiSigTransactionIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SafeSafeMultiSigTransaction represents a SafeMultiSigTransaction event raised by the Safe contract.
type SafeSafeMultiSigTransaction struct {
	To             common.Address
	Value          *big.Int
	Data           []byte
	Operation      uint8
	SafeTxGas      *big.Int
	BaseGas        *big.Int
	GasPrice       *big.Int
	GasToken       common.Address
	RefundReceiver common.Address
	Signatures     []byte
	AdditionalInfo []byte
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterSafeMultiSigTransaction is a free log retrieval operation binding the contract event 0x66753cd2356569ee081232e3be8909b950e0a76c1f8460c3a5e3c2be32b11bed.
//
// Solidity: event SafeMultiSigTransaction(address to, uint256 value, bytes data, uint8 operation, uint256 safeTxGas, uint256 baseGas, uint256 gasPrice, address gasToken, address refundReceiver, bytes signatures, bytes additionalInfo)
func (_Safe *SafeFilterer) FilterSafeMultiSigTransaction(opts *bind.FilterOpts) (*SafeSafeMultiSigTransactionIterator, error) {

	logs, sub, err := _Safe.contract.FilterLogs(opts, "SafeMultiSigTransaction")
	if err != nil {
		return nil, err
	}
	return &SafeSafeMultiSigTransactionIterator{contract: _Safe.contract, event: "SafeMultiSigTransaction", logs: logs, sub: sub}, nil
}

// WatchSafeMultiSigTransaction is a free log subscription operation binding the contract event 0x66753cd2356569ee081232e3be8909b950e0a76c1f8460c3a5e3c2be32b11bed.
//
// Solidity: event SafeMultiSigTransaction(address to, uint256 value, bytes data, uint8 operation, uint256 safeTxGas, uint256 baseGas, uint256 gasPrice, address gasToken, address refundReceiver, bytes signatures, bytes additionalInfo)
func (_Safe *SafeFilterer) WatchSafeMultiSigTransaction(opts *bind.WatchOpts, sink chan<- *SafeSafeMultiSigTransaction) (event.Subscription, error) {

	logs, sub, err := _Safe.contract.WatchLogs(opts, "SafeMultiSigTransaction")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SafeSafeMultiSigTransaction)
				if err := _Safe.contract.UnpackLog(event, "SafeMultiSigTransaction", log); err != nil {
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

// ParseSafeMultiSigTransaction is a log parse operation binding the contract event 0x66753cd2356569ee081232e3be8909b950e0a76c1f8460c3a5e3c2be32b11bed.
//
// Solidity: event SafeMultiSigTransaction(address to, uint256 value, bytes data, uint8 operation, uint256 safeTxGas, uint256 baseGas, uint256 gasPrice, address gasToken, address refundReceiver, bytes signatures, bytes additionalInfo)
func (_Safe *SafeFilterer) ParseSafeMultiSigTransaction(log types.Log) (*SafeSafeMultiSigTransaction, error) {
	event := new(SafeSafeMultiSigTransaction)
	if err := _Safe.contract.UnpackLog(event, "SafeMultiSigTransaction", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SafeSafeReceivedIterator is returned from FilterSafeReceived and is used to iterate over the raw logs and unpacked data for SafeReceived events raised by the Safe contract.
type SafeSafeReceivedIterator struct {
	Event *SafeSafeReceived // Event containing the contract specifics and raw log

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
func (it *SafeSafeReceivedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SafeSafeReceived)
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
		it.Event = new(SafeSafeReceived)
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
func (it *SafeSafeReceivedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SafeSafeReceivedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SafeSafeReceived represents a SafeReceived event raised by the Safe contract.
type SafeSafeReceived struct {
	Sender common.Address
	Value  *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterSafeReceived is a free log retrieval operation binding the contract event 0x3d0ce9bfc3ed7d6862dbb28b2dea94561fe714a1b4d019aa8af39730d1ad7c3d.
//
// Solidity: event SafeReceived(address indexed sender, uint256 value)
func (_Safe *SafeFilterer) FilterSafeReceived(opts *bind.FilterOpts, sender []common.Address) (*SafeSafeReceivedIterator, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Safe.contract.FilterLogs(opts, "SafeReceived", senderRule)
	if err != nil {
		return nil, err
	}
	return &SafeSafeReceivedIterator{contract: _Safe.contract, event: "SafeReceived", logs: logs, sub: sub}, nil
}

// WatchSafeReceived is a free log subscription operation binding the contract event 0x3d0ce9bfc3ed7d6862dbb28b2dea94561fe714a1b4d019aa8af39730d1ad7c3d.
//
// Solidity: event SafeReceived(address indexed sender, uint256 value)
func (_Safe *SafeFilterer) WatchSafeReceived(opts *bind.WatchOpts, sink chan<- *SafeSafeReceived, sender []common.Address) (event.Subscription, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Safe.contract.WatchLogs(opts, "SafeReceived", senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SafeSafeReceived)
				if err := _Safe.contract.UnpackLog(event, "SafeReceived", log); err != nil {
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

// ParseSafeReceived is a log parse operation binding the contract event 0x3d0ce9bfc3ed7d6862dbb28b2dea94561fe714a1b4d019aa8af39730d1ad7c3d.
//
// Solidity: event SafeReceived(address indexed sender, uint256 value)
func (_Safe *SafeFilterer) ParseSafeReceived(log types.Log) (*SafeSafeReceived, error) {
	event := new(SafeSafeReceived)
	if err := _Safe.contract.UnpackLog(event, "SafeReceived", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SafeSafeSetupIterator is returned from FilterSafeSetup and is used to iterate over the raw logs and unpacked data for SafeSetup events raised by the Safe contract.
type SafeSafeSetupIterator struct {
	Event *SafeSafeSetup // Event containing the contract specifics and raw log

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
func (it *SafeSafeSetupIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SafeSafeSetup)
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
		it.Event = new(SafeSafeSetup)
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
func (it *SafeSafeSetupIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SafeSafeSetupIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SafeSafeSetup represents a SafeSetup event raised by the Safe contract.
type SafeSafeSetup struct {
	Initiator       common.Address
	Owners          []common.Address
	Threshold       *big.Int
	Initializer     common.Address
	FallbackHandler common.Address
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterSafeSetup is a free log retrieval operation binding the contract event 0x141df868a6331af528e38c83b7aa03edc19be66e37ae67f9285bf4f8e3c6a1a8.
//
// Solidity: event SafeSetup(address indexed initiator, address[] owners, uint256 threshold, address initializer, address fallbackHandler)
func (_Safe *SafeFilterer) FilterSafeSetup(opts *bind.FilterOpts, initiator []common.Address) (*SafeSafeSetupIterator, error) {

	var initiatorRule []interface{}
	for _, initiatorItem := range initiator {
		initiatorRule = append(initiatorRule, initiatorItem)
	}

	logs, sub, err := _Safe.contract.FilterLogs(opts, "SafeSetup", initiatorRule)
	if err != nil {
		return nil, err
	}
	return &SafeSafeSetupIterator{contract: _Safe.contract, event: "SafeSetup", logs: logs, sub: sub}, nil
}

// WatchSafeSetup is a free log subscription operation binding the contract event 0x141df868a6331af528e38c83b7aa03edc19be66e37ae67f9285bf4f8e3c6a1a8.
//
// Solidity: event SafeSetup(address indexed initiator, address[] owners, uint256 threshold, address initializer, address fallbackHandler)
func (_Safe *SafeFilterer) WatchSafeSetup(opts *bind.WatchOpts, sink chan<- *SafeSafeSetup, initiator []common.Address) (event.Subscription, error) {

	var initiatorRule []interface{}
	for _, initiatorItem := range initiator {
		initiatorRule = append(initiatorRule, initiatorItem)
	}

	logs, sub, err := _Safe.contract.WatchLogs(opts, "SafeSetup", initiatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SafeSafeSetup)
				if err := _Safe.contract.UnpackLog(event, "SafeSetup", log); err != nil {
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

// ParseSafeSetup is a log parse operation binding the contract event 0x141df868a6331af528e38c83b7aa03edc19be66e37ae67f9285bf4f8e3c6a1a8.
//
// Solidity: event SafeSetup(address indexed initiator, address[] owners, uint256 threshold, address initializer, address fallbackHandler)
func (_Safe *SafeFilterer) ParseSafeSetup(log types.Log) (*SafeSafeSetup, error) {
	event := new(SafeSafeSetup)
	if err := _Safe.contract.UnpackLog(event, "SafeSetup", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SafeSignMsgIterator is returned from FilterSignMsg and is used to iterate over the raw logs and unpacked data for SignMsg events raised by the Safe contract.
type SafeSignMsgIterator struct {
	Event *SafeSignMsg // Event containing the contract specifics and raw log

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
func (it *SafeSignMsgIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SafeSignMsg)
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
		it.Event = new(SafeSignMsg)
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
func (it *SafeSignMsgIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SafeSignMsgIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SafeSignMsg represents a SignMsg event raised by the Safe contract.
type SafeSignMsg struct {
	MsgHash [32]byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterSignMsg is a free log retrieval operation binding the contract event 0xe7f4675038f4f6034dfcbbb24c4dc08e4ebf10eb9d257d3d02c0f38d122ac6e4.
//
// Solidity: event SignMsg(bytes32 indexed msgHash)
func (_Safe *SafeFilterer) FilterSignMsg(opts *bind.FilterOpts, msgHash [][32]byte) (*SafeSignMsgIterator, error) {

	var msgHashRule []interface{}
	for _, msgHashItem := range msgHash {
		msgHashRule = append(msgHashRule, msgHashItem)
	}

	logs, sub, err := _Safe.contract.FilterLogs(opts, "SignMsg", msgHashRule)
	if err != nil {
		return nil, err
	}
	return &SafeSignMsgIterator{contract: _Safe.contract, event: "SignMsg", logs: logs, sub: sub}, nil
}

// WatchSignMsg is a free log subscription operation binding the contract event 0xe7f4675038f4f6034dfcbbb24c4dc08e4ebf10eb9d257d3d02c0f38d122ac6e4.
//
// Solidity: event SignMsg(bytes32 indexed msgHash)
func (_Safe *SafeFilterer) WatchSignMsg(opts *bind.WatchOpts, sink chan<- *SafeSignMsg, msgHash [][32]byte) (event.Subscription, error) {

	var msgHashRule []interface{}
	for _, msgHashItem := range msgHash {
		msgHashRule = append(msgHashRule, msgHashItem)
	}

	logs, sub, err := _Safe.contract.WatchLogs(opts, "SignMsg", msgHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SafeSignMsg)
				if err := _Safe.contract.UnpackLog(event, "SignMsg", log); err != nil {
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

// ParseSignMsg is a log parse operation binding the contract event 0xe7f4675038f4f6034dfcbbb24c4dc08e4ebf10eb9d257d3d02c0f38d122ac6e4.
//
// Solidity: event SignMsg(bytes32 indexed msgHash)
func (_Safe *SafeFilterer) ParseSignMsg(log types.Log) (*SafeSignMsg, error) {
	event := new(SafeSignMsg)
	if err := _Safe.contract.UnpackLog(event, "SignMsg", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
