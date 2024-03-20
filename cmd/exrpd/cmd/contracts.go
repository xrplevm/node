package cmd

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

const (
	witnessInitCoins = "10000000000000000000000" + "token"
	safeInitCoins    = "10000000000000000000000000000000000" + "token"
)

type GenesisContract struct {
	name     string
	address  string
	bytecode string
	memory   evmtypes.Storage
}

type BridgeInitInfo struct {
	lockingAddress  string
	minCreateAmount *big.Int
	signatureReward *big.Int
}

func getGenesisContracts(witnesses []string, threshold int64, bridge *BridgeInitInfo) ([]GenesisContract, error) {
	// Build safe storage
	safeStorage := evmtypes.Storage{
		evmtypes.State{
			Key:   "0x" + padZeroes("0"),
			Value: "0x" + padZeroes(gnosisSafeL2Address),
		},
	}
	modulesStorage, err := buildSentinel([]string{BridgeProxyModuleAddress}, padZeroes("1"))
	if err != nil {
		return nil, err
	}
	safeStorage = append(safeStorage, modulesStorage...)
	witnessesStorage, err := buildSentinel(witnesses, padZeroes("2"))
	if err != nil {
		return nil, err
	}
	safeStorage = append(safeStorage, witnessesStorage...)
	safeStorage = append(safeStorage, evmtypes.State{
		Key:   "0x" + padZeroes("3"),
		Value: "0x" + padZeroes(strconv.FormatInt(int64(len(witnesses)), 16)),
	})
	safeStorage = append(safeStorage, evmtypes.State{
		Key:   "0x" + padZeroes("4"),
		Value: "0x" + padZeroes(strconv.FormatInt(threshold, 16)),
	})
	safeStorage = append(safeStorage, evmtypes.State{
		Key:   "0x6c9a6c4a39284e37ed1cf53d337577d14212a4870fb976a4366c693b939918d5",
		Value: "0x" + padZeroes(fallbackHandlerAddress),
	})

	bridgeProxyStorage := evmtypes.Storage{
		evmtypes.State{
			Key:   "0x" + padZeroes("0"),
			Value: "0x" + padZeroes("1"),
		},
		evmtypes.State{
			Key:   "0x" + padZeroes("97"),
			Value: "0x" + padZeroes(SafeProxyAddress),
		},
		evmtypes.State{
			Key:   "0x" + padZeroes("fd"),
			Value: "0x" + padZeroes(SafeProxyAddress),
		},
		evmtypes.State{
			Key:   "0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc",
			Value: "0x" + padZeroes(bridgeDoorMultiTokenAddress),
		},
	}

	if bridge != nil {
		// Array of bridge keys size
		bridgeKeysArrMem := padZeroes("fc")
		bridgeProxyStorage = append(bridgeProxyStorage, evmtypes.State{
			Key:   "0x" + bridgeKeysArrMem,
			Value: "0x" + padZeroes("1"),
		})

		// Bridge hash save
		bridgeKeysArrMemBytes, err := hex.DecodeString(bridgeKeysArrMem)
		if err != nil {
			return nil, err
		}
		encodedBridge, err := hex.DecodeString(getBridgeEncoding(bridge))
		if err != nil {
			return nil, err
		}
		bridgeKeyBytes := crypto.Keccak256(encodedBridge)
		bridgeProxyStorage = append(bridgeProxyStorage, evmtypes.State{
			Key:   "0x" + hex.EncodeToString(crypto.Keccak256(bridgeKeysArrMemBytes)),
			Value: "0x" + hex.EncodeToString(bridgeKeyBytes),
		})

		// Compute bridge initial memory
		baseSlot, err := hex.DecodeString(padZeroes("fb")) // Map memory slot
		if err != nil {
			return nil, err
		}

		startMem := hex.EncodeToString(crypto.Keccak256(append(bridgeKeyBytes, baseSlot...)))
		memBI, success := big.NewInt(0).SetString(startMem, 16)
		if !success {
			return nil, fmt.Errorf("error setting hex string to big int")
		}

		// Bridge lock value
		bridgeProxyStorage = append(bridgeProxyStorage, evmtypes.State{
			Key:   "0x" + padZeroes(fmt.Sprintf("%x", memBI)),
			Value: "0x" + padZeroes("1"),
		})
		memBI = memBI.Add(memBI, big.NewInt(1))

		// Bridge minCreateAmount
		bridgeProxyStorage = append(bridgeProxyStorage, evmtypes.State{
			Key:   "0x" + padZeroes(fmt.Sprintf("%x", memBI)),
			Value: "0x" + padZeroes(fmt.Sprintf("%x", bridge.minCreateAmount)),
		})
		memBI = memBI.Add(memBI, big.NewInt(1))

		// Bridge signatureReward
		bridgeProxyStorage = append(bridgeProxyStorage, evmtypes.State{
			Key:   "0x" + padZeroes(fmt.Sprintf("%x", memBI)),
			Value: "0x" + padZeroes(fmt.Sprintf("%x", bridge.signatureReward)),
		})
		memBI = memBI.Add(memBI, big.NewInt(1))

		// Bridge lockingAddress
		bridgeProxyStorage = append(bridgeProxyStorage, evmtypes.State{
			Key:   "0x" + padZeroes(fmt.Sprintf("%x", memBI)),
			Value: "0x" + padZeroes(bridge.lockingAddress),
		})
		memBI = memBI.Add(memBI, big.NewInt(2)) // Sum 2 as lockingIssuer is 0 and not present

		// Bridge lockingIssue currency (XRP in hex + size (6))
		bridgeProxyStorage = append(bridgeProxyStorage, evmtypes.State{
			Key:   "0x" + padZeroes(fmt.Sprintf("%x", memBI)),
			Value: "0x5852500000000000000000000000000000000000000000000000000000000006",
		})
		memBI = memBI.Add(memBI, big.NewInt(1))

		// Bridge issuingAddress
		bridgeProxyStorage = append(bridgeProxyStorage, evmtypes.State{
			Key:   "0x" + padZeroes(fmt.Sprintf("%x", memBI)),
			Value: "0x" + padZeroes(SafeProxyAddress),
		})
		memBI = memBI.Add(memBI, big.NewInt(2)) // Sum 2 as issuingIssuer is 0 and not present

		// Bridge issuingIssue currency (XRP in hex + size (6))
		bridgeProxyStorage = append(bridgeProxyStorage, evmtypes.State{
			Key:   "0x" + padZeroes(fmt.Sprintf("%x", memBI)),
			Value: "0x5852500000000000000000000000000000000000000000000000000000000006",
		})

	}

	return append([]GenesisContract{}, GenesisContract{
		name:     deployerContractName,
		address:  deployerContractAddress,
		bytecode: deployerContractBytecode,
		memory:   evmtypes.Storage{},
	}, GenesisContract{
		name:     simulateTxAccessorName,
		address:  simulateTxAccessorAddress,
		bytecode: simulateTxAccessorBytecode,
		memory:   evmtypes.Storage{},
	}, GenesisContract{
		name:     GnosisSafeProxyFactoryName,
		address:  GnosisSafeProxyFactoryAddress,
		bytecode: GnosisSafeProxyFactoryBytecode,
		memory:   evmtypes.Storage{},
	}, GenesisContract{
		name:     defaultCallbackHandlerName,
		address:  defaultCallbackHandlerAddress,
		bytecode: defaultCallbackHandlerBytecode,
		memory:   evmtypes.Storage{},
	}, GenesisContract{
		name:     fallbackHandlerName,
		address:  fallbackHandlerAddress,
		bytecode: fallbackHandlerBytecode,
		memory:   evmtypes.Storage{},
	}, GenesisContract{
		name:     createCallName,
		address:  createCallAddress,
		bytecode: createCallBytecode,
		memory:   evmtypes.Storage{},
	}, GenesisContract{
		name:     multiSendName,
		address:  multiSendAddress,
		bytecode: multiSendBytecode,
		memory:   evmtypes.Storage{},
	}, GenesisContract{
		name:     multiSendCallOnlyName,
		address:  multiSendCallOnlyAddress,
		bytecode: multiSendCallOnlyBytecode,
		memory:   evmtypes.Storage{},
	}, GenesisContract{
		name:     gnosisSafeL2Name,
		address:  gnosisSafeL2Address,
		bytecode: gnosisSafeL2Bytecode,
		memory: evmtypes.Storage{
			evmtypes.State{
				Key:   "0x0000000000000000000000000000000000000000000000000000000000000004",
				Value: "0x0000000000000000000000000000000000000000000000000000000000000001",
			},
		},
	}, GenesisContract{
		name:     safeProxyName,
		address:  SafeProxyAddress,
		bytecode: safeProxyBytecode,
		memory:   safeStorage,
	}, GenesisContract{
		name:     UtilsName,
		address:  UtilsAddress,
		bytecode: UtilsBytecode,
		memory:   evmtypes.Storage{},
	}, GenesisContract{
		name:     XChainUtilsName,
		address:  XChainUtilsAddress,
		bytecode: XChainUtilsBytecode,
		memory:   evmtypes.Storage{},
	}, GenesisContract{
		name:     bridgeDoorMultiTokenName,
		address:  bridgeDoorMultiTokenAddress,
		bytecode: bridgeDoorMultiTokenBytecode,
		memory: evmtypes.Storage{
			evmtypes.State{
				Key:   "0x" + padZeroes("0"),
				Value: "0x" + padZeroes("ff"),
			},
		},
	}, GenesisContract{
		name:     BridgeProxyModuleName,
		address:  BridgeProxyModuleAddress,
		bytecode: BridgeProxyModuleBytecode,
		memory:   bridgeProxyStorage,
	}), nil
}

