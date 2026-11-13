package keeper_test

import (
	"errors"
	"math"
	"strconv"
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

// When all validators vote Yes, the block ends with a timer and one
// fastupgrade_started event. The start runs gov's tally, which deletes the
// votes it counts, so gov must still have every vote and the same tally after.
func TestKeeper_EndBlock_TimerStartKeepsVotesAndTally(t *testing.T) {
	f := newFixture(t)
	id := f.proposeMinor()
	f.votes(id, numValidators, yes())

	cacheCtx, _ := f.ctx.CacheContext()
	_, _, before, err := f.app.GovKeeper.Tally(cacheCtx, f.proposal(id))
	require.NoError(t, err)

	f.endBlock()

	timer, ok := f.timer()
	require.True(t, ok)
	require.Equal(t, types.Timer{ProposalId: id, UpgradeHeight: upgradeHeight}, timer)
	f.requireSingleEvent(types.EventTypeStarted, map[string]string{
		types.AttributeKeyProposalID:    firstProposalID,
		types.AttributeKeyPlanName:      minorVersion,
		types.AttributeKeyPlanInfo:      "info-" + minorVersion,
		types.AttributeKeyUpgradeHeight: upgradeHeightAttr,
	})

	require.Equal(t, numValidators, f.storedVotes(id))
	cacheCtx, _ = f.ctx.CacheContext()
	_, _, after, err := f.app.GovKeeper.Tally(cacheCtx, f.proposal(id))
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
			name:      "should not start the timer - 4 of 5 is below 0.85",
			threshold: "0.85",
			votes:     func(f *fixture, id uint64) { f.votes(id, 4, yes()) },
		},
		{
			name:            "should start the timer - 4 of 5 equals 0.8",
			threshold:       "0.8",
			votes:           func(f *fixture, id uint64) { f.votes(id, 4, yes()) },
			expectedStarted: true,
		},
		{
			name:      "should not start the timer - 4 of 5 is one unit below the threshold",
			threshold: "0.800000000000000001",
			votes:     func(f *fixture, id uint64) { f.votes(id, 4, yes()) },
		},
		{
			name:      "should start the timer - weighted yes counts its weight",
			threshold: "0.85",
			votes: func(f *fixture, id uint64) {
				f.votes(id, 4, yes())
				f.vote(id, f.vals[4], split("0.5", "0.5"))
			},
			expectedStarted: true,
		},
		{
			name:      "should not start the timer - weighted yes is below the threshold",
			threshold: "0.85",
			votes: func(f *fixture, id uint64) {
				f.votes(id, 3, yes())
				f.vote(id, f.vals[3], split("0.5", "0.5"))
				f.vote(id, f.vals[4], split("0.5", "0.5"))
			},
		},
		{
			name:      "should not start the timer - non-validator votes are not counted",
			threshold: "0.85",
			votes: func(f *fixture, id uint64) {
				f.votes(id, 4, yes())
				f.vote(id, f.other, yes())
			},
		},
		{
			name:      "should start the timer - jailed validator is not in the bonded pool",
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
			name:      "should start the timer - weighted terms are not rounded per validator",
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

			f.requireStartOutcome(id, tc.expectedStarted, "")
		})
	}
}

