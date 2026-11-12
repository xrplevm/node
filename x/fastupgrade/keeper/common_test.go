package keeper_test

import (
	"context"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/header"
	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/evm/testutil/keyring"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/stretchr/testify/require"

	"github.com/xrplevm/node/v12/app"
	exrpcommon "github.com/xrplevm/node/v12/testutil/integration/exrp/common"
	exrpintegration "github.com/xrplevm/node/v12/testutil/integration/exrp/integration"
	"github.com/xrplevm/node/v12/x/fastupgrade/keeper"
	"github.com/xrplevm/node/v12/x/fastupgrade/types"
)

const (
	// incompatibleGovThreshold exceeds the 0.85 fast-path threshold.
	incompatibleGovThreshold = "0.9"

	numValidators  = 5
	originalHeight = int64(1_000_000)
	startHeight    = int64(100)

	// upgradeHeight is H of a timer that starts at startHeight with the
	// default timelock.
	upgradeHeight = startHeight + int64(types.DefaultTimelockBlocks)

	// firstProposalID is the id of a test's first proposal, as event text.
	firstProposalID = "1"

	// The tests use their own versions, so they do not depend on the upgrade
	// handler of the node version. baseVersion is the completed upgrade of
	// newFixture, and minorVersion is its minor bump.
	baseVersion  = "v1.0.0"
	minorVersion = "v1.1.0"
)

// upgradeHeightAttr is upgradeHeight as event text.
var upgradeHeightAttr = strconv.FormatInt(upgradeHeight, 10)

// TestMain runs one time for the package, before all tests. It sets the
// address prefixes of the chain.
func TestMain(m *testing.M) {
	exrpcommon.SetupSdkConfig()
	os.Exit(m.Run())
}

// fixture runs keeper and gov code on a real app, with the real gov, staking,
// and x/upgrade keepers, not with mocks. Most of the module's logic is its
// interaction with those modules, for example gov's tally and vote hook. A
// mock would return only what the test told it, so it would prove nothing.
//
// Blocks are simulated on one context: nextBlock bumps height and time.
// tests/integration covers real block commits.
type fixture struct {
	t       *testing.T
	app     *app.App
	ctx     sdk.Context
	k       keeper.Keeper
	vals    []sdk.AccAddress
	other   sdk.AccAddress
	govAddr string
}

// newFixture has the module enabled with its default params, and baseVersion
// as the last completed upgrade.
func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := newFixtureWithoutBaseVersion(t)
	f.applyUpgrade(baseVersion)
	return f
}

// newFixtureWithoutBaseVersion has the module enabled with its default params,
// and no completed upgrade.
func newFixtureWithoutBaseVersion(t *testing.T) *fixture {
	t.Helper()
	kr := keyring.New(numValidators + 1)
	operators := kr.GetAllAccAddrs()[:numValidators]
	net := exrpintegration.NewUnitTestNetwork(
		exrpcommon.WithPreFundedAccounts(kr.GetAllAccAddrs()...),
		exrpcommon.WithAmountOfValidators(numValidators),
		exrpcommon.WithValidatorOperators(operators),
		exrpcommon.WithBondDenom(app.BaseDenom),
		exrpcommon.WithMaxValidators(10),
		exrpcommon.WithMinDepositAmt(sdkmath.NewInt(1)),
		exrpcommon.WithCustomGenesis(exrpcommon.CustomGenesisState{evmtypes.ModuleName: evmGenesis()}),
	)
	f := &fixture{
		t:       t,
		app:     net.App,
		k:       net.App.FastUpgradeKeeper,
		vals:    operators,
		other:   kr.GetAccAddr(numValidators),
		govAddr: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	}
	f.ctx = atBlock(net.GetContext(), startHeight, net.GetContext().BlockTime())
	f.setParams(true, types.DefaultTimelockBlocks, types.DefaultThreshold.String())
	return f
}

func atBlock(ctx sdk.Context, height int64, t time.Time) sdk.Context {
	return ctx.
		WithBlockHeight(height).
		WithBlockTime(t).
		WithHeaderInfo(header.Info{Height: height, Time: t}).
		WithEventManager(sdk.NewEventManager())
}

func (f *fixture) setBlock(height int64, t time.Time) {
	f.ctx = atBlock(f.ctx, height, t)
}

func (f *fixture) nextBlock() {
	f.setBlock(f.ctx.BlockHeight()+1, f.ctx.BlockTime().Add(time.Second))
}

// advanceTo moves to the given height, one block and one second at a time.
// It does not run EndBlock for the blocks it passes: call f.endBlock() where
// the test needs the module to run.
func (f *fixture) advanceTo(height int64) {
	for f.ctx.BlockHeight() < height {
		f.nextBlock()
	}
}