func buildSentinel(addresses []string, slot string) ([]evmtypes.State, error) {
	baseSlot, err := hex.DecodeString(slot)
	if err != nil {
		return nil, err
	}
	entries := make([]evmtypes.State, len(addresses)+1)

	sentinel := "0000000000000000000000000000000000000001"
	key := sentinel
	for i, addr := range addresses {
		entry, err := getEntryFromValues(baseSlot, key, addr)
		if err != nil {
			return nil, err
		}

		entries[i] = *entry
		key = addr
	}

	entry, err := getEntryFromValues(baseSlot, key, sentinel)
	if err != nil {
		return nil, err
	}
	entries[len(addresses)] = *entry

	return entries, nil
}

func getEntryFromValues(slot []byte, key, addr string) (*evmtypes.State, error) {
	paddedKey, err := hex.DecodeString(padZeroes(key))
	if err != nil {
		return nil, err
	}

	derivedKey := common.BytesToHash(crypto.Keccak256(append(paddedKey, slot...))).Hex()
	entry := evmtypes.State{
		Key:   derivedKey,
		Value: "0x" + padZeroes(addr),
	}

	return &entry, nil
}

func padZeroes(input string) string {
	for len(input) < 64 {
		input = "0" + input
	}
	return input
}

func padZeroesRight(input string) string {
	for len(input) < 64 {
		input += "0"
	}
	return input
}

func getBridgeEncoding(bridge *BridgeInitInfo) string {
	currency := "XRP"
	currencyEncoded := []byte(currency)
	encoding := padZeroes(bridge.lockingAddress)                              // Locking address
	encoding += padZeroes("0")                                                // Locking issuer
	encoding += padZeroes("c0")                                               // c0 = 192 => bytes before string start
	encoding += padZeroes(SafeProxyAddress)                                   // Issuing address
	encoding += padZeroes("0")                                                // Issuing issuer
	encoding += padZeroes("100")                                              // c0 = 256 => bytes before string start
	encoding += padZeroes(strconv.FormatInt(int64(len(currencyEncoded)), 16)) // Position 192, first string, indicates size
	encoding += padZeroesRight(hex.EncodeToString(currencyEncoded))           // String encoded
	encoding += padZeroes(strconv.FormatInt(int64(len(currencyEncoded)), 16)) // Position 256, second string, indicates size
	encoding += padZeroesRight(hex.EncodeToString(currencyEncoded))           // String encoded

	return encoding
}
