package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const (
	numWorkers  = 32
	logsToFetch = 10
)

type FinalizeBlockEvent struct {
	Type string `json:"type"`
	Attributes []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
		Index bool   `json:"index"`
	} `json:"attributes"`
}

// Custom struct to handle string fields in the response
type BlockResult struct {
	Height string `json:"height"`
	TxsResults []interface{} `json:"txs_results"`
	FinalizeBlockEvents []FinalizeBlockEvent `json:"finalize_block_events"`
	ValidatorUpdates interface{} `json:"validator_updates"`
	ConsensusParamUpdates struct {
		Block struct {
			MaxBytes string `json:"max_bytes"`
			MaxGas   string `json:"max_gas"`
		} `json:"block"`
		Evidence struct {
			MaxAgeNumBlocks string `json:"max_age_num_blocks"`
			MaxAgeDuration  string `json:"max_age_duration"`
			MaxBytes        string `json:"max_bytes"`
		} `json:"evidence"`
		Validator struct {
			PubKeyTypes []string `json:"pub_key_types"`
		} `json:"validator"`
	} `json:"consensus_param_updates"`
	AppHash string `json:"app_hash"`
}

func fetchBlockResults(blockHeight uint64, wg *sync.WaitGroup, workerID int) {
	defer wg.Done()

	var prevResult BlockResult
	firstIteration := true

	for i := 0; i < logsToFetch; i++ {
		// Make HTTP request to get block results
		resp, err := http.Get(fmt.Sprintf("http://cosmos.xrplevm.org:26657/block_results?height=%d", blockHeight))
		if err != nil {
			log.Printf("Worker %d: Error fetching block results: %v", workerID, err)
			continue
		}
		defer resp.Body.Close()

		// Parse response into BlockResult
		var result struct {
			Result BlockResult `json:"result"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Printf("Worker %d: Error decoding response: %v", workerID, err)
			continue
		}

		// Compare with previous result
		if !firstIteration {
			// Convert both results to JSON for comparison
			prevJSON, err := json.Marshal(prevResult)
			if err != nil {
				log.Printf("Worker %d: Error marshaling previous result: %v", workerID, err)
				continue
			}

			currentJSON, err := json.Marshal(result.Result)
			if err != nil {
				log.Printf("Worker %d: Error marshaling current result: %v", workerID, err)
				continue
			}

			if string(prevJSON) != string(currentJSON) {
				log.Printf("Worker %d (iteration %d): Block results changed for height %d | prev: %d | current: %d | diff: %+v", workerID, i+1, blockHeight, len(prevResult.FinalizeBlockEvents), len(result.Result.FinalizeBlockEvents), prevResult)
			}
		}

		// Store current result for next iteration
		prevResult = result.Result
		firstIteration = false

		time.Sleep(20 * time.Millisecond)
	}
}

func main() {
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
		req, err := http.NewRequest("POST", "http://65.108.250.207:8545", reqBody)
		// req, err := http.NewRequest("POST", "http://localhost:8545", reqBody)
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

			// Launch goroutines to fetch block results
			for i := 0; i < numWorkers; i++ {
				wg.Add(1)
				go fetchBlockResults(currentHeight, &wg, i+1)
			}

			// Update previous block height
			prevBlockHeight = currentHeight
		}
	}
}
