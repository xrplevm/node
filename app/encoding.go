package app

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/evm/types"
	legacytypes "github.com/xrplevm/node/v9/legacy/types"
)

// MakeEncodingConfig creates an EncodingConfig for testing
// func MakeEncodingConfig() params.EncodingConfig {
// 	return evmenc.MakeConfig(ModuleBasics)
// }

func (app *App) RegisterLegacyInterfaces(reg codectypes.InterfaceRegistry) {
	reg.RegisterImplementations((*authtypes.AccountI)(nil),
		&types.EthAccount{},       // cosmos/evm
		&legacytypes.EthAccount{}, // evmos
	)
}
