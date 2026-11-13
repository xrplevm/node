package ante

import (
	errorsmod "cosmossdk.io/errors"
	evidencetypes "cosmossdk.io/x/evidence/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	clientv2types "github.com/cosmos/ibc-go/v10/modules/core/02-client/v2/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	channelv2types "github.com/cosmos/ibc-go/v10/modules/core/04-channel/v2/types"
)

// blockedUserMsgs holds the type URLs of the msgs that a user tx must not
// carry. This chain does not use them, and each group below gives its own
// reason. The modules keep their routes, so gov proposals and modules that
// call the msg router still reach them. The staking msgs that x/poa blocks
// are in x/poa/ante.
var blockedUserMsgs = msgTypeURLs(
	// Users do not add funds to the community pool. Gov can still spend the
	// pool with MsgCommunityPoolSpend.
	&distrtypes.MsgFundCommunityPool{},

	// External deposits into the reward pool of one validator are not part of
	// this chain's reward flow.
	&distrtypes.MsgDepositValidatorRewardsPool{},

	// No evidence router is set. CometBFT misbehavior is handled in the
	// x/evidence BeginBlocker.
	&evidencetypes.MsgSubmitEvidence{},

	// gov v1beta1. Clients send gov v1.
	&govv1beta1.MsgSubmitProposal{},
	&govv1beta1.MsgVote{},
	&govv1beta1.MsgVoteWeighted{},
	&govv1beta1.MsgDeposit{},

	// A weighted vote splits one voter's power between options, for a voter
	// that represents several stakeholders. Each x/poa validator votes only
	// with its own self-delegation, so a single-option MsgVote is enough.
	&govv1.MsgVoteWeighted{},

	// ICS20 rejects user-initiated channel closing.
	&channeltypes.MsgChannelCloseInit{},

	// IBC v2 has no app router (SetRouterV2 is never called).
	&channelv2types.MsgSendPacket{},
	&channelv2types.MsgRecvPacket{},
	&channelv2types.MsgAcknowledgement{},
	&channelv2types.MsgTimeout{},

	// A client counterparty only prepares IBC v2 packet flow, which the app
	// does not support.
	&clientv2types.MsgRegisterCounterparty{},
)

func msgTypeURLs(msgs ...sdk.Msg) map[string]struct{} {
	typeURLs := make(map[string]struct{}, len(msgs))
	for _, msg := range msgs {
		typeURLs[sdk.MsgTypeURL(msg)] = struct{}{}
	}
	return typeURLs
}

// blockedMsgsDecorator rejects a tx that carries a msg of blockedUserMsgs. It
// checks only the top-level msgs of the tx. A msg nested in a gov proposal
// passes, because gov is trusted to execute it.
type blockedMsgsDecorator struct{}

func (blockedMsgsDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	for _, msg := range tx.GetMsgs() {
		typeURL := sdk.MsgTypeURL(msg)
		if _, blocked := blockedUserMsgs[typeURL]; blocked {
			return ctx, errorsmod.Wrapf(errortypes.ErrUnauthorized, "msg %s is not allowed in a tx", typeURL)
		}
	}
	return next(ctx, tx, simulate)
}
