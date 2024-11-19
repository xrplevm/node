package network

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/evmos/evmos/v19/server/config"
	"github.com/stretchr/testify/require"
)

// package-wide network lock to only allow one test network at a time
var lock = new(sync.Mutex)

type NetworkManager struct {

	network *Network
}

func NewManager(t *testing.T, numValidators int, numBondedValidators int, blockTime time.Duration, unbondingBlocks int64, nodeDir string) *NetworkManager {
	cfg := DefaultConfig(numValidators, numBondedValidators, blockTime, unbondingBlocks)

	network, err := NewNetwork(t, nodeDir, cfg)
	require.NoError(t, err)

	return &NetworkManager{network}
}

func (m *NetworkManager) JSONRPCClient() *ethclient.Client {
	return m.network.Validators[0].JSONRPCClient
}

func (m *NetworkManager) SetJSONRPCClient(client *ethclient.Client) {
	m.network.Validators[0].JSONRPCClient = client
}

func (m *NetworkManager) Cleanup() {
	m.network.Cleanup()
}

// LatestHeight returns the latest height of the network or an error if the
// query fails or no validators exist.
func (m *NetworkManager) LatestHeight() (int64, error) {
	if len(m.network.Validators) == 0 {
		return 0, errors.New("no validators available")
	}

	status, err := m.network.Validators[0].RPCClient.Status(context.Background())
	if err != nil {
		return 0, err
	}

	return status.SyncInfo.LatestBlockHeight, nil
}

// WaitForHeight performs a blocking check where it waits for a block to be
// committed after a given block. If that height is not reached within a timeout,
// an error is returned. Regardless, the latest height queried is returned.
func (m *NetworkManager) WaitForHeight(h int64) (int64, error) {
	return m.WaitForHeightWithTimeout(h, 10*time.Second)
}

// WaitForHeightWithTimeout is the same as WaitForHeight except the caller can
// provide a custom timeout.
func (m *NetworkManager) WaitForHeightWithTimeout(h int64, t time.Duration) (int64, error) {
	ticker := time.NewTicker(time.Second)
	timeout := time.After(t)

	if len(m.network.Validators) == 0 {
		return 0, errors.New("no validators available")
	}

	var latestHeight int64
	val := m.network.Validators[0]

	for {
		select {
		case <-timeout:
			ticker.Stop()
			return latestHeight, errors.New("timeout exceeded waiting for block")
		case <-ticker.C:
			status, err := val.RPCClient.Status(context.Background())
			if err == nil && status != nil {
				latestHeight = status.SyncInfo.LatestBlockHeight
				if latestHeight >= h {
					return latestHeight, nil
				}
			}
		}
	}
}

// WaitForNextBlock waits for the next block to be committed, returning an error
// upon failure.
func (m *NetworkManager) WaitForNextBlock() error {
	lastBlock, err := m.LatestHeight()
	if err != nil {
		return err
	}

	_, err = m.WaitForHeight(lastBlock + 1)
	if err != nil {
		return err
	}

	return err
}

func (m *NetworkManager) MustWaitForNextBlock() {
	if err := m.WaitForNextBlock(); err != nil {
		panic(err)
	}
}

func (m *NetworkManager) AppConfig() *config.Config {
	return m.network.Validators[0].AppConfig
}
