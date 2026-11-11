package integration

import (
	"context"
	"errors"
	"time"

	"cosmossdk.io/collections"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/module"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/evm/testutil/integration/base/factory"
	"github.com/stretchr/testify/require"

	"github.com/xrplevm/node/v12/testutil/integration/exrp/utils"
	fastupgradetypes "github.com/xrplevm/node/v12/x/fastupgrade/types"
	poatypes "github.com/xrplevm/node/v12/x/poa/types"
)

const (
	// Test-only upgrade names: a base version and a minor bump of it.
	fastUpgradeBaseName   = "v1.0.0"
	fastUpgradeMinorName  = "v1.1.0"
	fastUpgradePlanInfo   = "binaries"
	fastUpgradePlanHeight = int64(1_000_000)
)

func (s *TestSuite) govAuthority() string {
	return sdktypes.AccAddress(address.Module("gov")).String()
}

// passProposal submits a proposal with msgs, has every validator vote Yes in
// one block, and commits the block after the voting period, where gov
// executes it. Blocks: submit (h+1, h+2), votes (h+3), voting end (h+4).
func (s *TestSuite) passProposal(title string, msgs ...sdktypes.Msg) {
	id, err := utils.SubmitProposal(s.factory, s.Network(), s.keyring.GetKey(0).Priv, title, msgs...)
	require.NoError(s.T(), err)
	s.voteInOneBlock(id, govv1.OptionYes, s.allValidators()...)

	require.NoError(s.T(), s.Network().NextBlockAfter(votingPeriod))
	require.Equal(s.T(), govv1.StatusPassed, s.committedProposal(id).Status)
}

// enableFastUpgrade enables x/fastupgrade and completes a no-op upgrade named
// fastUpgradeBaseName. The fast path accepts only a minor bump of the last
// completed upgrade, and a new chain has no completed upgrade.
func (s *TestSuite) enableFastUpgrade() {
	s.Network().UpgradeKeeper().SetUpgradeHandler(fastUpgradeBaseName, func(_ context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		return vm, nil
	})
	params := fastupgradetypes.DefaultParams()
	params.Enabled = true

	// x/upgrade panics if a plan with a handler is pending before its
	// height: gov schedules the plan at h+4 and the upgrade runs at h+5.
	planHeight := s.Network().GetContext().BlockHeight() + 5
	s.passProposal("setup",
		&fastupgradetypes.MsgUpdateParams{Authority: s.govAuthority(), Params: params},
		&upgradetypes.MsgSoftwareUpgrade{
			Authority: s.govAuthority(),
			Plan:      upgradetypes.Plan{Name: fastUpgradeBaseName, Height: planHeight},
		},
	)
	require.NoError(s.T(), s.Network().NextBlock())

	name, height, err := s.Network().UpgradeKeeper().GetLastCompletedUpgrade(s.Network().GetContext())
	require.NoError(s.T(), err)
	require.Equal(s.T(), fastUpgradeBaseName, name)
	require.Equal(s.T(), planHeight, height)
}

func (s *TestSuite) proposeMinorUpgrade() uint64 {
	id, err := utils.SubmitProposal(s.factory, s.Network(), s.keyring.GetKey(0).Priv, fastUpgradeMinorName, &upgradetypes.MsgSoftwareUpgrade{
		Authority: s.govAuthority(),
		Plan:      upgradetypes.Plan{Name: fastUpgradeMinorName, Height: fastUpgradePlanHeight, Info: fastUpgradePlanInfo},
	})
	require.NoError(s.T(), err)
	return id
}

// voteInOneBlock commits one block holding a vote from each given key index,
// and requires each vote to succeed.
func (s *TestSuite) voteInOneBlock(id uint64, option govv1.VoteOption, indexes ...int) *abcitypes.ResponseFinalizeBlock {
	txs := make([][]byte, 0, len(indexes))
	for _, i := range indexes {
		key := s.keyring.GetKey(i)
		tx, err := s.factory.BuildCosmosTx(key.Priv, factory.CosmosTxArgs{
			Msgs: []sdktypes.Msg{govv1.NewMsgVote(key.AccAddr, id, option, "")},
		})
		require.NoError(s.T(), err)
		bz, err := s.factory.EncodeTx(tx)
		require.NoError(s.T(), err)
		txs = append(txs, bz)
	}
	res, err := s.Network().NextBlockWithTxs(txs...)
	require.NoError(s.T(), err)
	for _, txRes := range res.TxResults {
		require.True(s.T(), txRes.IsOK(), txRes.Log)
	}
	return res
}

func (s *TestSuite) allValidators() []int {
	indexes := make([]int, numValidators)
	for i := range indexes {
		indexes[i] = i
	}
	return indexes
}

