package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	numWorkers = 2
	logsToFetch = 10
)

func fetchBlockLogs(client *ethclient.Client, blockHash common.Hash, wg *sync.WaitGroup, workerID int) {
	defer wg.Done()

	for i := 0; i < logsToFetch; i++ { // Default to 5 iterations
		logs, err := client.FilterLogs(context.Background(), ethereum.FilterQuery{
			BlockHash: &blockHash,
		})
		time.Sleep(2 * time.Second)
		if err != nil {
			// log.Printf("Error fetching logs for block %s: %v", blockHash.Hex(), err)
		}

		log.Printf("Worker %d (iteration %d): Found %d logs in block %s", workerID, i+1, len(logs), blockHash.Hex())
	}
}

func main() {
	// Connect to the local EVM RPC client
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatalf("Failed to connect to the EVM RPC client: %v", err)
	}
	// Track the previous block height
	var prevBlockHeight uint64

	// Create a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Continuously fetch new blocks
	for {
		// Get the latest block
		block, err := client.BlockByNumber(context.Background(), nil)
		if err != nil {
			log.Printf("Failed to get latest block: %v", err)
			continue
		}

		// Verify if the block is valid
		if block == nil {
			log.Printf("Received nil block")
			continue
		}

		currentHeight := block.Number().Uint64()

		// Only process if block height has changed
		if currentHeight != prevBlockHeight {
			fmt.Printf("Block %d: %s\n", currentHeight, block.Hash().Hex())

			// Launch goroutines to fetch block logs
			for i := 0; i < numWorkers; i++ {
				wg.Add(1)
				go fetchBlockLogs(client, block.Hash(), &wg, i+1)
			}

			// Update previous block height
			prevBlockHeight = currentHeight
		}

		time.Sleep(1 * time.Second)
	}
}