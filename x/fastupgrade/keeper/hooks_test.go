package keeper_test

import (
	"testing"
	"time"

	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/require"

	"github.com/xrplevm/node/v12/x/fastupgrade/types"
)

func TestHooks_AfterProposalVote(t *testing.T) {
	tt := []struct {
		name string
		// setup runs before the proposal. It is optional.
		setup func(f *fixture)
		// propose submits the proposal. The default is f.proposeMinor().
		propose func(f *fixture) uint64
		// nonValidator makes f.other vote. The default voter is f.vals[0].
		nonValidator bool
		// options is the vote. The default is yes().
		options govv1.WeightedVoteOptions
		// noBaseVersion uses a fixture with the module enabled and no completed
		// upgrade.
		noBaseVersion     bool
		expectedNominated bool
	}{
		{
			name:              "should nominate - validator yes",
			expectedNominated: true,
		},
		{
			name:              "should nominate - validator weighted positive yes",
			options:           split("0.000000000000000001", "0.999999999999999999"),
			expectedNominated: true,
		},
		{
			name:    "should not nominate - validator no",
			options: no(),
		},
		{
			name:    "should not nominate - validator abstain",
			options: govv1.NewNonSplitVoteOption(govv1.OptionAbstain),
		},
		{
			name:    "should not nominate - validator no with veto",
			options: govv1.NewNonSplitVoteOption(govv1.OptionNoWithVeto),
		},
		{
			name: "should not nominate - validator weighted without yes",
			options: govv1.WeightedVoteOptions{
				{Option: govv1.OptionNo, Weight: "0.5"},
				{Option: govv1.OptionAbstain, Weight: "0.5"},
			},
		},
		{
			name:         "should not nominate - non-validator yes",
			nonValidator: true,
		},
		{
			name: "should not nominate - disabled",
			setup: func(f *fixture) {
				f.setParams(false, 150, "0.85")
			},
		},
		{
			name: "should not nominate - jailed validator yes",
			setup: func(f *fixture) {
				f.jail(f.vals[0])
			},
		},
		{
			name: "should not nominate - timer running",
			setup: func(f *fixture) {
				require.NoError(f.t, f.k.Timer.Set(f.ctx, types.Timer{ProposalId: 99, UpgradeHeight: 500}))
			},
		},
		{
			name:          "should not nominate - no completed upgrade",
			noBaseVersion: true,
		},
		{
			name: "should not nominate - two messages",
			propose: func(f *fixture) uint64 {
				return f.propose(f.upgradeMsg(minorVersion), f.upgradeMsg("v1.2.0"))
			},
		},
		{
			name: "should not nominate - not an upgrade message",
			propose: func(f *fixture) uint64 {
				return f.propose(&types.MsgUpdateParams{Authority: f.govAddr, Params: types.DefaultParams()})
			},
		},
		{
			name: "should not nominate - major bump",
			propose: func(f *fixture) uint64 {
				return f.propose(f.upgradeMsg("v2.0.0"))
			},
		},
		{
			name: "should not nominate - patch only",
			propose: func(f *fixture) uint64 {
				return f.propose(f.upgradeMsg("v1.0.1"))
			},
		},
		{
			name: "should not nominate - non-strict name",
			propose: func(f *fixture) uint64 {
				return f.propose(f.upgradeMsg("v1.01.0"))
			},
		},
		{
			name:          "should not nominate - already completed name",
			setup:         func(f *fixture) { f.completeUpgrades(minorVersion, baseVersion) },
			noBaseVersion: true,
		},
		{
			// The last completed upgrade is v1.1.0, the latest by height, and
			// not v1.5.0, the highest version.
			name:  "should nominate - minor bump of the latest upgrade by height",
			setup: func(f *fixture) { f.completeUpgrades("v1.5.0", minorVersion) },
			propose: func(f *fixture) uint64 {
				return f.propose(f.upgradeMsg("v1.2.0"))
			},
			noBaseVersion:     true,
			expectedNominated: true,
		},
		{
			name:  "should not nominate - same minor as the latest upgrade by height",
			setup: func(f *fixture) { f.completeUpgrades("v1.5.0", minorVersion) },
			propose: func(f *fixture) uint64 {
				return f.propose(f.upgradeMsg("v1.1.5"))
			},
			noBaseVersion: true,
		},
		{
			name:          "should not nominate - latest upgrade name is not a version",
			setup:         func(f *fixture) { f.completeUpgrades(baseVersion, "hotfix") },
			noBaseVersion: true,
		},
		{
			name: "should nominate - minor bump with patch",
			propose: func(f *fixture) uint64 {
				return f.propose(f.upgradeMsg("v1.3.7"))
			},
			expectedNominated: true,
		},
		{
			name: "should not nominate - plan fails validation",
			propose: func(f *fixture) uint64 {
				msg := f.upgradeMsg(minorVersion)
				msg.Plan.Time = f.ctx.BlockTime().Add(time.Hour)
				return f.propose(msg)
			},
		},
		{
			name: "should not nominate - plan height in the past",
			propose: func(f *fixture) uint64 {
				msg := f.upgradeMsg(minorVersion)
				msg.Plan.Height = startHeight - 1
				return f.propose(msg)
			},
		},
		{
			name: "should nominate - plan height is the current height",
			propose: func(f *fixture) uint64 {
				msg := f.upgradeMsg(minorVersion)
				msg.Plan.Height = startHeight
				return f.propose(msg)
			},
			expectedNominated: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := newHooksFixture(t, tc.noBaseVersion)
			if tc.setup != nil {
				tc.setup(f)
			}
			if tc.propose == nil {
				tc.propose = (*fixture).proposeMinor
			}
			if tc.options == nil {
				tc.options = yes()
			}
			voter := f.vals[0]
			if tc.nonValidator {
				voter = f.other
			}

			id := tc.propose(f)
			f.vote(id, voter, tc.options)

			require.Equal(t, tc.expectedNominated, f.nominated(id))
		})
	}
}

// newHooksFixture returns the normal fixture, or with noBaseVersion the one
// with no completed upgrade.
func newHooksFixture(t *testing.T, noBaseVersion bool) *fixture {
	if noBaseVersion {
		return newFixtureWithoutBaseVersion(t)
	}
	return newFixture(t)
}

func TestHooks_AfterProposalVote_ErrorNeverRejectsVote(t *testing.T) {
	f := newFixture(t)
	id := f.proposeMinor()
	require.NoError(t, f.k.Params.Remove(f.ctx))
	f.vote(id, f.vals[0], yes())
	require.False(t, f.nominated(id))
	require.Equal(t, 1, f.storedVotes(id))
}
