package keeper_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/require"

	"github.com/xrplevm/node/v12/x/fastupgrade/types"
	poatypes "github.com/xrplevm/node/v12/x/poa/types"
)

func TestKeeper_EndBlock_TimerStartKeepsVotesAndTally(t *testing.T) {
	f := newFixture(t)
	id := f.proposeMinor()
	f.votes(id, numValidators, yes())

	scratch, _ := f.ctx.CacheContext()
	_, _, before, err := f.app.GovKeeper.Tally(scratch, f.proposal(id))
	require.NoError(t, err)

	f.endBlock()

	timer, ok := f.timer()
	require.True(t, ok)
	require.Equal(t, types.Timer{ProposalId: id, UpgradeHeight: startHeight + 150}, timer)
	f.requireSingleEvent(types.EventTypeStarted, map[string]string{
		types.AttributeKeyProposalID:    "1",
		types.AttributeKeyPlanName:      "v12.1.0",
		types.AttributeKeyPlanInfo:      "info-v12.1.0",
		types.AttributeKeyUpgradeHeight: "250",
	})

	require.Equal(t, numValidators, f.storedVotes(id))
	scratch, _ = f.ctx.CacheContext()
	_, _, after, err := f.app.GovKeeper.Tally(scratch, f.proposal(id))
	require.NoError(t, err)
	require.Equal(t, before, after)
	require.Equal(t, tokens.MulRaw(numValidators).String(), after.YesCount)
}

func TestKeeper_EndBlock_ValidatorGate(t *testing.T) {
	tt := []struct {
		name            string
		threshold       string
		votes           func(f *fixture, id uint64)
		expectedStarted bool
	}{
		{
			name:      "should fail - 4 of 5 is below 0.85",
			threshold: "0.85",
			votes:     func(f *fixture, id uint64) { f.votes(id, 4, yes()) },
		},
		{
			name:            "should pass - 4 of 5 equals 0.8",
			threshold:       "0.8",
			votes:           func(f *fixture, id uint64) { f.votes(id, 4, yes()) },
			expectedStarted: true,
		},
		{
			name:      "should fail - 4 of 5 is one unit below the threshold",
			threshold: "0.800000000000000001",
			votes:     func(f *fixture, id uint64) { f.votes(id, 4, yes()) },
		},
		{
			name:      "should pass - weighted yes counts its weight",
			threshold: "0.85",
			votes: func(f *fixture, id uint64) {
				f.votes(id, 4, yes())
				f.vote(id, f.vals[4], split("0.5", "0.5"))
			},
			expectedStarted: true,
		},
		{
			name:      "should fail - weighted yes is below the threshold",
			threshold: "0.85",
			votes: func(f *fixture, id uint64) {
				f.votes(id, 3, yes())
				f.vote(id, f.vals[3], split("0.5", "0.5"))
				f.vote(id, f.vals[4], split("0.5", "0.5"))
			},
		},
		{
			name:      "should fail - non-validator votes are not counted",
			threshold: "0.85",
			votes: func(f *fixture, id uint64) {
				f.votes(id, 4, yes())
				f.vote(id, f.other, yes())
			},
		},
		{
			name:      "should pass - jailed validator is not in the bonded pool",
			threshold: "0.85",
			votes: func(f *fixture, id uint64) {
				f.jail(f.vals[4])
				f.votes(id, 4, yes())
			},
			expectedStarted: true,
		},
		{
			// One extra base unit per validator makes each weight * tokens
			// term fractional, while the sum equals threshold * bonded exactly.
			name:      "should pass - weighted terms are not rounded per validator",
			threshold: "0.8",
			votes: func(f *fixture, id uint64) {
				for _, v := range f.vals {
					f.delegate(f.other, v, sdkmath.OneInt())
				}
				f.votes(id, numValidators, split("0.8", "0.2"))
			},
			expectedStarted: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			f.setParams(true, 150, tc.threshold)
			id := f.proposeMinor()
			tc.votes(f, id)
			f.endBlock()
			_, ok := f.timer()
			require.Equal(t, tc.expectedStarted, ok)
			if !tc.expectedStarted {
				require.Zero(t, f.fastUpgradeEvents())
			}
		})
	}
}