// evmGenesis is only wiring for the test app. The EVM module's default gas
// coin is aatom, and the app cannot start with a coin it does not have.
func evmGenesis() *evmtypes.GenesisState {
	gen := evmtypes.DefaultGenesisState()
	gen.Params.EvmDenom = app.BaseDenom
	return gen
}

// applyUpgrade marks name completed at the current height through x/upgrade.
// A name with no handler in the app gets a no-op test handler.
func (f *fixture) applyUpgrade(name string) {
	f.t.Helper()
	if !f.app.UpgradeKeeper.HasHandler(name) {
		f.app.UpgradeKeeper.SetUpgradeHandler(name, func(_ context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
			return vm, nil
		})
	}
	require.NoError(f.t, f.app.UpgradeKeeper.ApplyUpgrade(f.ctx, upgradetypes.Plan{Name: name, Height: f.ctx.BlockHeight()}))
}

// completeUpgrades writes an upgrade history in the past. It goes back from
// startHeight to blocks 10, 20, and so on, records one upgrade at each, and
// then returns to startHeight. The last name is the last completed upgrade.
//
// Example: completeUpgrades("v1.0.0", "v1.1.0") at block 100 records v1.0.0
// at block 10 and v1.1.0 at block 20, and the test continues at block 100.
func (f *fixture) completeUpgrades(names ...string) {
	f.t.Helper()
	for i, name := range names {
		f.setBlock(int64(10*(i+1)), f.ctx.BlockTime())
		f.applyUpgrade(name)
	}
	f.setBlock(startHeight, f.ctx.BlockTime())
}

func (f *fixture) upgradeMsg(name string) *upgradetypes.MsgSoftwareUpgrade {
	return &upgradetypes.MsgSoftwareUpgrade{
		Authority: f.govAddr,
		Plan:      upgradetypes.Plan{Name: name, Height: originalHeight, Info: "info-" + name},
	}
}

func (f *fixture) propose(msgs ...sdk.Msg) uint64 {
	f.t.Helper()
	p, err := f.app.GovKeeper.SubmitProposal(f.ctx, msgs, "", "title", "summary", f.other, false)
	require.NoError(f.t, err)
	require.NoError(f.t, f.app.GovKeeper.ActivateVotingPeriod(f.ctx, p))
	return p.Id
}

func (f *fixture) proposeMinor() uint64 {
	return f.propose(f.upgradeMsg(minorVersion))
}

func (f *fixture) votes(id uint64, n int, options govv1.WeightedVoteOptions) {
	for _, v := range f.vals[:n] {
		f.vote(id, v, options)
	}
}

func (f *fixture) vote(id uint64, voter sdk.AccAddress, options govv1.WeightedVoteOptions) {
	f.t.Helper()
	require.NoError(f.t, f.app.GovKeeper.AddVote(f.ctx, id, voter, options, ""))
}

func yes() govv1.WeightedVoteOptions { return govv1.NewNonSplitVoteOption(govv1.OptionYes) }
func no() govv1.WeightedVoteOptions  { return govv1.NewNonSplitVoteOption(govv1.OptionNo) }

func split(yesWeight, noWeight string) govv1.WeightedVoteOptions {
	return govv1.WeightedVoteOptions{
		{Option: govv1.OptionYes, Weight: yesWeight},
		{Option: govv1.OptionNo, Weight: noWeight},
	}
}

func (f *fixture) nominated(id uint64) bool {
	has, err := f.k.Nominations.Has(f.ctx, id)
	require.NoError(f.t, err)
	return has
}

// fullEndBlock runs our EndBlocker and then gov's, the app's order.
func (f *fixture) fullEndBlock() {
	f.t.Helper()
	f.endBlock()
	require.NoError(f.t, gov.EndBlocker(f.ctx, &f.app.GovKeeper))
}

// endBlock runs the module's EndBlock on demand, at the height and time of
// f.ctx.
func (f *fixture) endBlock() {
	f.t.Helper()
	require.NoError(f.t, f.k.EndBlock(f.ctx))
}

// timer returns the current timer. The second value is false when no timer
// is running.
func (f *fixture) timer() (types.Timer, bool) {
	t, err := f.k.Timer.Get(f.ctx)
	if errors.Is(err, collections.ErrNotFound) {
		return types.Timer{}, false
	}
	require.NoError(f.t, err)
	return t, true
}

func (f *fixture) setParams(enabled bool, timelock uint64, threshold string) {
	f.t.Helper()
	_, err := keeper.NewMsgServerImpl(f.k).UpdateParams(f.ctx, &types.MsgUpdateParams{
		Authority: f.govAddr,
		Params:    types.Params{Enabled: enabled, TimelockBlocks: timelock, Threshold: threshold, PostponeBlocks: types.DefaultPostponeBlocks},
	})
	require.NoError(f.t, err)
}

