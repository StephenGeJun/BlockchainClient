package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Blockchain defines the interface for blockchain client with basic methods.
type Blockchain interface {
	GetBlockNumber() (uint64, error)
	GetBlockByNumber(number uint64) (*Block, error)
}

// Client is a simple blockchain client implementing the Blockchain interface.
type Client struct {
	endpoint   string
	httpClient *http.Client
}

// rpcError represents a JSON-RPC error response.
type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// rpcResponse represents a generic JSON-RPC response.
type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error,omitempty"`
	ID     int             `json:"id"`
}

// Uint64Hex is a helper type for parsing hex-encoded uint64 values from JSON.
type Uint64Hex uint64

// Block represents a simplified blockchain block structure.
type Block struct {
	Number       Uint64Hex     `json:"number"`       // Block number (hex string in JSON, decoded to uint64)
	Hash         string        `json:"hash"`         // Block hash
	ParentHash   string        `json:"parentHash"`   // Parent block hash
	Timestamp    Uint64Hex     `json:"timestamp"`    // Timestamp (hex in JSON, decoded to uint64)
	Transactions []Transaction `json:"transactions"` // List of transactions in the block
}

// Transaction represents a simplified blockchain transaction structure.
type Transaction struct {
	Hash  string `json:"hash"`  // Transaction hash
	From  string `json:"from"`  // Sender address
	To    string `json:"to"`    // Recipient address
	Value string `json:"value"` // Value transferred (in Wei, hex string in JSON)
}

// NewClient creates a new blockchain client for the given JSON-RPC endpoint.
func NewClient(endpoint string) *Client {
	if endpoint == "" {
		endpoint = "https://polygon-rpc.com"
	}
	// Create an HTTP client with a timeout for safety.
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	return &Client{
		endpoint:   endpoint,
		httpClient: httpClient,
	}
}

// rpcCall is a helper to perform a JSON-RPC call to the endpoint.
// method is the RPC method name, params is a slice of parameters.
// result is a pointer to an object into which the result will be unmarshaled.
func (c *Client) rpcCall(method string, params []interface{}, result interface{}) error {
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
	}
	if params != nil {
		reqBody["params"] = params
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to encode request: %v", err)
	}
	// Send HTTP POST request
	resp, err := c.httpClient.Post(c.endpoint, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("RPC request error: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status: %d - %s", resp.StatusCode, string(body))
	}
	// Parse JSON-RPC response
	var rpcResp rpcResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return fmt.Errorf("failed to decode JSON-RPC response: %v", err)
	}
	if rpcResp.Error != nil {
		return fmt.Errorf("JSON-RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}
	// If a result container is provided, unmarshal the result into it
	if result != nil {
		if err := json.Unmarshal(rpcResp.Result, result); err != nil {
			return fmt.Errorf("failed to decode result: %v", err)
		}
	}
	return nil
}

// GetBlockNumber returns the latest block number in the blockchain.
func (c *Client) GetBlockNumber() (uint64, error) {
	var resultHex string
	// Perform RPC call;
	if err := c.rpcCall("eth_blockNumber", nil, &resultHex); err != nil {
		return 0, err
	}
	if !strings.HasPrefix(resultHex, "0x") {
		return 0, fmt.Errorf("invalid block number format: %s", resultHex)
	}
	// Parse hex string to uint64
	num, err := strconv.ParseUint(resultHex[2:], 16, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse block number: %v", err)
	}
	return num, nil
}

// GetBlockByNumber returns the full block information for the given block number.
func (c *Client) GetBlockByNumber(number uint64) (*Block, error) {
	// Convert number to hex string with "0x" prefix, as required by Ethereum JSON-RPC
	hexNum := "0x" + strconv.FormatUint(number, 16)
	params := []interface{}{hexNum, true}
	var block Block
	if err := c.rpcCall("eth_getBlockByNumber", params, &block); err != nil {
		return nil, err
	}
	// Check if the block was found
	if block.Hash == "" {
		return nil, fmt.Errorf("block %d not found", number)
	}
	return &block, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface for Uint64Hex.
// It expects a JSON string in hex format (e.g., "0x1a") and unmarshals it into a uint64.
func (u *Uint64Hex) UnmarshalJSON(data []byte) error {
	var hexStr string
	if err := json.Unmarshal(data, &hexStr); err != nil {
		return err
	}
	// Ethereum JSON-RPC returns numeric values as hex strings prefixed with "0x".
	if len(hexStr) > 1 && strings.HasPrefix(hexStr, "0x") {
		val, err := strconv.ParseUint(hexStr[2:], 16, 64)
		if err != nil {
			return fmt.Errorf("invalid hex number %q: %v", hexStr, err)
		}
		*u = Uint64Hex(val)
		return nil
	}
	return fmt.Errorf("expected hex string, got %q", hexStr)
}