func TestKeeper_EndBlock_NewestCandidateDecides(t *testing.T) {
	t.Run("newest ready proposal starts", func(t *testing.T) {
		f := newFixture(t)
		older := f.proposeMinor()
		newer := f.propose(f.upgradeMsg("v12.2.0"))
		f.votes(older, numValidators, yes())
		f.votes(newer, numValidators, yes())
		f.endBlock()
		timer, ok := f.timer()
		require.True(t, ok)
		require.Equal(t, newer, timer.ProposalId)
	})

	t.Run("candidate below gate is skipped", func(t *testing.T) {
		f := newFixture(t)
		older := f.proposeMinor()
		newer := f.propose(f.upgradeMsg("v12.2.0"))
		f.votes(older, numValidators, yes())
		f.votes(newer, 3, yes())
		f.endBlock()
		timer, ok := f.timer()
		require.True(t, ok)
		require.Equal(t, older, timer.ProposalId)
	})

	t.Run("candidate passing the gate decides even when its tally rejects", func(t *testing.T) {
		f := newFixture(t)
		older := f.proposeMinor()
		newer := f.propose(f.upgradeMsg("v12.2.0"))
		f.delegate(f.other, f.vals[0], tokens.MulRaw(3))
		f.votes(older, numValidators, yes())
		f.votes(newer, numValidators, yes())
		f.vote(newer, f.other, govv1.NewNonSplitVoteOption(govv1.OptionNoWithVeto))
		f.endBlock()
		_, ok := f.timer()
		require.False(t, ok)
		f.requireSingleEvent(types.EventTypeNotStarted, map[string]string{
			types.AttributeKeyProposalID: "2",
			types.AttributeKeyReason:     "gov_tally_rejected",
		})
	})
}

func TestKeeper_EndBlock_DelegatorVotesCanOnlyStopTheStart(t *testing.T) {
	tt := []struct {
		name            string
		amount          sdkmath.Int
		option          govv1.VoteOption
		expectedStarted bool
		expectedReason  string
	}{
		{
			name:            "should pass - delegator yes",
			amount:          tokens,
			option:          govv1.OptionYes,
			expectedStarted: true,
		},
		{
			name:           "should fail - delegator no lowers gov yes count",
			amount:         tokens,
			option:         govv1.OptionNo,
			expectedReason: "tally_below_threshold",
		},
		{
			name:           "should fail - delegator veto rejects gov tally",
			amount:         tokens.MulRaw(3),
			option:         govv1.OptionNoWithVeto,
			expectedReason: "gov_tally_rejected",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			id := f.proposeMinor()
			f.delegate(f.other, f.vals[0], tc.amount)
			f.votes(id, numValidators, yes())
			f.vote(id, f.other, govv1.NewNonSplitVoteOption(tc.option))
			f.endBlock()
			_, ok := f.timer()
			require.Equal(t, tc.expectedStarted, ok)
			if !tc.expectedStarted {
				f.requireSingleEvent(types.EventTypeNotStarted, map[string]string{types.AttributeKeyReason: tc.expectedReason})
			}
			require.Equal(t, numValidators+1, f.storedVotes(id))
		})
	}
}

