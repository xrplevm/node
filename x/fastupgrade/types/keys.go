package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name.
	ModuleName = "fastupgrade"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName
)

var (
	ParamsKey      = collections.NewPrefix(0)
	TimerKey       = collections.NewPrefix(1)
	NominationsKey = collections.NewPrefix(2)
)
