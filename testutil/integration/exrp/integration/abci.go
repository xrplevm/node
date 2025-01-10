// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package exrpintegration

import (
	"time"

	storetypes "cosmossdk.io/store/types"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	exrpcommon "github.com/xrplevm/node/v5/testutil/integration/exrp/common"
)

// NextBlock is a private helper function that runs the EndBlocker logic, commits the changes,
// updates the header and runs the BeginBlocker
func (n *IntegrationNetwork) NextBlock() error {
	return n.NextBlockAfter(time.Second)
}

// NextBlockAfter is a private helper function that runs the FinalizeBlock logic, updates the context and
// commits the changes to have a block time after the given duration.
func (n *IntegrationNetwork) NextBlockAfter(duration time.Duration) error {
	_, err := n.finalizeBlockAndCommit(duration, nil, nil)
	return err
}

// NextBlockWithTxs is a helper function that runs the FinalizeBlock logic
// with the provided tx bytes, updates the context and
// commits the changes to have a block time after the given duration.
func (n *IntegrationNetwork) NextBlockWithTxs(txBytes ...[]byte) (*abcitypes.ResponseFinalizeBlock, error) {
	return n.finalizeBlockAndCommit(time.Second, n.valFlags, nil, txBytes...)
}

// NextNBlocksWithValidatorFlags is a helper function that runs the FinalizeBlock logic
// with the provided validator flags, updates the context and
// commits the changes to have a block time after the given duration.
func (n *IntegrationNetwork) NextNBlocksWithValidatorFlags(blocks int64, validatorFlags []cmtproto.BlockIDFlag) error {
	for i := int64(0); i < blocks; i++ {
		_, err := n.finalizeBlockAndCommit(time.Second, validatorFlags, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// NextBlockWithMisBehaviors is a helper function that runs the FinalizeBlock logic
// with the provided misbehaviors, updates the context and
// commits the changes to have a block time after the given duration.
func (n *IntegrationNetwork) NextBlockWithMisBehaviors(misbehaviors []abcitypes.Misbehavior) error {
	_, err := n.finalizeBlockAndCommit(time.Second, n.valFlags, misbehaviors)
	return err
}

// finalizeBlockAndCommit is a private helper function that runs the FinalizeBlock logic
// with the provided txBytes, updates the context and
// commits the changes to have a block time after the given duration.
func (n *IntegrationNetwork) finalizeBlockAndCommit(duration time.Duration, vFlags []cmtproto.BlockIDFlag, misbehaviors []abcitypes.Misbehavior, txBytes ...[]byte) (*abcitypes.ResponseFinalizeBlock, error) {
	header := n.ctx.BlockHeader()
	// Update block header and BeginBlock
	header.Height++
	header.AppHash = n.app.LastCommitID().Hash
	// Calculate new block time after duration
	newBlockTime := header.Time.Add(duration)
	header.Time = newBlockTime

	var validatorFlags []cmtproto.BlockIDFlag
	if len(vFlags) > 0 {
		validatorFlags = vFlags
	} else {
		validatorFlags = n.valFlags
	}

	// FinalizeBlock to run endBlock, deliverTx & beginBlock logic
	req := exrpcommon.BuildFinalizeBlockReq(header, n.valSet.Validators, validatorFlags, misbehaviors, txBytes...)

	res, err := n.app.FinalizeBlock(req)
	if err != nil {
		return nil, err
	}

	newCtx := n.app.BaseApp.NewContextLegacy(false, header)

	// Update context header
	newCtx = newCtx.WithMinGasPrices(n.ctx.MinGasPrices())
	newCtx = newCtx.WithKVGasConfig(n.ctx.KVGasConfig())
	newCtx = newCtx.WithTransientKVGasConfig(n.ctx.TransientKVGasConfig())
	newCtx = newCtx.WithConsensusParams(n.ctx.ConsensusParams())
	// This might have to be changed with time if we want to test gas limits
	newCtx = newCtx.WithBlockGasMeter(storetypes.NewInfiniteGasMeter())
	newCtx = newCtx.WithVoteInfos(req.DecidedLastCommit.GetVotes())
	n.ctx = newCtx

	// commit changes
	_, err = n.app.Commit()

	return res, err
}
