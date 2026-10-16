package app

import (
	"fmt"
	"sort"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/types/module"
)

// unroutedMsgMethods lists the Msg service methods that the app does not
// route. The modules register them, but this chain does not use them. Each
// group below gives its own reason. Their types stay registered in the
// interface registry, so historical txs and proposals still decode. A tx or a
// gov proposal that carries one of them fails.
var unroutedMsgMethods = []string{
	// x/poa owns the validator set. Delegations outside x/poa would change
	// voting power without a gov decision.
	"/cosmos.staking.v1beta1.Msg/Delegate",
	"/cosmos.staking.v1beta1.Msg/Undelegate",
	"/cosmos.staking.v1beta1.Msg/BeginRedelegate",
	"/cosmos.staking.v1beta1.Msg/CancelUnbondingDelegation",

	// Users do not add funds to the community pool. Gov can still spend the
	// pool with MsgCommunityPoolSpend.
	"/cosmos.distribution.v1beta1.Msg/FundCommunityPool",

	// External deposits into the reward pool of one validator are not part of
	// this chain's reward flow.
	"/cosmos.distribution.v1beta1.Msg/DepositValidatorRewardsPool",

	// No evidence router is set. CometBFT misbehavior is handled in the
	// x/evidence BeginBlocker.
	"/cosmos.evidence.v1beta1.Msg/SubmitEvidence",

	// gov v1beta1 and legacy content. Clients send gov v1.
	"/cosmos.gov.v1beta1.Msg/SubmitProposal",
	"/cosmos.gov.v1beta1.Msg/Vote",
	"/cosmos.gov.v1beta1.Msg/VoteWeighted",
	"/cosmos.gov.v1beta1.Msg/Deposit",
	"/cosmos.gov.v1.Msg/ExecLegacyContent",

	// A weighted vote splits one voter's power between options, for a voter
	// that represents several stakeholders. Each x/poa validator votes only
	// with its own self-delegation, so a single-option MsgVote is enough.
	"/cosmos.gov.v1.Msg/VoteWeighted",

	// ICS20 rejects user-initiated channel closing.
	"/ibc.core.channel.v1.Msg/ChannelCloseInit",

	// IBC v2 has no app router (SetRouterV2 is never called).
	"/ibc.core.channel.v2.Msg/SendPacket",
	"/ibc.core.channel.v2.Msg/RecvPacket",
	"/ibc.core.channel.v2.Msg/Acknowledgement",
	"/ibc.core.channel.v2.Msg/Timeout",

	// A client counterparty only prepares IBC v2 packet flow, which the app
	// does not support.
	"/ibc.core.client.v2.Msg/RegisterCounterparty",
}

// msgFilterConfigurator drops the listed Msg service methods before the
// module services reach the MsgServiceRouter. Query services and migrations
// pass through unchanged.
type msgFilterConfigurator struct {
	module.Configurator
	drop    map[string]struct{}
	dropped map[string]struct{}
}

func newMsgFilterConfigurator(cfg module.Configurator, methods []string) *msgFilterConfigurator {
	drop := make(map[string]struct{}, len(methods))
	for _, m := range methods {
		drop[m] = struct{}{}
	}
	return &msgFilterConfigurator{Configurator: cfg, drop: drop, dropped: map[string]struct{}{}}
}

// MsgServer is used by modules that implement module.HasServices.
func (c *msgFilterConfigurator) MsgServer() gogogrpc.Server {
	return msgFilterServer{inner: c.Configurator.MsgServer(), cfg: c}
}

// RegisterService is used by modules that implement appmodule.HasServices.
func (c *msgFilterConfigurator) RegisterService(sd *grpc.ServiceDesc, ss any) {
	if filtered := c.filter(sd); filtered != nil {
		c.Configurator.RegisterService(filtered, ss)
	}
}

// filter returns sd without the dropped methods, or nil if no method is left.
func (c *msgFilterConfigurator) filter(sd *grpc.ServiceDesc) *grpc.ServiceDesc {
	filtered := *sd
	filtered.Methods = make([]grpc.MethodDesc, 0, len(sd.Methods))
	for _, m := range sd.Methods {
		name := fmt.Sprintf("/%s/%s", sd.ServiceName, m.MethodName)
		if _, ok := c.drop[name]; ok {
			c.dropped[name] = struct{}{}
			continue
		}
		filtered.Methods = append(filtered.Methods, m)
	}
	if len(filtered.Methods) == 0 && len(filtered.Streams) == 0 {
		return nil
	}
	return &filtered
}

// unmatched returns the listed methods that no module registered. A non-empty
// result means a dependency renamed or removed a method.
func (c *msgFilterConfigurator) unmatched() []string {
	var missing []string
	for m := range c.drop {
		if _, ok := c.dropped[m]; !ok {
			missing = append(missing, m)
		}
	}
	sort.Strings(missing)
	return missing
}

type msgFilterServer struct {
	inner gogogrpc.Server
	cfg   *msgFilterConfigurator
}

func (s msgFilterServer) RegisterService(sd *grpc.ServiceDesc, ss any) {
	if filtered := s.cfg.filter(sd); filtered != nil {
		s.inner.RegisterService(filtered, ss)
	}
}
