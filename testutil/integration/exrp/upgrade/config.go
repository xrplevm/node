// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package exrpupgrade

import (
	"os"

	exrpcommon "github.com/xrplevm/node/v4/testutil/integration/exrp/common"
)

const (
	ChainID = "exrp_1440002-1"
)

// Config defines the configuration for a chain.
// It allows for customization of the network to adjust to
// testing needs.
type UpgradeConfig exrpcommon.Config 

func DefaultUpgradeConfig() UpgradeConfig {
	return UpgradeConfig(exrpcommon.DefaultConfig())
}

func (cfg UpgradeConfig) Config() *exrpcommon.Config {
	rootCfg := exrpcommon.Config(cfg)
	return &rootCfg
}

type UpgradeConfigOption func(cfg *UpgradeConfig)

// WithGenesisFile sets the genesis file for the network.
func WithGenesisFile(genesisFile string) UpgradeConfigOption {
	return func(cfg *UpgradeConfig) {
		genesisBytes, err := os.ReadFile(genesisFile)
		if err != nil {
			panic(err)
		}
		cfg.GenesisBytes = genesisBytes
	}
}
