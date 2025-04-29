package blockchain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestClient_GetBlockNumber tests the GetBlockNumber method.
func TestClient_GetBlockNumber(t *testing.T) {
	// Set up a dummy server to simulate blockchain JSON-RPC responses.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only respond to POST JSON requests
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		method, ok := req["method"].(string)
		if !ok {
			t.Errorf("no method in request")
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		// Respond based on the method
		switch method {
		case "eth_blockNumber":
			// Return a fixed block number, e.g., 0x5 (5 in decimal)
			resp := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"result":  "0x5",
			}
			_ = json.NewEncoder(w).Encode(resp)
		case "eth_getBlockByNumber":
			// Extract params to determine which block number was requested
			params, ok := req["params"].([]interface{})
			if !ok || len(params) < 2 {
				t.Errorf("params not provided correctly")
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			numHex, ok := params[0].(string)
			if !ok {
				t.Errorf("block number param not a string")
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			// If a specific block number is requested, return dummy block data
			if numHex == "0x5" {
				// Provide dummy block data for block number 5
				block := map[string]interface{}{
					"number":     "0x5",
					"hash":       "0xabcde",    // dummy hash
					"parentHash": "0x00000",    // dummy parent hash
					"timestamp":  "0x611ad38c", // some dummy timestamp
					"transactions": []map[string]interface{}{ // one dummy transaction
						{
							"hash":  "0xtxhash1",
							"from":  "0xfromaddr",
							"to":    "0xtoaddr",
							"value": "0xde0b6b3a7640000",
						},
					},
				}
				resp := map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      1,
					"result":  block,
				}
				_ = json.NewEncoder(w).Encode(resp)
			} else {
				// If block not found or some other block, return an error JSON-RPC response
				resp := map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      1,
					"error": map[string]interface{}{
						"code":    -32602,
						"message": fmt.Sprintf("Block %s not found", numHex),
					},
				}
				_ = json.NewEncoder(w).Encode(resp)
			}
		default:
			// Method not supported
			resp := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"error": map[string]interface{}{
					"code":    -32601,
					"message": "Method not found",
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		}
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	// Create client pointing to our test server URL
	client := NewClient(server.URL)

	// Test GetBlockNumber
	number, err := client.GetBlockNumber()
	if err != nil {
		t.Fatalf("GetBlockNumber returned error: %v", err)
	}
	if number != 5 {
		t.Errorf("expected block number 5, got %d", number)
	}

	// Test GetBlockByNumber for a valid block (5)
	block, err := client.GetBlockByNumber(5)
	if err != nil {
		t.Fatalf("GetBlockByNumber(5) returned error: %v", err)
	}
	// Verify block fields
	if uint64(block.Number) != 5 {
		t.Errorf("expected block.Number 5, got %d", uint64(block.Number))
	}
	if block.Hash != "0xabcde" {
		t.Errorf("unexpected block.Hash, got %s", block.Hash)
	}
	if len(block.Transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(block.Transactions))
	}
	if len(block.Transactions) > 0 {
		tx := block.Transactions[0]
		if tx.Hash != "0xtxhash1" || tx.From != "0xfromaddr" || tx.To != "0xtoaddr" {
			t.Errorf("transaction fields not matching expected values")
		}
	}

	// Test GetBlockByNumber for a block that triggers an error (simulate not found)
	_, err = client.GetBlockByNumber(9999)
	if err == nil {
		t.Fatal("expected error for non-existent block, got nil")
	}
	// The error message should contain "not found"
	if err != nil && !contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

// contains is a helper to check substring in a string (to avoid import strings in test again).
func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && (func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	})())
}
