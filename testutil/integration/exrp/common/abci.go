package exrpcommon

import (
	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
)

// buildFinalizeBlockReq is a helper function to build
// properly the FinalizeBlock request
func BuildFinalizeBlockReq(header cmtproto.Header, validators []*cmttypes.Validator, txs ...[]byte) *abcitypes.RequestFinalizeBlock {
	// add validator's commit info to allocate corresponding tokens to validators
	ci := GetCommitInfo(validators)
	return &abcitypes.RequestFinalizeBlock{
		Height:             header.Height,
		DecidedLastCommit:  ci,
		Hash:               header.AppHash,
		NextValidatorsHash: header.ValidatorsHash,
		ProposerAddress:    header.ProposerAddress,
		Time:               header.Time,
		Txs:                txs,
	}
}

func GetCommitInfo(validators []*cmttypes.Validator) abcitypes.CommitInfo {
	voteInfos := make([]abcitypes.VoteInfo, len(validators))
	for i, val := range validators {
		voteInfos[i] = abcitypes.VoteInfo{
			Validator: abcitypes.Validator{
				Address: val.Address,
				Power:   val.VotingPower,
			},
			BlockIdFlag: cmtproto.BlockIDFlagCommit,
		}
	}
	return abcitypes.CommitInfo{Votes: voteInfos}
}
