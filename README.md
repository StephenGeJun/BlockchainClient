# Simple Blockchain Client

This project is a simple blockchain client written in Go. It provides basic functionality to interact with an Ethereum-compatible blockchain (using JSON-RPC), specifically to retrieve the latest block number and fetch block details by block number. The client is exposed via a small HTTP API server.

## Features

- **Get Latest Block Number**: Retrieve the most recent block number from a blockchain node (using the `eth_blockNumber` JSON-RPC call).
- **Get Block by Number**: Fetch detailed information of a block by its number (using the `eth_getBlockByNumber` JSON-RPC call, including transactions).
- **HTTP API**: Exposes the above functionality through a simple HTTP server with two endpoints:
  - `GET /blockNumber` - returns the latest block number.
  - `GET /block/<number>` - returns the block details for the specified block number. The block number can be provided in decimal or hex format (e.g., `12345` or `0x3039`).

## Project Structure
```
BlockchainClient/
├── README.md
├── go.mod
├── main.go
├── blockchain/
│   ├── client.go
│   └── client_test.go
├── Dockerfile
└── terraform/
    ├── main.tf
    ├── variables.tf
    └── outputs.tf
```
## Getting Started

### Prerequisites
- Go 1.18+ (to build and run the application)
- Docker (for containerization)
- Terraform (for deployment)

### Building and Running Locally
1. **Clone the repository** and navigate into it.
2. **Install dependencies** (if any) by running `go mod download`. *(The project uses only standard library packages.)*
3. **Run the tests** with `go test ./...` to ensure everything is working.
4. **Build the application** using `go build -o blockchain-client .`. 
5. **Run the application**: Execute `./blockchain-client`. By default, the server will start on port 8080 and use the Polygon mainnet public RPC endpoint. You can override:
   - **RPC endpoint**: set environment variable `RPC_URL` to the desired Ethereum JSON-RPC URL (e.g., an Infura endpoint or local Ethereum node).
   - **Port**: set the `PORT` environment variable to change the HTTP server port.

### Using the HTTP API
Once the server is running, you can query it:
- **Get latest block number**: `curl http://localhost:8080/blockNumber`
  - Response: `{ "blockNumber": 12345678 }`
- **Get block by number**: `curl http://localhost:8080/block/12345678`
  - Response: JSON object with block details (block hash, parent hash, timestamp, and an array of transactions with their hash, from, to, value, etc.). For example:
  ```json
  {
    "Number": 12345678,
    "Hash": "0xabc123...",
    "ParentHash": "0xdef456...",
    "Timestamp": 1616586482,
    "Transactions": [
       {
         "Hash": "0x7890...",
         "From": "0x sender address ...",
         "To": "0x recipient address ...",
         "Value": "0xde0b6b3a7640000"
       },
       ...
    ]
  }

### Using terraform to deploy blockchain client to AWS ECS
**navigate into terraform folder and execute follow terraform commands step by step**
1. `terraform init`
2. `terraform plan --out=tf.plan --var-file=ecs.tfvars`
3. `terraform apply tf.plan`