func (f *fixture) setGovParams(mutate func(*govv1.Params)) {
	f.t.Helper()
	p, err := f.app.GovKeeper.Params.Get(f.ctx)
	require.NoError(f.t, err)
	mutate(&p)
	require.NoError(f.t, f.app.GovKeeper.Params.Set(f.ctx, p))
}

// proposal fetches the proposal from gov's store. It fails the test when the
// proposal does not exist.
func (f *fixture) proposal(id uint64) govv1.Proposal {
	p, err := f.app.GovKeeper.Proposals.Get(f.ctx, id)
	require.NoError(f.t, err)
	return p
}

// storedVotes returns how many votes gov stores for the proposal with this
// id.
func (f *fixture) storedVotes(id uint64) int {
	iter, err := f.app.GovKeeper.Votes.Iterate(f.ctx, collections.NewPrefixedPairRange[uint64, sdk.AccAddress](id))
	require.NoError(f.t, err)
	keys, err := iter.Keys()
	require.NoError(f.t, err)
	return len(keys)
}

// queued reports whether gov's active queue has the proposal with this voting
// end time.
func (f *fixture) queued(endTime time.Time, id uint64) bool {
	has, err := f.app.GovKeeper.ActiveProposalsQueue.Has(f.ctx, collections.Join(endTime, id))
	require.NoError(f.t, err)
	return has
}

// planHeight returns the plan height in the proposal's upgrade message.
func (f *fixture) planHeight(p govv1.Proposal) int64 {
	msg, ok := p.Messages[0].GetCachedValue().(*upgradetypes.MsgSoftwareUpgrade)
	require.True(f.t, ok)
	return msg.Plan.Height
}

// jail jails the validator and runs staking's EndBlocker, so it leaves the
// bonded set and the bonded pool.
func (f *fixture) jail(operator sdk.AccAddress) {
	f.t.Helper()
	val, err := f.app.StakingKeeper.GetValidator(f.ctx, sdk.ValAddress(operator))
	require.NoError(f.t, err)
	cons, err := val.GetConsAddr()
	require.NoError(f.t, err)
	require.NoError(f.t, f.app.StakingKeeper.Jail(f.ctx, cons))
	_, err = f.app.StakingKeeper.EndBlocker(f.ctx)
	require.NoError(f.t, err)
}

// delegate bonds amount from delegator to the validator. The delegator then
// has voting power of its own in gov's tally.
func (f *fixture) delegate(delegator sdk.AccAddress, validator sdk.AccAddress, amount sdkmath.Int) {
	f.t.Helper()
	val, err := f.app.StakingKeeper.GetValidator(f.ctx, sdk.ValAddress(validator))
	require.NoError(f.t, err)
	_, err = f.app.StakingKeeper.Delegate(f.ctx, delegator, amount, stakingtypes.Unbonded, val, true)
	require.NoError(f.t, err)
}

// events returns the events of this type in the current block, each as a map
// of its attributes.
func (f *fixture) events(eventType string) []map[string]string {
	all := f.ctx.EventManager().Events()
	out := make([]map[string]string, 0, len(all))
	for _, e := range all {
		if e.Type != eventType {
			continue
		}
		attrs := map[string]string{}
		for _, a := range e.Attributes {
			attrs[a.Key] = a.Value
		}
		out = append(out, attrs)
	}
	return out
}

// fastUpgradeEvents returns how many events of the module the current block
// has, over all five event types.
func (f *fixture) fastUpgradeEvents() int {
	n := 0
	for _, et := range []string{types.EventTypeStarted, types.EventTypeNotStarted, types.EventTypeDropped, types.EventTypePostponed, types.EventTypeEarlyClose} {
		n += len(f.events(et))
	}
	return n
}

func (f *fixture) requireSingleEvent(eventType string, want map[string]string) {
	f.t.Helper()
	got := f.events(eventType)
	require.Len(f.t, got, 1, eventType)
	for k, v := range want {
		require.Equal(f.t, v, got[0][k], "%s.%s", eventType, k)
	}
	require.Equal(f.t, 1, f.fastUpgradeEvents())
}

// startTimer has every validator vote Yes and runs the start block.
func (f *fixture) startTimer(id uint64) types.Timer {
	f.t.Helper()
	f.votes(id, numValidators, yes())
	f.endBlock()
	timer, ok := f.timer()
	require.True(f.t, ok)
	return timer
}

// tokens is the stake of each validator. The tests use it as the unit for a
// delegation and for the expected Yes count.
var tokens = exrpintegration.DefaultBondedAmount
