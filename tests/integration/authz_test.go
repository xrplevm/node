package integration

import (
	"fmt"
	"math/rand"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/evm/testutil/integration/base/factory"
	"github.com/cosmos/evm/testutil/keyring"
	"github.com/stretchr/testify/require"
	poatypes "github.com/xrplevm/node/v12/x/poa/types"
)

// x/authz was removed in v12, but its msgs still decode so that historical txs
// stay readable. A tx that fails to decode fails with ErrTxDecode, so the
// expected errors below also prove that the msgs decoded.
func (s *TestSuite) TestAuthz_RemovedMsgs() {
	// Generate a random account
	randomAccs := simtypes.RandomAccounts(rand.New(rand.NewSource(time.Now().UnixNano())), 1) //nolint:gosec
	randomAcc := randomAccs[0]
	randomValAddr := sdktypes.ValAddress(randomAcc.Address.Bytes())

	authority := sdktypes.AccAddress(address.Module("gov"))
	granter := s.keyring.GetKey(0)
	grantee := s.keyring.GetKey(1)

	addValidator, err := poatypes.NewMsgAddValidator(
		authority.String(),
		randomAcc.Address.String(),
		randomAcc.ConsKey.PubKey(),
		stakingtypes.Description{
			Moniker: "test",
		},
	)
	require.NoError(s.T(), err)
	addValidatorURL := sdktypes.MsgTypeURL(addValidator)
	expiration := s.Network().GetContext().BlockTime().Add(24 * time.Hour)

	grant, err := authz.NewMsgGrant(granter.AccAddr, grantee.AccAddr, authz.NewGenericAuthorization(addValidatorURL), &expiration)
	require.NoError(s.T(), err)
	exec := authz.NewMsgExec(grantee.AccAddr, []sdktypes.Msg{addValidator})
	revoke := authz.NewMsgRevoke(granter.AccAddr, grantee.AccAddr, addValidatorURL)

	// Gov must sign every proposal msg, so gov is the granter or the grantee.
	govGrant, err := authz.NewMsgGrant(authority, grantee.AccAddr, authz.NewGenericAuthorization(addValidatorURL), &expiration)
	require.NoError(s.T(), err)
	govExec := authz.NewMsgExec(authority, []sdktypes.Msg{addValidator})
	govRevoke := authz.NewMsgRevoke(authority, grantee.AccAddr, addValidatorURL)
	proposal := func(msg sdktypes.Msg) sdktypes.Msg {
		p, err := govv1.NewMsgSubmitProposal(
			[]sdktypes.Msg{msg},
			sdktypes.NewCoins(sdktypes.NewCoin(s.Network().GetBaseDenom(), s.Network().GetMinDepositAmt())),
			granter.AccAddr.String(),
			"",
			"test",
			"test",
			false,
		)
		require.NoError(s.T(), err)
		return p
	}

	requireValidatorNotAdded := func() {
		_, err := s.Network().GetStakingClient().Validator(
			s.Network().GetContext(),
			&stakingtypes.QueryValidatorRequest{
				ValidatorAddr: randomValAddr.String(),
			},
		)
		require.Contains(s.T(), err.Error(), fmt.Sprintf("validator %s not found", randomValAddr.String()))
	}

	tt := []struct {
		name          string
		signer        keyring.Key
		msg           sdktypes.Msg
		expectedError error
		afterRun      func()
	}{
		{
			name:          "authz - MsgGrant",
			signer:        granter,
			msg:           grant,
			expectedError: sdkerrors.ErrUnknownRequest,
		},
		{
			name:          "authz - MsgExec",
			signer:        grantee,
			msg:           &exec,
			expectedError: sdkerrors.ErrUnknownRequest,
			afterRun:      requireValidatorNotAdded,
		},
		{
			name:          "authz - MsgRevoke",
			signer:        granter,
			msg:           &revoke,
			expectedError: sdkerrors.ErrUnknownRequest,
		},
		{
			name:          "gov proposal - MsgGrant",
			signer:        granter,
			msg:           proposal(govGrant),
			expectedError: govtypes.ErrUnroutableProposalMsg,
		},
		{
			name:          "gov proposal - MsgExec",
			signer:        granter,
			msg:           proposal(&govExec),
			expectedError: govtypes.ErrUnroutableProposalMsg,
			afterRun:      requireValidatorNotAdded,
		},
		{
			name:          "gov proposal - MsgRevoke",
			signer:        granter,
			msg:           proposal(&govRevoke),
			expectedError: govtypes.ErrUnroutableProposalMsg,
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			gas := uint64(1_000_000)
			res, err := s.factory.CommitCosmosTx(tc.signer.Priv, factory.CosmosTxArgs{
				Msgs: []sdktypes.Msg{tc.msg},
				Gas:  &gas,
			})
			require.NoError(s.T(), err)                                // The block did not fail.
			require.NotEqual(s.T(), uint32(0), res.Code)               // The tx failed.
			require.Contains(s.T(), res.Log, tc.expectedError.Error()) // The tx failed with the expected error.

			if tc.afterRun != nil {
				tc.afterRun()
			}
		})
	}
}