func TestKeeper_EndBlock_StartSkips(t *testing.T) {
	t.Run("pending plan", func(t *testing.T) {
		f := newFixture(t)
		require.NoError(t, f.app.UpgradeKeeper.ScheduleUpgrade(f.ctx, upgradetypes.Plan{Name: "other", Height: 5000}))
		id := f.proposeMinor()
		f.votes(id, numValidators, yes())
		f.endBlock()
		_, ok := f.timer()
		require.False(t, ok)
		require.Zero(t, f.fastUpgradeEvents())
	})

	t.Run("natural deadline due", func(t *testing.T) {
		f := newFixture(t)
		id := f.proposeMinor()
		f.votes(id, numValidators, yes())
		f.setBlock(f.ctx.BlockHeight(), *f.proposal(id).VotingEndTime)
		f.endBlock()
		_, ok := f.timer()
		require.False(t, ok)
		require.Zero(t, f.fastUpgradeEvents())
	})

	t.Run("gov params incompatible", func(t *testing.T) {
		f := newFixture(t)
		id := f.proposeMinor()
		f.votes(id, numValidators, yes())
		f.setGovParams(func(p *govv1.Params) { p.Threshold = incompatibleGovThreshold })
		f.endBlock()
		_, ok := f.timer()
		require.False(t, ok)
		f.requireSingleEvent(types.EventTypeNotStarted, map[string]string{
			types.AttributeKeyProposalID: "1",
			types.AttributeKeyReason:     "incompatible_with_gov",
		})
	})

	t.Run("gov params incompatible below the gate stays silent", func(t *testing.T) {
		f := newFixture(t)
		id := f.proposeMinor()
		f.votes(id, 4, yes())
		f.setGovParams(func(p *govv1.Params) { p.Threshold = incompatibleGovThreshold })
		f.endBlock()
		require.Zero(t, f.fastUpgradeEvents())
	})

	// The newer proposal could only drop at H-1, so the older one starts.
	t.Run("plan height below the final check", func(t *testing.T) {
		f := newFixture(t)
		healthy := f.proposeMinor()
		msg := f.upgradeMsg("v12.2.0")
		msg.Plan.Height = startHeight + int64(types.DefaultTimelockBlocks) - 2
		stale := f.propose(msg)
		f.votes(healthy, numValidators, yes())
		f.votes(stale, numValidators, yes())
		f.endBlock()
		timer, ok := f.timer()
		require.True(t, ok)
		require.Equal(t, healthy, timer.ProposalId)
		f.requireSingleEvent(types.EventTypeStarted, nil)
	})

	t.Run("plan height is the final check", func(t *testing.T) {
		f := newFixture(t)
		msg := f.upgradeMsg("v12.1.0")
		msg.Plan.Height = startHeight + int64(types.DefaultTimelockBlocks) - 1
		id := f.propose(msg)
		f.votes(id, numValidators, yes())
		f.endBlock()
		_, ok := f.timer()
		require.True(t, ok)
	})

	t.Run("upgrade height overflow falls back", func(t *testing.T) {
		f := newFixture(t)
		msg := f.upgradeMsg("v12.1.0")
		msg.Plan.Height = math.MaxInt64
		id := f.propose(msg)
		f.votes(id, numValidators, yes())
		f.setBlock(math.MaxInt64-10, f.ctx.BlockTime())
		f.endBlock()
		_, ok := f.timer()
		require.False(t, ok)
		require.Zero(t, f.fastUpgradeEvents())
		require.Equal(t, numValidators, f.storedVotes(id))
	})
}

func TestKeeper_EndBlock_EarlyCloseExecutesInSameBlock(t *testing.T) {
	f := newFixture(t)
	id := f.proposeMinor()
	original := f.proposal(id)
	timer := f.startTimer(id)
	H := timer.UpgradeHeight

	f.advanceTo(H - 2)
	f.fullEndBlock()
	_, ok := f.timer()
	require.True(t, ok)
	require.Zero(t, f.fastUpgradeEvents())
	require.True(t, f.queued(*original.VotingEndTime, id))

	f.nextBlock()
	require.Equal(t, H-1, f.ctx.BlockHeight())
	f.fullEndBlock()

	_, ok = f.timer()
	require.False(t, ok)
	f.requireSingleEvent(types.EventTypeEarlyClose, map[string]string{
		types.AttributeKeyProposalID:    "1",
		types.AttributeKeyPlanName:      "v12.1.0",
		types.AttributeKeyUpgradeHeight: "250",
	})

	p := f.proposal(id)
	require.Equal(t, govv1.StatusPassed, p.Status)
	require.Equal(t, f.ctx.BlockTime(), *p.VotingEndTime)
	require.Equal(t, H, f.planHeight(p))
	require.Equal(t, tokens.MulRaw(numValidators).String(), p.FinalTallyResult.YesCount)
	require.False(t, f.queued(*original.VotingEndTime, id))
	require.False(t, f.queued(f.ctx.BlockTime(), id))

	plan, err := f.app.UpgradeKeeper.GetUpgradePlan(f.ctx)
	require.NoError(t, err)
	require.Equal(t, upgradetypes.Plan{Name: "v12.1.0", Height: H, Info: "info-v12.1.0"}, plan)
}

