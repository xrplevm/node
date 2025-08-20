package simulation

import (
	"fmt"
	"math/rand"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v9/x/poa/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestProposalMsgs(t *testing.T) {
	s := rand.NewSource(1)
	//nolint:gosec
	r := rand.New(s)

	ctx := sdk.NewContext(nil, tmproto.Header{}, true, nil)
	accounts := simtypes.RandomAccounts(r, 3)

	// execute ProposalMsgs function
	weightedProposalMsgs := ProposalMsgs()
	require.Equal(t, 2, len(weightedProposalMsgs))

	w0 := weightedProposalMsgs[0]

	// tests w0 interface:
	require.Equal(t, OpWeightMsgAddValidator, w0.AppParamsKey())
	require.Equal(t, DefaultWeightMsgAddValidator, w0.DefaultWeight())

	msg := w0.MsgSimulatorFn()(r, ctx, accounts)
	msgAddValidator, ok := msg.(*types.MsgAddValidator)
	require.True(t, ok)

	fmt.Println(msgAddValidator)
	require.Equal(t, sdk.AccAddress(address.Module("gov")).String(), msgAddValidator.Authority)

	w1 := weightedProposalMsgs[1]

	// tests w0 interface:
	require.Equal(t, OpWeightMsgRemoveValidator, w1.AppParamsKey())
	require.Equal(t, DefaultWeightMsgRemoveValidator, w1.DefaultWeight())

	msg = w1.MsgSimulatorFn()(r, ctx, accounts)
	msgRemoveValidator, ok := msg.(*types.MsgRemoveValidator)
	require.True(t, ok)

	fmt.Println(msgRemoveValidator)
	require.Equal(t, sdk.AccAddress(address.Module("gov")).String(), msgRemoveValidator.Authority)
}
