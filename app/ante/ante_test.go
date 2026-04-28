package ante

import (
	"errors"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	baseevmante "github.com/cosmos/evm/ante"
	antetypes "github.com/cosmos/evm/ante/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"
)

// Sentinel errors returned by the stub ante handlers to signal which branch
// the router selected, so tests can assert routing without a real ante chain.
var (
	errRoutedToCosmosAnte = errors.New("routed to cosmos ante")
	errRoutedToEVMAnte    = errors.New("routed to evm ante")
)

type routingTx struct {
	msgs             []sdk.Msg
	extensionOptions []*codectypes.Any
}

var _ sdk.Tx = routingTx{}
var _ authante.HasExtensionOptionsTx = routingTx{}

func (tx routingTx) GetMsgs() []sdk.Msg {
	return tx.msgs
}

func (tx routingTx) GetMsgsV2() ([]protov2.Message, error) {
	return nil, nil
}

func (tx routingTx) GetExtensionOptions() []*codectypes.Any {
	return tx.extensionOptions
}

func (tx routingTx) GetNonCriticalExtensionOptions() []*codectypes.Any {
	return nil
}

func stubAnteHandlerFactory(routeErr error) anteHandlerFactory {
	return func(sdk.Context, baseevmante.HandlerOptions) sdk.AnteHandler {
		return func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
			return ctx, routeErr
		}
	}
}

func TestNewAnteHandlerRouting(t *testing.T) {
	ethExtension, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	require.NoError(t, err)
	dynamicFeeExtension, err := codectypes.NewAnyWithValue(&antetypes.ExtensionOptionDynamicFeeTx{})
	require.NoError(t, err)

	cases := []struct {
		name            string
		tx              sdk.Tx
		wantErr         error
		wantErrContains string
	}{
		{
			name: "ethereum extension routes to evm ante",
			tx: routingTx{
				msgs:             []sdk.Msg{&evmtypes.MsgEthereumTx{}},
				extensionOptions: []*codectypes.Any{ethExtension},
			},
			wantErr: errRoutedToEVMAnte,
		},
		{
			name:    "dynamic fee extension routes to cosmos ante",
			tx:      routingTx{extensionOptions: []*codectypes.Any{dynamicFeeExtension}},
			wantErr: errRoutedToCosmosAnte,
		},
		{
			name:    "no extension routes to cosmos ante",
			tx:      routingTx{},
			wantErr: errRoutedToCosmosAnte,
		},
		{
			name: "unknown extension is rejected",
			tx: routingTx{
				extensionOptions: []*codectypes.Any{{
					TypeUrl: "/cosmos.evm.test.v1.UnknownExtensionOption",
				}},
			},
			wantErr:         errortypes.ErrUnknownExtensionOptions,
			wantErrContains: "rejecting tx with unsupported extension option: /cosmos.evm.test.v1.UnknownExtensionOption",
		},
		{
			name:            "nil tx is rejected",
			tx:              nil,
			wantErr:         errortypes.ErrUnknownRequest,
			wantErrContains: "invalid transaction type: <nil>",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := newAnteHandler(
				baseevmante.HandlerOptions{},
				stubAnteHandlerFactory(errRoutedToCosmosAnte),
				stubAnteHandlerFactory(errRoutedToEVMAnte),
			)

			_, err := handler(sdk.Context{}, tc.tx, false)

			require.ErrorIs(t, err, tc.wantErr)
			if tc.wantErrContains != "" {
				require.ErrorContains(t, err, tc.wantErrContains)
			}
		})
	}
}
