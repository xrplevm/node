package v12

import (
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/authz"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"

	fastupgradetypes "github.com/xrplevm/node/v12/x/fastupgrade/types"
)

const (
	UpgradeName = "v12.0.0"
)

// RetiredStoreKeys are the stores of the modules removed in v12. They stay
// mounted in v12 so the upgrade handler can empty them, and StoreUpgrades.Deleted
// is not used: rootmulti drops a deleted store from CommitInfo, so a later
// restart that still mounts it fails with a version mismatch. Emptied by the
// handler, each store stays in CommitInfo. Remove the mounts in the next
// coordinated upgrade: they are part of the app hash.
var RetiredStoreKeys = []string{
	authz.ModuleName,
	icahosttypes.StoreKey,
	feegrant.StoreKey,
	crisistypes.StoreKey,
	paramstypes.StoreKey,
}

// StoreUpgrades returns the stores v12.0.0 adds.
func StoreUpgrades() storetypes.StoreUpgrades {
	return storetypes.StoreUpgrades{
		Added: []string{fastupgradetypes.StoreKey},
	}
}