func blockEvents(res *abcitypes.ResponseFinalizeBlock, eventType string) []map[string]string {
	out := make([]map[string]string, 0, len(res.Events))
	for _, e := range res.Events {
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

func (s *TestSuite) committedVotes(id uint64) int {
	gq := s.Network().GetGovClient()
	res, err := gq.Votes(s.Network().GetContext(), &govv1.QueryVotesRequest{ProposalId: id})
	require.NoError(s.T(), err)
	return len(res.Votes)
}

func (s *TestSuite) committedProposal(id uint64) *govv1.Proposal {
	res, err := s.Network().GetGovClient().Proposal(s.Network().GetContext(), &govv1.QueryProposalRequest{ProposalId: id})
	require.NoError(s.T(), err)
	return res.Proposal
}

func (s *TestSuite) committedTimer() (fastupgradetypes.Timer, bool) {
	timer, err := s.Network().FastUpgradeKeeper().Timer.Get(s.Network().GetContext())
	if errors.Is(err, collections.ErrNotFound) {
		return fastupgradetypes.Timer{}, false
	}
	require.NoError(s.T(), err)
	return timer, true
}

func (s *TestSuite) advanceToHeight(height int64) *abcitypes.ResponseFinalizeBlock {
	var res *abcitypes.ResponseFinalizeBlock
	for s.Network().GetContext().BlockHeight() < height {
		var err error
		res, err = s.Network().NextBlockWithTxs()
		require.NoError(s.T(), err)
	}
	return res
}

// TestFastUpgrade_EarlyCloseInCommittedBlocks drives the fast path through
// committed blocks: validators vote in one block, the timer starts in that
// block's EndBlock, and at H-1 gov executes the repacked plan in the same
// block. H-1 is one timelock after the vote, long before votingPeriod ends.
func (s *TestSuite) TestFastUpgrade_EarlyCloseInCommittedBlocks() {
	s.enableFastUpgrade()
	id := s.proposeMinorUpgrade()

	res := s.voteInOneBlock(id, govv1.OptionYes, s.allValidators()...)
	voteHeight := s.Network().GetContext().BlockHeight()
	H := voteHeight + int64(fastupgradetypes.DefaultTimelockBlocks)

	started := blockEvents(res, fastupgradetypes.EventTypeStarted)
	require.Len(s.T(), started, 1)
	require.Equal(s.T(), fastUpgradeMinorName, started[0][fastupgradetypes.AttributeKeyPlanName])
	require.Equal(s.T(), fastUpgradePlanInfo, started[0][fastupgradetypes.AttributeKeyPlanInfo])

	timer, ok := s.committedTimer()
	require.True(s.T(), ok)
	require.Equal(s.T(), fastupgradetypes.Timer{ProposalId: id, UpgradeHeight: H}, timer)

	// The nomination existed during the vote block's EndBlock (the timer
	// started), and that EndBlock cleared it before the commit.
	has, err := s.Network().FastUpgradeKeeper().Nominations.Has(s.Network().GetContext(), id)
	require.NoError(s.T(), err)
	require.False(s.T(), has)

	// The start tally ran on a scratch copy: every vote is still committed.
	require.Equal(s.T(), numValidators, s.committedVotes(id))

	res = s.advanceToHeight(H - 2)
	require.Empty(s.T(), blockEvents(res, fastupgradetypes.EventTypeEarlyClose))
	require.Equal(s.T(), govv1.StatusVotingPeriod, s.committedProposal(id).Status)

	res = s.advanceToHeight(H - 1)
	closed := blockEvents(res, fastupgradetypes.EventTypeEarlyClose)
	require.Len(s.T(), closed, 1)

	var passedInSameBlock bool
	for _, e := range blockEvents(res, govtypes.EventTypeActiveProposal) {
		if e[govtypes.AttributeKeyProposalID] == closed[0][fastupgradetypes.AttributeKeyProposalID] && e[govtypes.AttributeKeyProposalResult] == govtypes.AttributeValueProposalPassed {
			passedInSameBlock = true
		}
	}
	require.True(s.T(), passedInSameBlock)

	p := s.committedProposal(id)
	require.Equal(s.T(), govv1.StatusPassed, p.Status)
	require.Equal(s.T(), sdktypes.TokensFromConsensusPower(numValidators, sdktypes.DefaultPowerReduction).String(), p.FinalTallyResult.YesCount)

	// The test stops at H-1, with the plan scheduled at H. Block H is stock
	// x/upgrade behavior: the chain halts there until the new binary runs.
	plan, err := s.Network().UpgradeKeeper().GetUpgradePlan(s.Network().GetContext())
	require.NoError(s.T(), err)
	require.Equal(s.T(), upgradetypes.Plan{Name: fastUpgradeMinorName, Height: H, Info: fastUpgradePlanInfo}, plan)

	_, ok = s.committedTimer()
	require.False(s.T(), ok)
}

// TestFastUpgrade_VoteChangeFallsBackToStockGov checks the veto in committed
// blocks: a validator switches to No during the timelock, the timer drops at
// H-1, and gov later tallies every vote at the original deadline and height.
func (s *TestSuite) TestFastUpgrade_VoteChangeFallsBackToStockGov() {
	s.enableFastUpgrade()
	id := s.proposeMinorUpgrade()
	original := s.committedProposal(id)

	s.voteInOneBlock(id, govv1.OptionYes, s.allValidators()...)
	timer, ok := s.committedTimer()
	require.True(s.T(), ok)

	s.voteInOneBlock(id, govv1.OptionNo, numValidators-1)

	res := s.advanceToHeight(timer.UpgradeHeight - 1)
	dropped := blockEvents(res, fastupgradetypes.EventTypeDropped)
	require.Len(s.T(), dropped, 1)
	require.Equal(s.T(), string(fastupgradetypes.ReasonGateBelowThreshold), dropped[0][fastupgradetypes.AttributeKeyReason])
	require.Empty(s.T(), blockEvents(res, fastupgradetypes.EventTypeEarlyClose))
	_, ok = s.committedTimer()
	require.False(s.T(), ok)

	p := s.committedProposal(id)
	require.Equal(s.T(), govv1.StatusVotingPeriod, p.Status)
	require.Equal(s.T(), original.VotingEndTime, p.VotingEndTime)
	require.Equal(s.T(), original.Messages, p.Messages)
	require.Equal(s.T(), numValidators, s.committedVotes(id))
	_, err := s.Network().UpgradeKeeper().GetUpgradePlan(s.Network().GetContext())
	require.ErrorIs(s.T(), err, upgradetypes.ErrNoUpgradePlanFound)

	require.NoError(s.T(), s.Network().NextBlockAfter(votingPeriod))

	p = s.committedProposal(id)
	require.Equal(s.T(), govv1.StatusPassed, p.Status)
	power := sdktypes.TokensFromConsensusPower(1, sdktypes.DefaultPowerReduction)
	require.Equal(s.T(), power.MulRaw(numValidators-1).String(), p.FinalTallyResult.YesCount)
	require.Equal(s.T(), power.String(), p.FinalTallyResult.NoCount)

	plan, err := s.Network().UpgradeKeeper().GetUpgradePlan(s.Network().GetContext())
	require.NoError(s.T(), err)
	require.Equal(s.T(), fastUpgradePlanHeight, plan.Height)
	require.Equal(s.T(), fastUpgradeMinorName, plan.Name)
}

// TestFastUpgrade_RemovalAtFinalCheckPostponesTimer checks the postponement
// in committed blocks: a validator removal whose deadline falls in block H-1
// executes in gov's EndBlocker of that block, so the timer moves to a later
// H, and the vote closes at the new H-1.
func (s *TestSuite) TestFastUpgrade_RemovalAtFinalCheckPostponesTimer() {
	s.enableFastUpgrade()

	removal, err := utils.SubmitProposal(s.factory, s.Network(), s.keyring.GetKey(0).Priv, "remove", &poatypes.MsgRemoveValidator{
		Authority:        s.govAuthority(),
		ValidatorAddress: s.keyring.GetKey(numValidators - 1).AccAddr.String(),
	})
	require.NoError(s.T(), err)
	// The upgrade's own deadline must stay after the postponed final check.
	require.NoError(s.T(), s.Network().NextBlockAfter(time.Minute))

	id := s.proposeMinorUpgrade()
	s.voteInOneBlock(removal, govv1.OptionYes, s.allValidators()...)
	s.voteInOneBlock(id, govv1.OptionYes, s.allValidators()...)
	timer, ok := s.committedTimer()
	require.True(s.T(), ok)

	s.advanceToHeight(timer.UpgradeHeight - 2)
	removalEnd := *s.committedProposal(removal).VotingEndTime
	now := s.Network().GetContext().BlockTime()
	require.True(s.T(), removalEnd.After(now))

	// Block H-1 lands on the removal's deadline.
	require.NoError(s.T(), s.Network().NextBlockAfter(removalEnd.Sub(now)))
	require.Equal(s.T(), govv1.StatusPassed, s.committedProposal(removal).Status)
	require.Equal(s.T(), govv1.StatusVotingPeriod, s.committedProposal(id).Status)
	H := timer.UpgradeHeight + int64(fastupgradetypes.DefaultPostponeBlocks)
	postponed, ok := s.committedTimer()
	require.True(s.T(), ok)
	require.Equal(s.T(), fastupgradetypes.Timer{ProposalId: id, UpgradeHeight: H}, postponed)

	res := s.advanceToHeight(H - 1)
	require.Len(s.T(), blockEvents(res, fastupgradetypes.EventTypeEarlyClose), 1)
	require.Equal(s.T(), govv1.StatusPassed, s.committedProposal(id).Status)
	plan, err := s.Network().UpgradeKeeper().GetUpgradePlan(s.Network().GetContext())
	require.NoError(s.T(), err)
	require.Equal(s.T(), H, plan.Height)
}
