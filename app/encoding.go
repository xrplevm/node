package app

import (
	"cosmossdk.io/simapp/params"
	evmenc "github.com/evmos/ethermint/encoding"
)

// MakeEncodingConfig creates an EncodingConfig for testing
func MakeEncodingConfig() params.EncodingConfig {
	return evmenc.MakeConfig(ModuleBasics)
}
