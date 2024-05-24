package parser

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ethHttpClient = NewEthHttpClient()
)

type Transaction struct {
	// Define the structure for transaction details
	Hash        string `json:"hash"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	BlockNumber int    `json:"blockNumber"`
}

type Parser interface {
	// last parsed block
	GetCurrentBlock() int

	// add address to observer
	Subscribe(address string) bool

	// list of inbound or outbound transactions for an address
	GetTransactions(address string) []Transaction
}

type EthParser struct {
	currentBlock        int
	subscribedAddresses map[string]bool
	transactions        map[string][]Transaction
	mu                  sync.Mutex
}

func NewEthereumParser() *EthParser {
	return &EthParser{
		subscribedAddresses: make(map[string]bool),
		transactions:        make(map[string][]Transaction),
	}
}

func (ep *EthParser) GetCurrentBlock() int {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	return ep.currentBlock
}

func (ep *EthParser) Subscribe(address string) bool {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	if _, exists := ep.subscribedAddresses[address]; exists {
		return false
	}
	ep.subscribedAddresses[address] = true
	return true
}

func (ep *EthParser) GetTransactions(address string) []Transaction {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	return ep.transactions[address]
}

type GetBlockByNumberResp struct {
	ID      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		Number       string        `json:"number"`
		Hash         string        `json:"hash"`
		Transactions []Transaction `json:"transactions"`
	} `json:"result"`
}

type GetBlockNumberResp struct {
	ID      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  string `json:"result"`
}

type EthHttpClient struct {
	httpc *http.Client
}

func NewEthHttpClient() *EthHttpClient {
	return &EthHttpClient{
		httpc: http.DefaultClient,
	}
}

func (ec *EthHttpClient) GetBlockByNumber(blockNumber int) (*GetBlockByNumberResp, error) {
	payload := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x%x", false],"id":1}`, blockNumber)

	resp, err := ec.httpc.Post("https://cloudflare-eth.com",
		"application/json",
		strings.NewReader(payload))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	ret := &GetBlockByNumberResp{}
	if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
		return nil, err
	}

	return ret, nil
}
func (ec *EthHttpClient) GetBlockNumber() (int, error) {
	resp, err := ec.httpc.Post("https://cloudflare-eth.com",
		"application/json",
		strings.NewReader(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`))

	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result GetBlockNumberResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	hexBlockNumber := result.Result
	blockNumber, err := strconv.ParseInt(hexBlockNumber[2:], 16, 64)
	if err != nil {
		return 0, err
	}

	return int(blockNumber), nil
}

func (ep *EthParser) fetchTransactions(blockNumber int) {
	block, err := ethHttpClient.GetBlockByNumber(blockNumber)
	if err != nil {
		fmt.Println("Error getting block:", err)
		return
	}

	for _, tx := range block.Result.Transactions {
		if _, exists := ep.subscribedAddresses[tx.From]; exists {
			ep.transactions[tx.From] = append(ep.transactions[tx.From], tx)
		}
		if _, exists := ep.subscribedAddresses[tx.To]; exists {
			ep.transactions[tx.To] = append(ep.transactions[tx.To], tx)
		}
	}
}

func main() {
	parser := NewEthereumParser()

	ticker := time.NewTicker(10 * time.Second)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for range ticker.C {
			blockNumber, err := ethHttpClient.GetBlockNumber()
			if err != nil {
				fmt.Println("Error getting block number:", err)
				continue
			}

			parser.mu.Lock()
			if blockNumber > parser.currentBlock {
				for i := parser.currentBlock + 1; i <= blockNumber; i++ {
					parser.fetchTransactions(i)
				}
				parser.currentBlock = blockNumber
			}
			parser.mu.Unlock()
		}
	}()

	wg.Wait()
}
