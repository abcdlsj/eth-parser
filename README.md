### ETH Parser

## Structure

```
├── Makefile
├── README.md
├── cmd
│   ├── cli
│   │   └── main.go
│   └── server
│       └── main.go
├── go.mod
├── internal
│   ├── endpoint.go
│   ├── env.go
│   ├── mock_endpoint.go
│   ├── parser.go
│   └── storage.go
└── testdata
    └── relay.json
```

- `cmd` - contains CLI and server binaries
  - `server` - daemon service, contains HTTP handler implementation
  - `cli` - parse command line and call server HTTP
- `internal`
  - `endpoint.go` - endpoint HTTP client implementation, provides JSON API wrapper and interface definition
  - `mock_endpoint.go` - implementation of mock endpoint HTTP client
  - `env.go` - defines and initializes environment variables
  - `parser.go` - implements parser, provides `Run` function to loop check for new blocks (every 12 seconds)
  - `storage.go` - abstracts storage (`TransStorage` | `SubscribeStorage`) interface, implements in-memory storage
- `unit tests` - contains unit tests
  - `endpoint_test.go` - unit tests for `endpoint`
  - `storage_test.go` - unit tests for `parser`

unit tests: `go test ./...`

## Usage

1. Build 
```shell
make all
```
2. Start service
```shell
./bin/ethparser-server
```

Support env flags:
- `PORT` - port to listen on
- `RELAY` - if set, starts relay mode, storing all fetched blocks to a file when the server closes
- `RELAY_FILE` - file to save blocks to, default: `testdata/relay.json`
- `MOCK` - if set, replaces `https://cloudflare-eth.com` endpoint with mock endpoint, which reads the relay file to simulate

Supported endpoints:
- `GET /getCurrentBlock` - get current block
- `POST /subscribe` - subscribe to new block
  - `payload` - `{"address": "0x8216874887415e2650d12d53ff53516f04a74fd7"}`
- `GET /getTransactions` - get transactions
  - `query` - `address=0x8216874887415e2650d12d53ff53516f04a74fd7`
- `POST /saveRelay` - save relayed block to file (requires `RELAY=true` env flag)

3. Run CLI or make HTTP calls

CLI
```shell
> ./bin/ethparser-cli -h 

Usage of ./bin/ethparser-cli:
  getCurrentBlock
  subscribe <ADDRESS>
  getTransactions <ADDRESS>
```
- `SERVER_URL` - server URL, default: `http://localhost:8080`

HTTP call
```shell
curl -X GET 'http://localhost:8080/getCurrentBlock'
```

## Test 
Can test with a relay file I had saved to `testdata/relay.json`

```fish
make test
```

> This command will initiate the server in relay mode. It will then prompt you to subscribe to an address using the CLI. After 20 seconds, it will fetch transactions using the CLI, and the process will be completed.

## Design
1. `storage` design
   
   `InMemoryTransStorage` use `map[string][]Transaction` to store transactions.
   key: address, value: transactions (contains inbound and outbound transactions)
   set complexity: O(1)
   get complexity: O(1)

   `InMemorySubsStorage` use `map[string]bool` to store subscribed addresses.
   key: address, value: true
   set complexity: O(1)
   get complexity: O(1)

2. `endpoint` design
   Masking the external perception of jsonAPI, endpoint is responsible for communicating with the eth service internally, which makes it easier to make other changes later on.

3. `parser` design
   `Run` function initiates a loop to check for new blocks every 12 seconds. It compares the latest block locally with the incoming block. If the local block is outdated, it synchronously fetches the new block (refer to `FAQ 2`) and updates the local latest block.

## FAQ

1. Why implement relay mode? What does it mean?

`Relay` means `中继` in Chinese, It allows recording and replaying data, functioning as a proxy for `cloudflare-eth.com`.

I couldn't find a good way to self-test the `parser` because I couldn't find suitable addresses to subscribe to and get transactions. So, I use relay mode to simulate the process.

2. Why remove the concurrent fetch block?

For the first few commits, I implemented concurrent fetch blocks.
but ETH generates a new block every 12 seconds, so there is no need for concurrent fetches.

## Futures

There also have some points that could be improved:
1. Add daemon configuration
   - `MacOS`: `plist` file
   - `Linux`: systemd service