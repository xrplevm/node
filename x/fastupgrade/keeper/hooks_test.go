package keeper_test

import (
	"testing"
	"time"

	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/require"

	v12 "github.com/xrplevm/node/v12/app/upgrades/v12"
	"github.com/xrplevm/node/v12/x/fastupgrade/types"
)

func TestHookNominations(t *testing.T) {
	for _, tc := range []struct {
		name         string
		setup        func(f *fixture)        // runs before the proposal
		propose      func(f *fixture) uint64 // default: f.proposeMinor()
		nonValidator bool                    // vote from f.other instead of f.vals[0]
		options      govv1.WeightedVoteOptions
		noBase       bool
		wantNom      bool
	}{
		{name: "validator yes", wantNom: true},
		{name: "validator weighted positive yes", options: split("0.000000000000000001", "0.999999999999999999"), wantNom: true},
		{name: "validator no", options: no()},
		{name: "validator abstain", options: govv1.NewNonSplitVoteOption(govv1.OptionAbstain)},
		{name: "validator no with veto", options: govv1.NewNonSplitVoteOption(govv1.OptionNoWithVeto)},
		{name: "validator weighted without yes", options: govv1.WeightedVoteOptions{
			{Option: govv1.OptionNo, Weight: "0.5"},
			{Option: govv1.OptionAbstain, Weight: "0.5"},
		}},
		{name: "non-validator yes", nonValidator: true},
		{name: "disabled", setup: func(f *fixture) {
			f.setParams(false, 150, "0.85")
		}},
		{name: "jailed validator yes", setup: func(f *fixture) {
			f.jail(f.vals[0])
		}},
		{name: "timer running", setup: func(f *fixture) {
			require.NoError(f.t, f.k.Timer.Set(f.ctx, types.Timer{ProposalId: 99, UpgradeHeight: 500}))
		}},
		{name: "no completed upgrade baseline", setup: func(f *fixture) {
			f.setParams(true, 150, "0.85")
		}, noBase: true},
		{name: "two messages", propose: func(f *fixture) uint64 {
			return f.propose(f.upgradeMsg("v12.1.0"), f.upgradeMsg("v12.2.0"))
		}},
		{name: "not an upgrade message", propose: func(f *fixture) uint64 {
			return f.propose(&types.MsgUpdateParams{Authority: f.govAddr, Params: types.DefaultParams()})
		}},
		{name: "major bump", propose: func(f *fixture) uint64 {
			return f.propose(f.upgradeMsg("v13.0.0"))
		}},
		{name: "patch only", propose: func(f *fixture) uint64 {
			return f.propose(f.upgradeMsg("v12.0.1"))
		}},
		{name: "non-strict name", propose: func(f *fixture) uint64 {
			return f.propose(f.upgradeMsg("v12.01.0"))
		}},
		{name: "already completed name", setup: func(f *fixture) {
			f.setParams(true, 150, "0.85")
			f.setBlock(10, f.ctx.BlockTime())
			f.applyUpgrade("v12.1.0")
			f.setBlock(20, f.ctx.BlockTime())
			f.applyUpgrade(v12.UpgradeName)
			f.setBlock(startHeight, f.ctx.BlockTime())
		}, noBase: true},
		{name: "minor bump with patch", propose: func(f *fixture) uint64 {
			return f.propose(f.upgradeMsg("v12.3.7"))
		}, wantNom: true},
		{name: "plan fails validation", propose: func(f *fixture) uint64 {
			msg := f.upgradeMsg("v12.1.0")
			msg.Plan.Time = f.ctx.BlockTime().Add(time.Hour)
			return f.propose(msg)
		}},
		{name: "plan height in the past", propose: func(f *fixture) uint64 {
			msg := f.upgradeMsg("v12.1.0")
			msg.Plan.Height = startHeight - 1
			return f.propose(msg)
		}},
		{name: "plan height is the current height", propose: func(f *fixture) uint64 {
			msg := f.upgradeMsg("v12.1.0")
			msg.Plan.Height = startHeight
			return f.propose(msg)
		}, wantNom: true},
	} {
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
			require.Equal(t, tc.wantNom, f.nominated(id))
		})
	}
}

func TestHookErrorNeverRejectsVote(t *testing.T) {
	f := newFixture(t)
	id := f.proposeMinor()
	require.NoError(t, f.k.Params.Remove(f.ctx))
	f.vote(id, f.vals[0], yes())
	require.False(t, f.nominated(id))
	require.Equal(t, 1, f.storedVotes(id))
}

func TestLastCompletedUpgradeIsLatestByHeight(t *testing.T) {
	t.Run("latest by height, not highest version", func(t *testing.T) {
		f := newFixtureWithoutBaseline(t)
		f.setParams(true, 150, "0.85")
		f.setBlock(10, f.ctx.BlockTime())
		f.applyUpgrade("v12.5.0")
		f.setBlock(20, f.ctx.BlockTime())
		f.applyUpgrade("v12.1.0")
		f.setBlock(startHeight, f.ctx.BlockTime())

		bump := f.propose(f.upgradeMsg("v12.2.0"))
		sameMinor := f.propose(f.upgradeMsg("v12.1.5"))
		f.vote(bump, f.vals[0], yes())
		f.vote(sameMinor, f.vals[0], yes())
		require.True(t, f.nominated(bump))
		require.False(t, f.nominated(sameMinor))
	})

	t.Run("unparsable latest name disables the fast path", func(t *testing.T) {
		f := newFixtureWithoutBaseline(t)
		f.setParams(true, 150, "0.85")
		f.setBlock(10, f.ctx.BlockTime())
		f.applyUpgrade("v12.0.0")
		f.setBlock(20, f.ctx.BlockTime())
		f.applyUpgrade("hotfix")
		f.setBlock(startHeight, f.ctx.BlockTime())

		id := f.proposeMinor()
		f.vote(id, f.vals[0], yes())
		require.False(t, f.nominated(id))
	})
}
