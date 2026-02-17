package cmd

import (
	"fmt"
	"strconv"
	"strings"

	clienthelpers "cosmossdk.io/client/v2/helpers"
)

// CosmosChainIDToEvmChainID extracts the EVM chain ID from a Cosmos chain ID.
// Expected format: {alphanumeric}_{number}-{revision}
// Example: "xrplevm_12345-1" returns 12345
func CosmosChainIDToEvmChainID(cosmosChainID string) (uint64, error) {
	parts := strings.Split(cosmosChainID, "_")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid chain ID format: expected format {name}_{number}-{revision}")
	}

	numberPart := strings.Split(parts[1], "-")
	if len(numberPart) != 2 {
		return 0, fmt.Errorf("invalid chain ID format: expected format {name}_{number}-{revision}")
	}

	evmChainID, err := strconv.ParseUint(numberPart[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse EVM chain ID: %w", err)
	}

	return evmChainID, nil
}

func MustGetDefaultNodeHome() string {
	defaultNodeHome, err := clienthelpers.GetNodeHomeDirectory(".exrpd")
	if err != nil {
		panic(err)
	}
	return defaultNodeHome
}