// Two proposals are nominated in one block, and the newest is checked first.
// These are subtests and not a table: each case sets up the newer proposal
// in its own way, so a table would need a field that only one case uses for
// almost every step.
func TestKeeper_EndBlock_NewestCandidateDecides(t *testing.T) {
	t.Run("should start the timer for the newest - both are ready", func(t *testing.T) {
		f := newFixture(t)
		older := f.proposeMinor()
		newer := f.propose(f.upgradeMsg("v1.2.0"))
		f.votes(older, numValidators, yes())
		f.votes(newer, numValidators, yes())
		f.endBlock()
		timer, ok := f.timer()
		require.True(t, ok)
		require.Equal(t, newer, timer.ProposalId)
		f.requireSingleEvent(types.EventTypeStarted, nil)
	})

	t.Run("should start the timer for the older - the newest is below the gate", func(t *testing.T) {
		f := newFixture(t)
		older := f.proposeMinor()
		newer := f.propose(f.upgradeMsg("v1.2.0"))
		f.votes(older, numValidators, yes())
		f.votes(newer, 3, yes())
		f.endBlock()
		timer, ok := f.timer()
		require.True(t, ok)
		require.Equal(t, older, timer.ProposalId)
		f.requireSingleEvent(types.EventTypeStarted, nil)
	})

	// The newest proposal could only drop at H-1, so the older one starts.
	t.Run("should start the timer for the older - the newest has a plan height below the final check", func(t *testing.T) {
		f := newFixture(t)
		older := f.proposeMinor()
		msg := f.upgradeMsg("v1.2.0")
		msg.Plan.Height = upgradeHeight - 2
		newer := f.propose(msg)
		f.votes(older, numValidators, yes())
		f.votes(newer, numValidators, yes())
		f.endBlock()
		timer, ok := f.timer()
		require.True(t, ok)
		require.Equal(t, older, timer.ProposalId)
		f.requireSingleEvent(types.EventTypeStarted, nil)
	})

	// The delegator's veto makes gov's tally reject the newer proposal while
	// the validator gate passes. The older proposal would pass every check,
	// so a timer here means that the module tried it after the tally.
	t.Run("should not start a timer - the newest passes the gate and its tally rejects", func(t *testing.T) {
		f := newFixture(t)
		older := f.proposeMinor()
		newer := f.propose(f.upgradeMsg("v1.2.0"))
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

// A delegation is the tests' way to make gov's tally differ from the
// validator gate: a delegator can vote against its validator. Mainnet has no
// such delegations, so these cases test tallyCheck, not a mainnet state.
func TestKeeper_EndBlock_DelegatorVotesCanOnlyStopTheStart(t *testing.T) {
	tt := []struct {
		name            string
		amount          sdkmath.Int
		option          govv1.VoteOption
		expectedStarted bool
		expectedReason  string
	}{
		{
			name:            "should start the timer - delegator yes",
			amount:          tokens,
			option:          govv1.OptionYes,
			expectedStarted: true,
		},
		{
			// Yes is 5 of 6 stakes, 0.833: gov would pass it, and it is
			// below our 0.85.
			name:           "should not start the timer - delegator no lowers gov yes count",
			amount:         tokens,
			option:         govv1.OptionNo,
			expectedReason: "tally_below_threshold",
		},
		{
			// The veto is 3 of 8 stakes, 0.375: above gov's veto threshold
			// of 0.334, so gov rejects it.
			name:           "should not start the timer - delegator veto rejects gov tally",
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

			f.requireStartOutcome(id, tc.expectedStarted, tc.expectedReason)
			require.Equal(t, numValidators+1, f.storedVotes(id))
		})
	}
}

// The checks of guards at the timer start, apart from the validator gate.
func TestKeeper_EndBlock_StartGuards(t *testing.T) {
	finalCheck := upgradeHeight - 1
	incompatibleGov := func(f *fixture, _ uint64) {
		f.setGovParams(func(p *govv1.Params) { p.Threshold = incompatibleGovThreshold })
	}

	tt := []struct {
		name string
		// planHeight is the plan height of the proposal. The default is
		// originalHeight.
		planHeight int64
		// voters is how many validators vote Yes. The default is all of them.
		voters int
		// before runs after the votes, before EndBlock. It is optional.
		before          func(f *fixture, id uint64)
		expectedStarted bool
		// expectedReason is the reason of fastupgrade_not_started. With no
		// reason, a case with no timer must emit no event.
		expectedReason string
	}{
		{
			name: "should not start the timer - pending plan",
			before: func(f *fixture, _ uint64) {
				require.NoError(f.t, f.app.UpgradeKeeper.ScheduleUpgrade(f.ctx, upgradetypes.Plan{Name: "other", Height: 5000}))
			},
		},
		{
			name: "should not start the timer - natural deadline due",
			before: func(f *fixture, id uint64) {
				f.setBlock(f.ctx.BlockHeight(), *f.proposal(id).VotingEndTime)
			},
		},
		{
			name:           "should not start the timer - gov params incompatible",
			before:         incompatibleGov,
			expectedReason: "incompatible_with_gov",
		},
		{
			// The gate fails first, and that is silent.
			name:   "should not start the timer - gov params incompatible below the gate",
			voters: 4,
			before: incompatibleGov,
		},
		{
			name:            "should start the timer - plan height is the final check",
			planHeight:      finalCheck,
			expectedStarted: true,
		},
		{
			name:       "should not start the timer - plan height below the final check",
			planHeight: finalCheck - 1,
		},
		{
			// The timelock does not fit above this block height, so EndBlock
			// fails before it looks at a proposal.
			name: "should not start the timer - upgrade height overflow",
			before: func(f *fixture, _ uint64) {
				f.setBlock(math.MaxInt64-10, f.ctx.BlockTime())
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			msg := f.upgradeMsg(minorVersion)
			if tc.planHeight != 0 {
				msg.Plan.Height = tc.planHeight
			}
			if tc.voters == 0 {
				tc.voters = numValidators
			}

			id := f.propose(msg)
			f.votes(id, tc.voters, yes())
			if tc.before != nil {
				tc.before(f, id)
			}
			f.endBlock()

			f.requireStartOutcome(id, tc.expectedStarted, tc.expectedReason)
			require.Equal(t, tc.voters, f.storedVotes(id))
		})
	}
}

// The whole successful path of one proposal, with gov running after the
// module as in the app. The vote must not close before H-1. At H-1 the
// module closes it, and gov passes the proposal in the same block and
// schedules the plan at H.
func TestKeeper_EndBlock_EarlyCloseExecutesInSameBlock(t *testing.T) {
	f := newFixture(t)
	id := f.proposeMinor()
	original := f.proposal(id)
	timer := f.startTimer(id)
	H := timer.UpgradeHeight

	// H-2, a wait block: nothing is touched.
	f.advanceTo(H - 2)
	f.fullEndBlock()
	_, ok := f.timer()
	require.True(t, ok)
	require.Zero(t, f.fastUpgradeEvents())
	require.True(t, f.queued(*original.VotingEndTime, id))

	// H-1, the final check.
	f.nextBlock()
	require.Equal(t, H-1, f.ctx.BlockHeight())
	f.fullEndBlock()

	// The module removed the timer and announced the early close.
	_, ok = f.timer()
	require.False(t, ok)
	f.requireSingleEvent(types.EventTypeEarlyClose, map[string]string{
		types.AttributeKeyProposalID:    firstProposalID,
		types.AttributeKeyPlanName:      minorVersion,
		types.AttributeKeyUpgradeHeight: upgradeHeightAttr,
	})

	// Gov passed the proposal in this block, with the new end time, the plan
	// height H, and all the votes. Its queue key is gone.
	p := f.proposal(id)
	require.Equal(t, govv1.StatusPassed, p.Status)
	require.Equal(t, f.ctx.BlockTime(), *p.VotingEndTime)
	require.Equal(t, H, f.planHeight(p))
	require.Equal(t, tokens.MulRaw(numValidators).String(), p.FinalTallyResult.YesCount)
	require.False(t, f.queued(*original.VotingEndTime, id))
	require.False(t, f.queued(f.ctx.BlockTime(), id))

	// x/upgrade has the plan at H. Only the height changed.
	plan, err := f.app.UpgradeKeeper.GetUpgradePlan(f.ctx)
	require.NoError(t, err)
	require.Equal(t, upgradetypes.Plan{Name: minorVersion, Height: H, Info: "info-" + minorVersion}, plan)
}

// A No vote during the timelock does not stop the early close while Yes
// power still meets the threshold.
func TestKeeper_EndBlock_NoVoteAtThresholdStillClosesEarly(t *testing.T) {
	f := newFixture(t)
	f.setParams(true, 150, "0.8")
	id := f.proposeMinor()
	timer := f.startTimer(id)

	f.nextBlock()
	f.vote(id, f.vals[0], no())

	f.advanceTo(timer.UpgradeHeight - 1)
	f.fullEndBlock()

	f.requireSingleEvent(types.EventTypeEarlyClose, nil)
	p := f.proposal(id)
	require.Equal(t, govv1.StatusPassed, p.Status)
	require.Equal(t, timer.UpgradeHeight, f.planHeight(p))
}

// Gov changes the params while a timer runs. A new timelock must not move
// the H of that timer, and a new threshold must apply at its final check.
func TestKeeper_EndBlock_TimelockAndThresholdUpdatesDuringTimer(t *testing.T) {
	// 4 of 5 Yes is exactly the threshold 0.8, so a timer starts.
	f := newFixture(t)
	f.setParams(true, 150, "0.8")
	id := f.proposeMinor()
	f.votes(id, 4, yes())
	f.endBlock()
	timer, ok := f.timer()
	require.True(t, ok)

	// The next block has a timelock of 10 and a threshold of 0.85. The timer
	// keeps its H.
	f.nextBlock()
	f.setParams(true, 10, "0.85")
	f.endBlock()
	unchanged, ok := f.timer()
	require.True(t, ok)
	require.Equal(t, timer, unchanged)

	// The final check uses the new threshold: 0.8 is below 0.85, so the
	// timer is dropped and stock gov continues.
	f.advanceTo(timer.UpgradeHeight - 1)
	f.fullEndBlock()
	f.requireSingleEvent(types.EventTypeDropped, map[string]string{
		types.AttributeKeyProposalID:    firstProposalID,
		types.AttributeKeyUpgradeHeight: upgradeHeightAttr,
		types.AttributeKeyReason:        "gate_below_threshold",
	})
	require.Equal(t, govv1.StatusVotingPeriod, f.proposal(id).Status)
}

// Something changes during the timelock, so the final check at H-1 must drop
// the timer. A drop leaves the proposal as it was: the stock gov vote goes on
// with the original plan height and deadline.
func TestKeeper_EndBlock_FinalCheckDropsKeepOriginalProposal(t *testing.T) {
	tt := []struct {
		name           string
		expectedReason string
		// during runs after the timer started, before H-1.
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
			// The gate fails first, so the event has its reason.
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
				f.applyUpgrade("v1.3.0")
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
			// Yes is 5 of 6 stakes, 0.833: gov would pass it, and it is
			// below our 0.85.
			name:           "should drop the timer - delegator lowers gov yes count",
			expectedReason: "tally_below_threshold",
			during: func(f *fixture, id uint64) {
				f.delegate(f.other, f.vals[0], tokens)
				f.vote(id, f.other, no())
			},
		},
		{
			// The veto is 3 of 8 stakes, 0.375: above gov's veto threshold
			// of 0.334, so gov would reject it.
			name:           "should drop the timer - delegator vetoes",
			expectedReason: "gov_tally_rejected",
			during: func(f *fixture, id uint64) {
				f.delegate(f.other, f.vals[0], tokens.MulRaw(3))
				f.vote(id, f.other, govv1.NewNonSplitVoteOption(govv1.OptionNoWithVeto))
			},
		},
		{
			// The proposal's own deadline wins over the postponement.
			name:           "should drop the timer - natural deadline due and another proposal due",
			expectedReason: "natural_deadline",
			during: func(f *fixture, _ uint64) {
				period := time.Second
				f.setGovParams(func(p *govv1.Params) { p.VotingPeriod = &period })
				f.propose(f.upgradeMsg("v2.0.0"))
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

			// A wait block with the change of the case. The timer stays:
			// only the final check decides.
			f.nextBlock()
			tc.during(f, id)
			votes := f.storedVotes(id)
			f.endBlock()
			_, ok := f.timer()
			require.True(t, ok)

			// H-1, the final check: the module drops the timer.
			f.advanceTo(timer.UpgradeHeight - 1)
			if tc.atFinal != nil {
				tc.atFinal(f, id)
			}
			f.endBlock()
			f.requireDropped(timer, tc.expectedReason)

			// The proposal is untouched: every field, the plan height, the
			// queue key, and the votes.
			p := f.proposal(id)
			require.Equal(t, original, p)
			require.Equal(t, originalHeight, f.planHeight(p))
			require.True(t, f.queued(*original.VotingEndTime, id))
			require.Equal(t, votes, f.storedVotes(id))
		})
	}
}

// The proposal's own deadline arrives in block H-1, the block of the final
// check. The deadline wins: the module drops the timer, and gov passes the
// proposal in the same block with the original plan height.
func TestKeeper_EndBlock_NaturalDeadlineAtFinalCheckUsesOriginalHeight(t *testing.T) {
	f := newFixture(t)
	id := f.proposeMinor()
	timer := f.startTimer(id)

	// H-1, and the block time is the deadline of the proposal.
	f.advanceTo(timer.UpgradeHeight - 1)
	f.setBlock(f.ctx.BlockHeight(), *f.proposal(id).VotingEndTime)
	f.fullEndBlock()

	// The module dropped the timer and did not close the vote early.
	f.requireDropped(timer, "natural_deadline")

	// Gov passed the proposal, and x/upgrade has the plan at the original
	// height, not at H.
	p := f.proposal(id)
	require.Equal(t, govv1.StatusPassed, p.Status)
	plan, err := f.app.UpgradeKeeper.GetUpgradePlan(f.ctx)
	require.NoError(t, err)
	require.Equal(t, originalHeight, plan.Height)
}

// The four conditions that the module checks in every block, not only at
// H-1. Each one drops the timer in the block where it becomes true.
func TestKeeper_EndBlock_TimerDropped(t *testing.T) {
	tt := []struct {
		name           string
		expectedReason string
		// during runs in the block after the timer started.
		during func(f *fixture, id uint64, timer types.Timer)
	}{
		{
			name:           "should drop the timer - module disabled",
			expectedReason: "disabled",
			during: func(f *fixture, _ uint64, _ types.Timer) {
				f.setParams(false, 150, "0.85")
			},
		},
		{
			name:           "should drop the timer - proposal canceled",
			expectedReason: "not_voting",
			during: func(f *fixture, id uint64, _ types.Timer) {
				require.NoError(f.t, f.app.GovKeeper.CancelProposal(f.ctx, id, f.other.String()))
			},
		},
		{
			name:           "should drop the timer - final check block missed",
			expectedReason: "missed_window",
			during: func(f *fixture, _ uint64, timer types.Timer) {
				f.setBlock(timer.UpgradeHeight, f.ctx.BlockTime())
			},
		},
		{
			// The module runs before gov, so it sees the deadline in the
			// block where gov ends the vote.
			name:           "should drop the timer - deadline before the final check",
			expectedReason: "natural_deadline",
			during: func(f *fixture, id uint64, _ types.Timer) {
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
			tc.during(f, id, timer)
			f.endBlock()

			f.requireDropped(timer, tc.expectedReason)

			// The drop is final: the next block has no second event and
			// no new timer.
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

// The module's own code fails at the final check. EndBlock must not halt the
// chain: it returns nil, drops only the timer, and discards the work of the
// block, so the proposal stays as it was.
//
// Both failures are forced on purpose. The code under test exists for
// failures that nobody foresaw, so the test has to make one.
func TestKeeper_EndBlock_FailureClearsOnlyTheTimer(t *testing.T) {
	tt := []struct {
		name    string
		failure func(f *fixture)
	}{
		{
			// A panic in gov's code, and a state of the fixture only. Its
			// validators have MinSelfDelegation 0, so a removed one stays
			// bonded with zero shares until staking's EndBlocker, and gov's
			// Tally divides by those shares. With MinSelfDelegation 1, as
			// x/poa sets it, staking jails the removed validator and the
			// tally does not see it.
			name: "should drop the timer - panic when gov tally divides by zero shares",
			failure: func(f *fixture) {
				require.NoError(f.t, f.app.PoaKeeper.ExecuteRemoveValidator(f.ctx, f.vals[0].String()))
			},
		},
		{
			// An error after a write. Gov's SetProposal fails after
			// endVotingNow removed the queue key. The proposal is still
			// queued only because the module discards the scratch copy.
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

			// H-1, the final check, with the failure in place.
			f.advanceTo(timer.UpgradeHeight - 1)
			tc.failure(f)

			// endBlock requires that EndBlock returns nil, and a panic
			// would fail the test: the chain does not halt.
			f.endBlock()

			// Only the timer is gone. The proposal is as it was before
			// the block.
			f.requireDropped(timer, "internal_error")
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
			// The forced panic of FailureClearsOnlyTheTimer, here in a start
			// check. It has no timer to retract, so it emits nothing.
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
			f.requireStartOutcome(id, tc.expectedTimer, "")
		})
	}
}

// An expedited proposal that becomes regular loses its timer at the
// expedited deadline. Gov deleted its votes, so new Yes votes start a new
// timer with a new H, and the vote closes early at the new H-1.
func TestKeeper_EndBlock_ExpeditedConvertedToRegularDuringTimer(t *testing.T) {
	f := newFixture(t)

	// An expedited proposal with all Yes votes: a timer starts.
	p, err := f.app.GovKeeper.SubmitProposal(f.ctx, []sdk.Msg{f.upgradeMsg(minorVersion)}, "", "title", "summary", f.other, true)
	require.NoError(t, err)
	require.NoError(t, f.app.GovKeeper.ActivateVotingPeriod(f.ctx, p))
	id := p.Id
	first := f.startTimer(id)

	// Two validators change to No. Yes is 3 of 5, 0.6: below gov's
	// expedited threshold of 0.667. This is a wait block, so nothing
	// happens yet.
	f.nextBlock()
	f.vote(id, f.vals[0], no())
	f.vote(id, f.vals[1], no())
	f.fullEndBlock()

	// The expedited deadline comes before H-1. The module drops the timer,
	// and gov's expedited tally fails in the same block.
	f.nextBlock()
	expeditedEnd := *f.proposal(id).VotingEndTime
	f.setBlock(f.ctx.BlockHeight(), expeditedEnd)
	f.fullEndBlock()
	f.requireDropped(first, "natural_deadline")

	// Gov made the proposal regular, with a later deadline and no votes.
	converted := f.proposal(id)
	require.False(t, converted.Expedited)
	require.Equal(t, govv1.StatusVotingPeriod, converted.Status)
	require.True(t, converted.VotingEndTime.After(expeditedEnd))
	require.Zero(t, f.storedVotes(id))

	// All validators vote Yes again: a new timer starts, and its H counts
	// from this block.
	f.nextBlock()
	second := f.startTimer(id)
	require.Equal(t, f.ctx.BlockHeight()+150, second.UpgradeHeight)

	// The new H-1, the final check.
	f.advanceTo(second.UpgradeHeight - 1)
	f.fullEndBlock()

	// The vote closed early. Gov passed the proposal with the new H, and
	// its queue key at the regular deadline is gone.
	f.requireSingleEvent(types.EventTypeEarlyClose, nil)
	passed := f.proposal(id)
	require.Equal(t, govv1.StatusPassed, passed.Status)
	require.Equal(t, second.UpgradeHeight, f.planHeight(passed))
	require.False(t, f.queued(*converted.VotingEndTime, id))
}

// Another proposal ends in block H-1. Gov executes it before it tallies
// ours, so gov's result could differ from the module's check. The module
// must not close the vote there: it moves H later by postpone_blocks and
// closes the vote at the new H-1.
func TestKeeper_EndBlock_OtherProposalDueAtFinalCheckPostponesTimer(t *testing.T) {
	const postponeBlocks = 7

	// The upgrade proposal gets its timer.
	f := newFixture(t)
	params := types.DefaultParams()
	params.Enabled = true
	params.PostponeBlocks = postponeBlocks
	require.NoError(t, f.k.SetParams(f.ctx, params))
	id := f.proposeMinor()
	timer := f.startTimer(id)
	original := f.proposal(id)

	// A second proposal removes a validator, and its voting period ends in
	// block H-1. Blocks are one second apart. The millisecond puts the
	// deadline inside that second and not on its edge.
	finalCheck := timer.UpgradeHeight - 1
	period := time.Duration(finalCheck-f.ctx.BlockHeight())*time.Second - time.Millisecond
	f.setGovParams(func(p *govv1.Params) { p.VotingPeriod = &period })
	removal := f.propose(&poatypes.MsgRemoveValidator{Authority: f.govAddr, ValidatorAddress: f.vals[4].String()})
	f.votes(removal, numValidators, yes())

	// H-1: the module runs, then gov. Staking's EndBlocker runs after gov
	// on a chain, and it takes the removed validator out of the set.
	f.advanceTo(finalCheck)
	f.fullEndBlock()
	_, err := f.app.StakingKeeper.EndBlocker(f.ctx)
	require.NoError(t, err)

	// The module moved H and announced it. Gov executed the removal, and
	// the upgrade proposal is untouched.
	H := timer.UpgradeHeight + postponeBlocks
	postponedHeightAttr := strconv.FormatInt(H, 10)
	f.requireSingleEvent(types.EventTypePostponed, map[string]string{
		types.AttributeKeyProposalID:    firstProposalID,
		types.AttributeKeyUpgradeHeight: postponedHeightAttr,
	})
	postponed, ok := f.timer()
	require.True(t, ok)
	require.Equal(t, types.Timer{ProposalId: id, UpgradeHeight: H}, postponed)
	require.Equal(t, govv1.StatusPassed, f.proposal(removal).Status)
	require.Equal(t, original, f.proposal(id))

	// The new H-1: no other proposal ends here, so the vote closes early
	// with the new H.
	f.advanceTo(H - 1)
	f.fullEndBlock()
	f.requireSingleEvent(types.EventTypeEarlyClose, map[string]string{types.AttributeKeyUpgradeHeight: postponedHeightAttr})
	p := f.proposal(id)
	require.Equal(t, govv1.StatusPassed, p.Status)
	require.Equal(t, H, f.planHeight(p))
}
