package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	"github.com/xrplevm/node/v12/x/fastupgrade/types"
)

type Keeper struct {
	authority string
	// gov is the concrete keeper, not an interface in expected_keepers.go: the
	// module uses gov's collections (Votes, Proposals, ActiveProposalsQueue,
	// Params), and those are fields, which an interface cannot expose.
	gov     *govkeeper.Keeper
	staking types.StakingKeeper
	upgrade types.UpgradeKeeper

	Params collections.Item[types.Params]
	// Timer is the one running timelock, if any.
	Timer collections.Item[types.Timer]
	// Nominations holds the proposals a bonded validator voted Yes on in this
	// block. It is in the KV store, not a transient store: the EVM gov
	// precompile runs votes on a store that holds only KV stores. EndBlock
	// clears it in every block.
	Nominations collections.KeySet[uint64]
}

// NewKeeper creates the keeper. gov may point at a keeper that is assigned
// later, because the gov keeper needs this keeper's hooks first.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	authority string,
	gov *govkeeper.Keeper,
	staking types.StakingKeeper,
	upgrade types.UpgradeKeeper,
) Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Errorf("invalid fastupgrade authority %q: %w", authority, err))
	}

	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		authority:   authority,
		gov:         gov,
		staking:     staking,
		upgrade:     upgrade,
		Params:      collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		Timer:       collections.NewItem(sb, types.TimerKey, "timer", codec.CollValue[types.Timer](cdc)),
		Nominations: collections.NewKeySet(sb, types.NominationsKey, "nominations", collections.Uint64Key),
	}
	if _, err := sb.Build(); err != nil {
		panic(err)
	}
	return k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// SetParams validates params and stores them. Enabled params must also be
// compatible with gov's current params, so gov can always disable the module.
// MsgUpdateParams and upgrade handlers use it.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if params.Enabled {
		threshold, err := params.ThresholdDec()
		if err != nil {
			return err
		}
		if err := k.checkGovCompatibility(ctx, threshold); err != nil {
			return err
		}
	}
	return k.Params.Set(ctx, params)
}
