package network

import (
	"net/http"

	"github.com/cometbft/cometbft/node"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"

	tmclient "github.com/cometbft/cometbft/rpc/client"
	"github.com/cosmos/cosmos-sdk/server/api"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/server/config"
)

type Validator struct {
	AppConfig     *config.Config
	ClientCtx     client.Context
	Ctx           *server.Context
	Dir           string
	NodeID        string
	PubKey        cryptotypes.PubKey
	Moniker       string
	APIAddress    string
	RPCAddress    string
	P2PAddress    string
	Address       sdk.AccAddress
	ValAddress    sdk.ValAddress
	RPCClient     tmclient.Client
	JSONRPCClient *ethclient.Client

	TmNode      *node.Node
	api         *api.Server
	grpc        *grpc.Server
	grpcWeb     *http.Server
	jsonrpc     *http.Server
	jsonrpcDone chan struct{}
}