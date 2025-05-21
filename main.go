package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	numWorkers  = 1
	logsToFetch = 10
)

func fetchBlockLogs(client *ethclient.Client, blockHash common.Hash, wg *sync.WaitGroup, workerID int) {
	defer wg.Done()
	address, _ := hex.DecodeString("2a5E2421D36Df7Df03868e6E51E5108871BA6E9a")
	for i := 0; i < logsToFetch; i++ { // Default to 5 iterations
		logs, err := client.FilterLogs(context.Background(), ethereum.FilterQuery{
			BlockHash: &blockHash,
			Addresses: []common.Address{common.Address(address)},
		})
		time.Sleep(20 * time.Millisecond)
		if err != nil {
			log.Printf("Error fetching logs for block %s: %v", blockHash.Hex(), err)
			continue
		}

		if len(logs) == 0 {
			log.Printf("Worker %d (iteration %d): Found %d logs in block %s", workerID, i+1, len(logs), blockHash.Hex())
			// os.Exit(1)
		}
	}
}

func main() {
	// Connect to the local EVM RPC client
	// client, err := ethclient.Dial("https://rpc.devnet.xrplevm.org")
	client, err := ethclient.Dial("http://78.47.240.16:8545")
	if err != nil {
		log.Fatalf("Failed to connect to the EVM RPC client: %v", err)
	}

	// Track the previous block height
	var prevBlockHeight uint64

	// Create a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Create HTTP client for JSON-RPC calls
	httpClient := &http.Client{}

	// Continuously fetch new blocks
	for {
		// Prepare JSON-RPC request
		reqBody := strings.NewReader(`{
			"jsonrpc": "2.0",
			"method": "eth_getBlockByNumber",
			"params": ["latest", false],
			"id": 1
		}`)

		// Make JSON-RPC request
		// req, err := http.NewRequest("POST", "http://78.47.240.16:8545", reqBody)
		req, err := http.NewRequest("POST", "http://65.108.250.207:8545", reqBody)
		if err != nil {
			log.Printf("Failed to create request: %v", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Printf("Failed to get latest block: %v", err)
			continue
		}
		defer resp.Body.Close()

		// Parse response
		var result struct {
			Result struct {
				Number string `json:"number"`
				Hash   string `json:"hash"`
			} `json:"result"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Printf("Failed to decode response: %v", err)
			continue
		}

		// Convert hex number to uint64
		currentHeight, err := strconv.ParseUint(strings.TrimPrefix(result.Result.Number, "0x"), 16, 64)
		if err != nil {
			log.Printf("Failed to parse block number: %v", err)
			continue
		}

		// Only process if block height has changed
		if currentHeight != prevBlockHeight {
			blockHash := common.HexToHash(result.Result.Hash)
			fmt.Printf("Block %d: %s\n", currentHeight, blockHash.Hex())

			// Launch goroutines to fetch block logs
			for i := 0; i < numWorkers; i++ {
				wg.Add(1)
				go fetchBlockLogs(client, blockHash, &wg, i+1)
			}

			// Update previous block height
			prevBlockHeight = currentHeight
		}
	}
}
