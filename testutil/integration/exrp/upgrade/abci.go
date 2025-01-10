// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package exrpupgrade

import (
	"time"

	storetypes "cosmossdk.io/store/types"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	exrpcommon "github.com/xrplevm/node/v5/testutil/integration/exrp/common"
)

// NextBlock is a private helper function that runs the EndBlocker logic, commits the changes,
// updates the header and runs the BeginBlocker
func (n *UpgradeIntegrationNetwork) NextBlock() error {
	return n.NextBlockAfter(time.Second)
}

// NextBlockAfter is a private helper function that runs the FinalizeBlock logic, updates the context and
// commits the changes to have a block time after the given duration.
func (n *UpgradeIntegrationNetwork) NextBlockAfter(duration time.Duration) error {
	_, err := n.finalizeBlockAndCommit(duration)
	return err
}

// NextBlockWithTxs is a helper function that runs the FinalizeBlock logic
// with the provided tx bytes, updates the context and
// commits the changes to have a block time after the given duration.
func (n *UpgradeIntegrationNetwork) NextBlockWithTxs(txBytes ...[]byte) (*abcitypes.ResponseFinalizeBlock, error) {
	return n.finalizeBlockAndCommit(time.Second, txBytes...)
}

// finalizeBlockAndCommit is a private helper function that runs the FinalizeBlock logic
// with the provided txBytes, updates the context and
// commits the changes to have a block time after the given duration.
func (n *UpgradeIntegrationNetwork) finalizeBlockAndCommit(duration time.Duration, txBytes ...[]byte) (*abcitypes.ResponseFinalizeBlock, error) {
	header := n.ctx.BlockHeader()
	// Update block header and BeginBlock
	header.Height++
	header.AppHash = n.app.LastCommitID().Hash
	// Calculate new block time after duration
	newBlockTime := header.Time.Add(duration)
	header.Time = newBlockTime

	// FinalizeBlock to run endBlock, deliverTx & beginBlock logic
	req := exrpcommon.BuildFinalizeBlockReq(header, n.valSet.Validators, nil, nil, txBytes...)

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
