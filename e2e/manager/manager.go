package manager

import (
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/evmos/evmos/v19/server/config"
)

const (
	NetworkManager = "network"
)

type Manager interface {
	// Add here more module operations
	// modules.PoaOps
	

	// Block related methods
	LatestHeight() (int64, error)
	WaitForHeight(h int64) (int64, error)
	WaitForHeightWithTimeout(h int64, t time.Duration) (int64, error)
	WaitForNextBlock() error
	MustWaitForNextBlock()

	// Manager related methods
	Cleanup()
	JSONRPCClient() *ethclient.Client
	SetJSONRPCClient(client *ethclient.Client)
	AppConfig() *config.Config
}


