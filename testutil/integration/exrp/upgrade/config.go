// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package exrpupgrade

import (
	"os"

	exrpcommon "github.com/xrplevm/node/v7/testutil/integration/exrp/common"
)

func DefaultUpgradeConfig() exrpcommon.Config {
	return exrpcommon.DefaultConfig()
}

// WithGenesisFile sets the genesis file for the network.
func WithGenesisFile(genesisFile string) exrpcommon.ConfigOption {
	return func(cfg *exrpcommon.Config) {
		genesisBytes, err := os.ReadFile(genesisFile)
		if err != nil {
			panic(err)
		}
		cfg.GenesisBytes = genesisBytes
	}
}
