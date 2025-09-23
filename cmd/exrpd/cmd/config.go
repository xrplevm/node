package cmd

import (
	cosmosevmserverconfig "github.com/cosmos/evm/server/config"

	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
)

// InitAppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func InitAppConfig(denom string, evmChainID uint64) (string, interface{}) {
	type CustomAppConfig struct {
		serverconfig.Config

		EVM     cosmosevmserverconfig.EVMConfig
		JSONRPC cosmosevmserverconfig.JSONRPCConfig
		TLS     cosmosevmserverconfig.TLSConfig
	}

	// Optionally allow the chain developer to overwrite the SDK's default
	// server config.
	srvCfg := serverconfig.DefaultConfig()
	// The SDK's default minimum gas price is set to "" (empty value) inside
	// app.toml. If left empty by validators, the node will halt on startup.
	// However, the chain developer can set a default app.toml value for their
	// validators here.
	//
	// In summary:
	// - if you leave srvCfg.MinGasPrices = "", all validators MUST tweak their
	//   own app.toml config,
	// - if you set srvCfg.MinGasPrices non-empty, validators CAN tweak their
	//   own app.toml to override, or use this default value.
	//
	// In this example application, we set the min gas prices to 0.
	srvCfg.MinGasPrices = "0" + denom

	evmCfg := cosmosevmserverconfig.DefaultEVMConfig()
	evmCfg.EVMChainID = evmChainID

	customAppConfig := CustomAppConfig{
		Config:  *srvCfg,
		EVM:     *evmCfg,
		JSONRPC: *cosmosevmserverconfig.DefaultJSONRPCConfig(),
		TLS:     *cosmosevmserverconfig.DefaultTLSConfig(),
	}

	customAppTemplate := serverconfig.DefaultConfigTemplate +
		cosmosevmserverconfig.DefaultEVMConfigTemplate

	return customAppTemplate, customAppConfig
}
