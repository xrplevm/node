package integration

import (
	"errors"
	"time"

	"cosmossdk.io/collections"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govprecompile "github.com/cosmos/evm/precompiles/gov"
	"github.com/cosmos/evm/testutil/integration/base/factory"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"

	v12 "github.com/xrplevm/node/v12/app/upgrades/v12"
	"github.com/xrplevm/node/v12/testutil/integration/exrp/utils"
	fastupgradetypes "github.com/xrplevm/node/v12/x/fastupgrade/types"
	poatypes "github.com/xrplevm/node/v12/x/poa/types"
)

const fastUpgradePlanHeight = int64(1_000_000)

func (s *TestSuite) govAuthority() string {
	return sdktypes.AccAddress(address.Module("gov")).String()
}

// passProposal submits a proposal with msg, has every validator vote Yes in
// one block, and commits the block after the voting period, where gov
// executes it. Blocks: submit (h+1, h+2), votes (h+3), voting end (h+4).
func (s *TestSuite) passProposal(title string, msg sdktypes.Msg) {
	id, err := utils.SubmitProposal(s.factory, s.Network(), s.keyring.GetKey(0).Priv, title, msg)
	s.Require().NoError(err)
	s.voteInOneBlock(id, govv1.OptionYes, s.allValidators()...)

	govParams, err := s.Network().GetGovClient().Params(s.Network().GetContext(), &govv1.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().NoError(s.Network().NextBlockAfter(*govParams.Params.VotingPeriod))
	s.Require().Equal(govv1.StatusPassed, s.committedProposal(id).Status)
}

// applyV12Upgrade passes and executes a real v12.0.0 software upgrade
// proposal, so the v12 handler enables x/fastupgrade and v12.0.0 becomes the
// last completed upgrade.
func (s *TestSuite) applyV12Upgrade() {
	// This binary already has the v12.0.0 handler, so x/upgrade panics if the
	// plan is pending before its height: gov schedules the plan at h+4 and
	// the upgrade runs at h+5.
	planHeight := s.Network().GetContext().BlockHeight() + 5
	s.passProposal("v12", &upgradetypes.MsgSoftwareUpgrade{
		Authority: s.govAuthority(),
		Plan:      upgradetypes.Plan{Name: v12.UpgradeName, Height: planHeight},
	})
	s.Require().NoError(s.Network().NextBlock())
	s.Require().Equal(planHeight, s.Network().GetContext().BlockHeight())

	ctx := s.Network().GetContext()
	doneHeight, err := s.Network().UpgradeKeeper().GetDoneHeight(ctx, v12.UpgradeName)
	s.Require().NoError(err)
	s.Require().Equal(planHeight, doneHeight)

	params, err := s.Network().FastUpgradeKeeper().Params.Get(ctx)
	s.Require().NoError(err)
	s.Require().Equal(fastupgradetypes.Params{Enabled: true, TimelockBlocks: 150, Threshold: "0.850000000000000000", PostponeBlocks: 10}, params)
}

// activateGovPrecompile makes the gov precompile the only active static
// precompile, through governance.
func (s *TestSuite) activateGovPrecompile() {
	res, err := s.Network().GetEvmClient().Params(s.Network().GetContext(), &evmtypes.QueryParamsRequest{})
	s.Require().NoError(err)
	params := res.Params
	params.ActiveStaticPrecompiles = []string{evmtypes.GovPrecompileAddress}
	s.passProposal("gov precompile", &evmtypes.MsgUpdateParams{Authority: s.govAuthority(), Params: params})
}

func (s *TestSuite) proposeMinorUpgrade() uint64 {
	id, err := utils.SubmitProposal(s.factory, s.Network(), s.keyring.GetKey(0).Priv, "v12.1.0", &upgradetypes.MsgSoftwareUpgrade{
		Authority: s.govAuthority(),
		Plan:      upgradetypes.Plan{Name: "v12.1.0", Height: fastUpgradePlanHeight, Info: "binaries"},
	})
	s.Require().NoError(err)
	return id
}

// commitInOneBlock commits one block holding txs and requires each to succeed.
func (s *TestSuite) commitInOneBlock(txs [][]byte) *abcitypes.ResponseFinalizeBlock {
	res, err := s.Network().NextBlockWithTxs(txs...)
	s.Require().NoError(err)
	for _, txRes := range res.TxResults {
		s.Require().True(txRes.IsOK(), txRes.Log)
	}
	return res
}

// voteInOneBlock commits one block holding a vote from each given key index.
func (s *TestSuite) voteInOneBlock(id uint64, option govv1.VoteOption, indexes ...int) *abcitypes.ResponseFinalizeBlock {
	txs := make([][]byte, 0, len(indexes))
	for _, i := range indexes {
		key := s.keyring.GetKey(i)
		tx, err := s.factory.BuildCosmosTx(key.Priv, factory.CosmosTxArgs{
			Msgs: []sdktypes.Msg{govv1.NewMsgVote(key.AccAddr, id, option, "")},
		})
		s.Require().NoError(err)
		bz, err := s.factory.EncodeTx(tx)
		s.Require().NoError(err)
		txs = append(txs, bz)
	}
	return s.commitInOneBlock(txs)
}

// precompileYesInOneBlock commits one block holding a Yes vote from each
// given key index, sent as an EVM transaction to the gov precompile.
func (s *TestSuite) precompileYesInOneBlock(id uint64, indexes ...int) *abcitypes.ResponseFinalizeBlock {
	gov := common.HexToAddress(evmtypes.GovPrecompileAddress)
	txs := make([][]byte, 0, len(indexes))
	for _, i := range indexes {
		key := s.keyring.GetKey(i)
		input, err := govprecompile.ABI.Pack(govprecompile.VoteMethod, key.Addr, id, uint8(govv1.OptionYes), "")
		s.Require().NoError(err)
		// An explicit gas limit: the factory would estimate it by running the call.
		tx, err := s.factory.GenerateSignedEthTx(key.Priv, evmtypes.EvmTxArgs{To: &gov, GasLimit: 500_000, Input: input})
		s.Require().NoError(err)
		bz, err := s.factory.EncodeTx(tx)
		s.Require().NoError(err)
		txs = append(txs, bz)
	}
	return s.commitInOneBlock(txs)
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
	s.Require().NoError(err)
	return len(res.Votes)
}

func (s *TestSuite) committedProposal(id uint64) *govv1.Proposal {
	res, err := s.Network().GetGovClient().Proposal(s.Network().GetContext(), &govv1.QueryProposalRequest{ProposalId: id})
	s.Require().NoError(err)
	return res.Proposal
}

func (s *TestSuite) committedTimer() (fastupgradetypes.Timer, bool) {
	timer, err := s.Network().FastUpgradeKeeper().Timer.Get(s.Network().GetContext())
	if errors.Is(err, collections.ErrNotFound) {
		return fastupgradetypes.Timer{}, false
	}
	s.Require().NoError(err)
	return timer, true
}

func (s *TestSuite) advanceToHeight(height int64) *abcitypes.ResponseFinalizeBlock {
	var res *abcitypes.ResponseFinalizeBlock
	for s.Network().GetContext().BlockHeight() < height {
		var err error
		res, err = s.Network().NextBlockWithTxs()
		s.Require().NoError(err)
	}
	return res
}

// TestFastUpgrade_EarlyCloseInCommittedBlocks drives the fast path through
// committed blocks: validators vote in one block, the timer starts in that
// block's EndBlock, and at H-1 gov executes the repacked plan in the same block.
func (s *TestSuite) TestFastUpgrade_EarlyCloseInCommittedBlocks() {
	s.applyV12Upgrade()
	id := s.proposeMinorUpgrade()

	res := s.voteInOneBlock(id, govv1.OptionYes, s.allValidators()...)
	voteHeight := s.Network().GetContext().BlockHeight()
	H := voteHeight + 150

	started := blockEvents(res, fastupgradetypes.EventTypeStarted)
	s.Require().Len(started, 1)
	s.Require().Equal("v12.1.0", started[0][fastupgradetypes.AttributeKeyPlanName])
	s.Require().Equal("binaries", started[0][fastupgradetypes.AttributeKeyPlanInfo])

	timer, ok := s.committedTimer()
	s.Require().True(ok)
	s.Require().Equal(fastupgradetypes.Timer{ProposalId: id, UpgradeHeight: H}, timer)

	// The nomination existed during the vote block's EndBlock (the timer
	// started), and that EndBlock cleared it before the commit.
	has, err := s.Network().FastUpgradeKeeper().Nominations.Has(s.Network().GetContext(), id)
	s.Require().NoError(err)
	s.Require().False(has)

	// The start tally ran on a scratch copy: every vote is still committed.
	s.Require().Equal(numValidators, s.committedVotes(id))

	res = s.advanceToHeight(H - 2)
	s.Require().Empty(blockEvents(res, fastupgradetypes.EventTypeEarlyClose))
	s.Require().Equal(govv1.StatusVotingPeriod, s.committedProposal(id).Status)

	res = s.advanceToHeight(H - 1)
	closed := blockEvents(res, fastupgradetypes.EventTypeEarlyClose)
	s.Require().Len(closed, 1)

	var passedInSameBlock bool
	for _, e := range blockEvents(res, "active_proposal") {
		if e["proposal_id"] == closed[0][fastupgradetypes.AttributeKeyProposalID] && e["proposal_result"] == "proposal_passed" {
			passedInSameBlock = true
		}
	}
	s.Require().True(passedInSameBlock)

	p := s.committedProposal(id)
	s.Require().Equal(govv1.StatusPassed, p.Status)
	s.Require().Equal(sdktypes.TokensFromConsensusPower(numValidators, sdktypes.DefaultPowerReduction).String(), p.FinalTallyResult.YesCount)

	plan, err := s.Network().UpgradeKeeper().GetUpgradePlan(s.Network().GetContext())
	s.Require().NoError(err)
	s.Require().Equal(upgradetypes.Plan{Name: "v12.1.0", Height: H, Info: "binaries"}, plan)

	_, ok = s.committedTimer()
	s.Require().False(ok)
}

// TestFastUpgrade_GovPrecompileVotesStartTimer sends the validators' Yes votes
// as EVM transactions to the gov precompile. The precompile runs MsgVote on
// the EVM's own store, which holds only the KV stores, so the vote hook must
// store its nomination there.
func (s *TestSuite) TestFastUpgrade_GovPrecompileVotesStartTimer() {
	s.applyV12Upgrade()
	s.activateGovPrecompile()
	id := s.proposeMinorUpgrade()

	res := s.precompileYesInOneBlock(id, s.allValidators()...)
	voteHeight := s.Network().GetContext().BlockHeight()

	s.Require().Len(blockEvents(res, fastupgradetypes.EventTypeStarted), 1)
	timer, ok := s.committedTimer()
	s.Require().True(ok)
	s.Require().Equal(fastupgradetypes.Timer{ProposalId: id, UpgradeHeight: voteHeight + 150}, timer)
	s.Require().Equal(numValidators, s.committedVotes(id))
}

// TestFastUpgrade_VoteChangeFallsBackToStockGov checks the veto in committed
// blocks: a validator switches to No during the timelock, the timer drops at
// H-1, and gov later tallies every vote at the original deadline and height.
func (s *TestSuite) TestFastUpgrade_VoteChangeFallsBackToStockGov() {
	s.applyV12Upgrade()
	id := s.proposeMinorUpgrade()
	original := s.committedProposal(id)

	s.voteInOneBlock(id, govv1.OptionYes, s.allValidators()...)
	timer, ok := s.committedTimer()
	s.Require().True(ok)

	s.voteInOneBlock(id, govv1.OptionNo, numValidators-1)

	res := s.advanceToHeight(timer.UpgradeHeight - 1)
	dropped := blockEvents(res, fastupgradetypes.EventTypeDropped)
	s.Require().Len(dropped, 1)
	s.Require().Equal("gate-below-threshold", dropped[0][fastupgradetypes.AttributeKeyReason])
	s.Require().Empty(blockEvents(res, fastupgradetypes.EventTypeEarlyClose))

	p := s.committedProposal(id)
	s.Require().Equal(govv1.StatusVotingPeriod, p.Status)
	s.Require().Equal(original.VotingEndTime, p.VotingEndTime)
	s.Require().Equal(original.Messages, p.Messages)
	s.Require().Equal(numValidators, s.committedVotes(id))
	_, err := s.Network().UpgradeKeeper().GetUpgradePlan(s.Network().GetContext())
	s.Require().ErrorIs(err, upgradetypes.ErrNoUpgradePlanFound)

	gq := s.Network().GetGovClient()
	params, err := gq.Params(s.Network().GetContext(), &govv1.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().NoError(s.Network().NextBlockAfter(*params.Params.VotingPeriod))
	s.Require().NoError(s.Network().NextBlock())

	p = s.committedProposal(id)
	s.Require().Equal(govv1.StatusPassed, p.Status)
	power := sdktypes.TokensFromConsensusPower(1, sdktypes.DefaultPowerReduction)
	s.Require().Equal(power.MulRaw(numValidators-1).String(), p.FinalTallyResult.YesCount)
	s.Require().Equal(power.String(), p.FinalTallyResult.NoCount)

	plan, err := s.Network().UpgradeKeeper().GetUpgradePlan(s.Network().GetContext())
	s.Require().NoError(err)
	s.Require().Equal(fastUpgradePlanHeight, plan.Height)
	s.Require().Equal("v12.1.0", plan.Name)
}

// TestFastUpgrade_RemovalAtFinalCheckPostponesTimer commits real blocks. A
// validator removal whose deadline falls in block H-1 executes in gov's
// EndBlocker of that block. The timer moves to a later H, where the vote
// closes on the smaller validator set.
func (s *TestSuite) TestFastUpgrade_RemovalAtFinalCheckPostponesTimer() {
	s.applyV12Upgrade()

	removal, err := utils.SubmitProposal(s.factory, s.Network(), s.keyring.GetKey(0).Priv, "remove", &poatypes.MsgRemoveValidator{
		Authority:        s.govAuthority(),
		ValidatorAddress: s.keyring.GetKey(numValidators - 1).AccAddr.String(),
	})
	s.Require().NoError(err)
	// The upgrade's own deadline must stay after the postponed final check.
	s.Require().NoError(s.Network().NextBlockAfter(time.Minute))

	id := s.proposeMinorUpgrade()
	s.voteInOneBlock(removal, govv1.OptionYes, s.allValidators()...)
	s.voteInOneBlock(id, govv1.OptionYes, s.allValidators()...)
	timer, ok := s.committedTimer()
	s.Require().True(ok)

	s.advanceToHeight(timer.UpgradeHeight - 2)
	removalEnd := *s.committedProposal(removal).VotingEndTime
	now := s.Network().GetContext().BlockTime()
	s.Require().True(removalEnd.After(now))

	// Block H-1 lands on the removal's deadline.
	s.Require().NoError(s.Network().NextBlockAfter(removalEnd.Sub(now)))
	s.Require().Equal(govv1.StatusPassed, s.committedProposal(removal).Status)
	s.Require().Equal(govv1.StatusVotingPeriod, s.committedProposal(id).Status)
	H := timer.UpgradeHeight + 10
	postponed, ok := s.committedTimer()
	s.Require().True(ok)
	s.Require().Equal(fastupgradetypes.Timer{ProposalId: id, UpgradeHeight: H}, postponed)

	res := s.advanceToHeight(H - 1)
	s.Require().Len(blockEvents(res, fastupgradetypes.EventTypeEarlyClose), 1)
	s.Require().Equal(govv1.StatusPassed, s.committedProposal(id).Status)
	plan, err := s.Network().UpgradeKeeper().GetUpgradePlan(s.Network().GetContext())
	s.Require().NoError(err)
	s.Require().Equal(H, plan.Height)
}