func TestKeeper_EndBlock_TimelockAndThresholdUpdatesDuringTimer(t *testing.T) {
	f := newFixture(t)
	f.setParams(true, 150, "0.8")
	id := f.proposeMinor()
	f.votes(id, 4, yes())
	f.endBlock()
	timer, ok := f.timer()
	require.True(t, ok)

	f.nextBlock()
	f.setParams(true, 10, "0.85")
	f.endBlock()
	unchanged, ok := f.timer()
	require.True(t, ok)
	require.Equal(t, timer, unchanged)

	f.advanceTo(timer.UpgradeHeight - 1)
	f.fullEndBlock()
	f.requireSingleEvent(types.EventTypeDropped, map[string]string{
		types.AttributeKeyProposalID:    "1",
		types.AttributeKeyUpgradeHeight: "250",
		types.AttributeKeyReason:        "gate_below_threshold",
	})
	require.Equal(t, govv1.StatusVotingPeriod, f.proposal(id).Status)
}

func TestKeeper_EndBlock_FinalCheckDropsKeepOriginalProposal(t *testing.T) {
	tt := []struct {
		name           string
		expectedReason string
		// during runs after the timer started, before H-1. It is optional.
		during func(f *fixture, id uint64)
		// atFinal runs in the H-1 block before EndBlock. It is optional.
		atFinal func(f *fixture, id uint64)
	}{
		{
			name:           "should drop the timer - vote change is the veto",
			expectedReason: "gate_below_threshold",
			during: func(f *fixture, id uint64) {
				f.vote(id, f.vals[0], no())
			},
		},
		{
			name:           "should drop the timer - gov params changed",
			expectedReason: "incompatible_with_gov",
			during: func(f *fixture, _ uint64) {
				f.setGovParams(func(p *govv1.Params) { p.ExpeditedThreshold = incompatibleGovThreshold })
			},
		},
		{
			name:           "should drop the timer - gate fails and gov params changed",
			expectedReason: "gate_below_threshold",
			during: func(f *fixture, id uint64) {
				f.vote(id, f.vals[0], no())
				f.setGovParams(func(p *govv1.Params) { p.ExpeditedThreshold = incompatibleGovThreshold })
			},
		},
		{
			name:           "should drop the timer - another minor completed",
			expectedReason: "not_eligible",
			during: func(f *fixture, _ uint64) {
				f.applyUpgrade("v12.3.0")
			},
		},
		{
			name:           "should drop the timer - plan pending",
			expectedReason: "plan_pending",
			during: func(f *fixture, _ uint64) {
				require.NoError(f.t, f.app.UpgradeKeeper.ScheduleUpgrade(f.ctx, upgradetypes.Plan{Name: "other", Height: 5000}))
			},
		},
		{
			name:           "should drop the timer - delegator lowers gov yes count",
			expectedReason: "tally_below_threshold",
			during: func(f *fixture, id uint64) {
				f.delegate(f.other, f.vals[0], tokens)
				f.vote(id, f.other, no())
			},
		},
		{
			name:           "should drop the timer - delegator vetoes",
			expectedReason: "gov_tally_rejected",
			during: func(f *fixture, id uint64) {
				f.delegate(f.other, f.vals[0], tokens.MulRaw(3))
				f.vote(id, f.other, govv1.NewNonSplitVoteOption(govv1.OptionNoWithVeto))
			},
		},
		{
			name:           "should drop the timer - natural deadline due",
			expectedReason: "natural_deadline",
			atFinal: func(f *fixture, id uint64) {
				f.setBlock(f.ctx.BlockHeight(), *f.proposal(id).VotingEndTime)
			},
		},
		{
			// The proposal's own deadline wins over the postponement.
			name:           "should drop the timer - natural deadline due and another proposal due",
			expectedReason: "natural_deadline",
			during: func(f *fixture, _ uint64) {
				period := time.Second
				f.setGovParams(func(p *govv1.Params) { p.VotingPeriod = &period })
				f.propose(f.upgradeMsg("v13.0.0"))
			},
			atFinal: func(f *fixture, id uint64) {
				f.setBlock(f.ctx.BlockHeight(), *f.proposal(id).VotingEndTime)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			id := f.proposeMinor()
			timer := f.startTimer(id)
			original := f.proposal(id)

			f.nextBlock()
			if tc.during != nil {
				tc.during(f, id)
			}
			votes := f.storedVotes(id)
			f.endBlock()
			_, ok := f.timer()
			require.True(t, ok)

			f.advanceTo(timer.UpgradeHeight - 1)
			if tc.atFinal != nil {
				tc.atFinal(f, id)
			}
			f.endBlock()

			_, ok = f.timer()
			require.False(t, ok)
			f.requireSingleEvent(types.EventTypeDropped, map[string]string{
				types.AttributeKeyProposalID:    "1",
				types.AttributeKeyUpgradeHeight: "250",
				types.AttributeKeyReason:        tc.expectedReason,
			})

			p := f.proposal(id)
			require.Equal(t, original, p)
			require.Equal(t, originalHeight, f.planHeight(p))
			require.True(t, f.queued(*original.VotingEndTime, id))
			require.Equal(t, votes, f.storedVotes(id))
		})
	}
}

