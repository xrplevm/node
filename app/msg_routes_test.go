package app_test

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	gogoproto "github.com/cosmos/gogoproto/proto"
	clientv2types "github.com/cosmos/ibc-go/v10/modules/core/02-client/v2/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/xrplevm/node/v12/app"
)

// Every method in unroutedMsgMethods must have no handler in the built app,
// and its msg must still decode so that historical txs and proposals stay
// readable. The test reads the router itself, so it also catches a route that
// a module registers outside the filter.
func TestNew_RemovedMsgsHaveNoHandler(t *testing.T) {
	exrpApp := sharedTestApp()

	for _, method := range app.UnroutedMsgMethods {
		service, name, ok := strings.Cut(strings.TrimPrefix(method, "/"), "/")
		require.True(t, ok, method)
		desc, err := gogoproto.HybridResolver.FindDescriptorByName(protoreflect.FullName(service))
		require.NoError(t, err, method)
		sd, ok := desc.(protoreflect.ServiceDescriptor)
		require.True(t, ok, method)
		md := sd.Methods().ByName(protoreflect.Name(name))
		require.NotNil(t, md, method)
		typeURL := "/" + string(md.Input().FullName())

		_, err = exrpApp.InterfaceRegistry().Resolve(typeURL)
		require.NoError(t, err, typeURL)
		require.Nil(t, exrpApp.MsgServiceRouter().HandlerByTypeURL(typeURL), typeURL)
	}
}

// Each msg below shares a Msg service with a removed msg, so it must keep its
// handler. This catches a filter that removes too much: the app still starts,
// but a feature breaks.
func TestNew_KeptMsgsNextToRemovedOnesHaveHandler(t *testing.T) {
	exrpApp := sharedTestApp()

	tt := []struct {
		name string
		msg  sdk.Msg
	}{
		{name: "x/poa creates validators through the router, and gentxs need it", msg: &stakingtypes.MsgCreateValidator{}},
		{name: "validators edit their description", msg: &stakingtypes.MsgEditValidator{}},
		{name: "gov sets MaxValidators, which x/poa reads", msg: &stakingtypes.MsgUpdateParams{}},
		{name: "validators collect fee rewards", msg: &distrtypes.MsgWithdrawDelegatorReward{}},
		{name: "validators collect commission", msg: &distrtypes.MsgWithdrawValidatorCommission{}},
		{name: "validators choose the reward destination", msg: &distrtypes.MsgSetWithdrawAddress{}},
		{name: "gov spends the community pool", msg: &distrtypes.MsgCommunityPoolSpend{}},
		{name: "gov v1 is the only proposal path", msg: &govv1.MsgSubmitProposal{}},
		{name: "validators vote", msg: &govv1.MsgVote{}},
		{name: "proposals need a deposit", msg: &govv1.MsgDeposit{}},
		{name: "IBC v1 channels close after the counterparty closes them", msg: &channeltypes.MsgChannelCloseConfirm{}},
		{name: "IBC v1 packets are relayed", msg: &channeltypes.MsgRecvPacket{}},
		{name: "sets the relayer allowlist that v1 MsgUpdateClient enforces", msg: &clientv2types.MsgUpdateClientConfig{}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require.NotNil(t, exrpApp.MsgServiceRouter().Handler(tc.msg), sdk.MsgTypeURL(tc.msg))
		})
	}
}

// The filter only touches Msg services. The query services of the filtered
// modules must stay registered.
func TestNew_FilteredModulesKeepQueries(t *testing.T) {
	queryRouter := sharedTestApp().GRPCQueryRouter()

	for _, method := range []string{
		"/cosmos.evidence.v1beta1.Query/AllEvidence",
		"/cosmos.gov.v1beta1.Query/Proposals",
		"/cosmos.distribution.v1beta1.Query/CommunityPool",
		"/ibc.core.channel.v2.Query/PacketCommitment",
	} {
		require.NotNil(t, queryRouter.Route(method), method)
	}
}
