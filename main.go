package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"BlockchainClient/blockchain"
)

func main() {
	// Determine RPC endpoint from environment or use default (Polygon mainnet)
	rpcURL := os.Getenv("RPC_URL")
	client := blockchain.NewClient(rpcURL)

	// HTTP handler for getting latest block number
	http.HandleFunc("/blockNumber", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		number, err := client.GetBlockNumber()
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			// If the error occurs, respond with 502 (bad gateway) since it's likely an upstream issue
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]uint64{"blockNumber": number})
	})

	// HTTP handler for getting block by number (e.g., /block/12345)
	http.HandleFunc("/block/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		// Extract the block number from URL path
		// Path is expected as /block/{number}
		path := r.URL.Path
		if len(path) <= len("/block/") {
			http.Error(w, "block number not specified", http.StatusBadRequest)
			return
		}
		numStr := path[len("/block/"):]
		var num uint64
		var err error
		if len(numStr) > 1 && numStr[:2] == "0x" {
			// parse hex number (exclude "0x")
			num, err = strconv.ParseUint(numStr[2:], 16, 64)
		} else {
			// parse decimal
			num, err = strconv.ParseUint(numStr, 10, 64)
		}
		if err != nil {
			http.Error(w, "invalid block number", http.StatusBadRequest)
			return
		}
		block, err := client.GetBlockByNumber(num)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			// Determine status code based on error message
			status := http.StatusBadGateway
			if strings.Contains(err.Error(), "not found") {
				status = http.StatusNotFound
			}
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		// Return the block as JSON
		_ = json.NewEncoder(w).Encode(block)
	})

	// Determine port for the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