func TestKeeper_EndBlock_NaturalDeadlineAtFinalCheckUsesOriginalHeight(t *testing.T) {
	f := newFixture(t)
	id := f.proposeMinor()
	timer := f.startTimer(id)
	f.advanceTo(timer.UpgradeHeight - 1)
	f.setBlock(f.ctx.BlockHeight(), *f.proposal(id).VotingEndTime)
	f.fullEndBlock()

	f.requireSingleEvent(types.EventTypeDropped, map[string]string{types.AttributeKeyReason: "natural_deadline"})
	p := f.proposal(id)
	require.Equal(t, govv1.StatusPassed, p.Status)
	plan, err := f.app.UpgradeKeeper.GetUpgradePlan(f.ctx)
	require.NoError(t, err)
	require.Equal(t, originalHeight, plan.Height)
}

func TestKeeper_EndBlock_TimerDropped(t *testing.T) {
	tt := []struct {
		name           string
		expectedReason string
		act            func(f *fixture, id uint64, timer types.Timer)
	}{
		{
			name:           "should drop the timer - module disabled",
			expectedReason: "disabled",
			act: func(f *fixture, _ uint64, _ types.Timer) {
				f.setParams(false, 150, "0.85")
			},
		},
		{
			name:           "should drop the timer - proposal canceled",
			expectedReason: "not_voting",
			act: func(f *fixture, id uint64, _ types.Timer) {
				require.NoError(f.t, f.app.GovKeeper.CancelProposal(f.ctx, id, f.other.String()))
			},
		},
		{
			name:           "should drop the timer - final check block missed",
			expectedReason: "missed_window",
			act: func(f *fixture, _ uint64, timer types.Timer) {
				f.setBlock(timer.UpgradeHeight, f.ctx.BlockTime())
			},
		},
		{
			// The module runs before gov, so it sees the deadline in the
			// block where gov ends the vote.
			name:           "should drop the timer - deadline before the final check",
			expectedReason: "natural_deadline",
			act: func(f *fixture, id uint64, _ types.Timer) {
				f.setBlock(f.ctx.BlockHeight(), *f.proposal(id).VotingEndTime)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			id := f.proposeMinor()
			timer := f.startTimer(id)
			f.nextBlock()
			tc.act(f, id, timer)
			f.endBlock()
			_, ok := f.timer()
			require.False(t, ok)
			f.requireSingleEvent(types.EventTypeDropped, map[string]string{
				types.AttributeKeyProposalID: "1",
				types.AttributeKeyReason:     tc.expectedReason,
			})

			f.nextBlock()
			f.endBlock()
			require.Zero(t, f.fastUpgradeEvents())
		})
	}
}

// failingEncode is a proposal codec that reads but cannot store.
type failingEncode struct {
	collcodec.ValueCodec[govv1.Proposal]
}

func (failingEncode) Encode(govv1.Proposal) ([]byte, error) {
	return nil, errors.New("encode failed")
}

