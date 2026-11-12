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
		// noBase uses a fixture with no completed upgrade.
		noBase            bool
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
			name: "should not nominate - no completed upgrade baseline",
			setup: func(f *fixture) {
				f.setParams(true, 150, "0.85")
			},
			noBase: true,
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
			name: "should not nominate - already completed name",
			setup: func(f *fixture) {
				f.setParams(true, 150, "0.85")
				f.setBlock(10, f.ctx.BlockTime())
				f.applyUpgrade(minorVersion)
				f.setBlock(20, f.ctx.BlockTime())
				f.applyUpgrade(baseVersion)
				f.setBlock(startHeight, f.ctx.BlockTime())
			},
			noBase: true,
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
			var f *fixture
			if tc.noBase {
				f = newFixtureWithoutBaseline(t)
			} else {
				f = newFixture(t)
			}
			if tc.setup != nil {
				tc.setup(f)
			}
			var id uint64
			if tc.propose != nil {
				id = tc.propose(f)
			} else {
				id = f.proposeMinor()
			}
			voter := f.vals[0]
			if tc.nonValidator {
				voter = f.other
			}
			options := tc.options
			if options == nil {
				options = yes()
			}
			f.vote(id, voter, options)
			require.Equal(t, tc.expectedNominated, f.nominated(id))
		})
	}
}

func TestHooks_AfterProposalVote_ErrorNeverRejectsVote(t *testing.T) {
	f := newFixture(t)
	id := f.proposeMinor()
	require.NoError(t, f.k.Params.Remove(f.ctx))
	f.vote(id, f.vals[0], yes())
	require.False(t, f.nominated(id))
	require.Equal(t, 1, f.storedVotes(id))
}

func TestHooks_AfterProposalVote_LastCompletedUpgradeIsLatestByHeight(t *testing.T) {
	t.Run("latest by height, not highest version", func(t *testing.T) {
		f := newFixtureWithoutBaseline(t)
		f.setParams(true, 150, "0.85")
		f.setBlock(10, f.ctx.BlockTime())
		f.applyUpgrade("v1.5.0")
		f.setBlock(20, f.ctx.BlockTime())
		f.applyUpgrade(minorVersion)
		f.setBlock(startHeight, f.ctx.BlockTime())

		bump := f.propose(f.upgradeMsg("v1.2.0"))
		sameMinor := f.propose(f.upgradeMsg("v1.1.5"))
		f.vote(bump, f.vals[0], yes())
		f.vote(sameMinor, f.vals[0], yes())
		require.True(t, f.nominated(bump))
		require.False(t, f.nominated(sameMinor))
	})

	t.Run("unparsable latest name disables the fast path", func(t *testing.T) {
		f := newFixtureWithoutBaseline(t)
		f.setParams(true, 150, "0.85")
		f.setBlock(10, f.ctx.BlockTime())
		f.applyUpgrade(baseVersion)
		f.setBlock(20, f.ctx.BlockTime())
		f.applyUpgrade("hotfix")
		f.setBlock(startHeight, f.ctx.BlockTime())

		id := f.proposeMinor()
		f.vote(id, f.vals[0], yes())
		require.False(t, f.nominated(id))
	})
}