func TestKeeper_EndBlock_FailureClearsOnlyTheTimer(t *testing.T) {
	tt := []struct {
		name    string
		failure func(f *fixture)
	}{
		{
			name: "should drop the timer - error on an invalid stored threshold",
			failure: func(f *fixture) {
				require.NoError(f.t, f.k.Params.Set(f.ctx, types.Params{Enabled: true, TimelockBlocks: 150, Threshold: "bad"}))
			},
		},
		{
			// The fixture's validators have MinSelfDelegation 0, so a removed
			// one stays bonded with zero shares until staking's EndBlocker,
			// and gov's Tally divides by those shares.
			name: "should drop the timer - panic when gov tally divides by zero shares",
			failure: func(f *fixture) {
				require.NoError(f.t, f.app.PoaKeeper.ExecuteRemoveValidator(f.ctx, f.vals[0].String()))
			},
		},
		{
			// Gov's SetProposal fails after endVotingNow removed the queue
			// key, so only the discarded scratch copy keeps the proposal
			// queued.
			name: "should drop the timer - error after a gov write, the proposal cannot be stored",
			failure: func(f *fixture) {
				sb := collections.NewSchemaBuilder(runtime.NewKVStoreService(f.app.GetKey(govtypes.StoreKey)))
				f.app.GovKeeper.Proposals = collections.NewMap(sb, govtypes.ProposalsKeyPrefix, "proposals", collections.Uint64Key,
					collcodec.ValueCodec[govv1.Proposal](failingEncode{codec.CollValue[govv1.Proposal](f.app.AppCodec())}))
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			id := f.proposeMinor()
			timer := f.startTimer(id)
			original := f.proposal(id)
			f.advanceTo(timer.UpgradeHeight - 1)
			tc.failure(f)

			f.endBlock()
			_, ok := f.timer()
			require.False(t, ok)
			f.requireSingleEvent(types.EventTypeDropped, map[string]string{
				types.AttributeKeyProposalID:    "1",
				types.AttributeKeyUpgradeHeight: "250",
				types.AttributeKeyReason:        "internal_error",
			})
			require.Equal(t, original, f.proposal(id))
			require.True(t, f.queued(*original.VotingEndTime, id))
		})
	}
}

// Nominations are in the committed KV store, so EndBlock must clear them on
// every path: a leftover would be checked again in the next block.
func TestKeeper_EndBlock_ClearsNominations(t *testing.T) {
	tt := []struct {
		name   string
		voters int
		// before runs after the votes. It is optional.
		before        func(f *fixture)
		expectedTimer bool
	}{
		{
			name:          "should clear the nominations - timer started",
			voters:        numValidators,
			expectedTimer: true,
		},
		{
			name:   "should clear the nominations - below the gate",
			voters: 1,
		},
		{
			name:   "should clear the nominations - module disabled",
			voters: numValidators,
			before: func(f *fixture) {
				f.setParams(false, 150, "0.85")
			},
		},
		{
			name:   "should clear the nominations - error on an invalid stored threshold",
			voters: numValidators,
			before: func(f *fixture) {
				require.NoError(f.t, f.k.Params.Set(f.ctx, types.Params{Enabled: true, TimelockBlocks: 150, Threshold: "bad"}))
			},
		},
		{
			name:   "should clear the nominations - panic when gov tally divides by zero shares",
			voters: numValidators,
			before: func(f *fixture) {
				require.NoError(f.t, f.app.PoaKeeper.ExecuteRemoveValidator(f.ctx, f.vals[0].String()))
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			id := f.proposeMinor()
			f.votes(id, tc.voters, yes())
			require.True(t, f.nominated(id))
			if tc.before != nil {
				tc.before(f)
			}

			f.endBlock()
			require.False(t, f.nominated(id))
			_, ok := f.timer()
			require.Equal(t, tc.expectedTimer, ok)
			// A failed start check has no timer to retract, so it emits nothing.
			if !tc.expectedTimer {
				require.Zero(t, f.fastUpgradeEvents())
			}
		})
	}
}

// convertExpedited starts a timer on an expedited proposal and then fails
// its expedited tally at its deadline: the timer drops there, and gov makes
// the proposal regular and deletes its votes.
func convertExpedited(f *fixture) govv1.Proposal {
	f.t.Helper()
	p, err := f.app.GovKeeper.SubmitProposal(f.ctx, []sdk.Msg{f.upgradeMsg("v12.1.0")}, "", "title", "summary", f.other, true)
	require.NoError(f.t, err)
	require.NoError(f.t, f.app.GovKeeper.ActivateVotingPeriod(f.ctx, p))
	id := p.Id
	f.startTimer(id)

	f.nextBlock()
	f.vote(id, f.vals[0], no())
	f.vote(id, f.vals[1], no())
	f.fullEndBlock()

	f.nextBlock()
	expeditedEnd := *f.proposal(id).VotingEndTime
	f.setBlock(f.ctx.BlockHeight(), expeditedEnd)
	f.fullEndBlock()
	f.requireSingleEvent(types.EventTypeDropped, map[string]string{types.AttributeKeyReason: "natural_deadline"})

	converted := f.proposal(id)
	require.False(f.t, converted.Expedited)
	require.Equal(f.t, govv1.StatusVotingPeriod, converted.Status)
	require.True(f.t, converted.VotingEndTime.After(expeditedEnd))
	require.Zero(f.t, f.storedVotes(id))
	return converted
}

// An expedited proposal that becomes regular loses its timer at the
// expedited deadline. Gov deleted its votes, so only new Yes votes start a
// new timer, with a new H.
func TestKeeper_EndBlock_ExpeditedConvertedToRegularDuringTimer(t *testing.T) {
	t.Run("no new votes, no timer", func(t *testing.T) {
		f := newFixture(t)
		converted := convertExpedited(f)

		f.nextBlock()
		f.fullEndBlock()

		require.Zero(t, f.fastUpgradeEvents())
		_, ok := f.timer()
		require.False(t, ok)
		require.Equal(t, converted, f.proposal(converted.Id))
		require.True(t, f.queued(*converted.VotingEndTime, converted.Id))
	})

	t.Run("validators vote again, new timer and early close", func(t *testing.T) {
		f := newFixture(t)
		converted := convertExpedited(f)

		f.nextBlock()
		timer := f.startTimer(converted.Id)
		require.Equal(t, f.ctx.BlockHeight()+150, timer.UpgradeHeight)

		f.advanceTo(timer.UpgradeHeight - 1)
		f.fullEndBlock()

		f.requireSingleEvent(types.EventTypeEarlyClose, nil)
		p := f.proposal(converted.Id)
		require.Equal(t, govv1.StatusPassed, p.Status)
		require.Equal(t, timer.UpgradeHeight, f.planHeight(p))
		require.False(t, f.queued(*converted.VotingEndTime, converted.Id))
	})
}

// Gov runs after this module. A removal that gov executes in block H-1 leaves
// the validator bonded with zero shares until staking's EndBlocker, and gov's
// tally of a proposal closed early in that block would divide by them.
func TestKeeper_EndBlock_OtherProposalDueAtFinalCheckPostponesTimer(t *testing.T) {
	f := newFixture(t)
	params := types.DefaultParams()
	params.Enabled = true
	params.PostponeBlocks = 7
	require.NoError(t, f.k.SetParams(f.ctx, params))
	id := f.proposeMinor()
	timer := f.startTimer(id)
	original := f.proposal(id)

	finalCheck := timer.UpgradeHeight - 1
	period := time.Duration(finalCheck-f.ctx.BlockHeight())*time.Second - time.Millisecond
	f.setGovParams(func(p *govv1.Params) { p.VotingPeriod = &period })
	removal := f.propose(&poatypes.MsgRemoveValidator{Authority: f.govAddr, ValidatorAddress: f.vals[4].String()})
	f.votes(removal, numValidators, yes())

	f.advanceTo(finalCheck)
	f.fullEndBlock()
	_, err := f.app.StakingKeeper.EndBlocker(f.ctx)
	require.NoError(t, err)

	H := timer.UpgradeHeight + 7
	f.requireSingleEvent(types.EventTypePostponed, map[string]string{
		types.AttributeKeyProposalID:    "1",
		types.AttributeKeyUpgradeHeight: "257",
	})
	postponed, ok := f.timer()
	require.True(t, ok)
	require.Equal(t, types.Timer{ProposalId: id, UpgradeHeight: H}, postponed)
	require.Equal(t, govv1.StatusPassed, f.proposal(removal).Status)
	require.Equal(t, original, f.proposal(id))

	f.advanceTo(H - 1)
	f.fullEndBlock()
	f.requireSingleEvent(types.EventTypeEarlyClose, map[string]string{types.AttributeKeyUpgradeHeight: "257"})
	p := f.proposal(id)
	require.Equal(t, govv1.StatusPassed, p.Status)
	require.Equal(t, H, f.planHeight(p))
}
