package internal

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type Parser interface {
	// last parsed block
	GetCurrentBlock() int

	// add address to observer
	Subscribe(address string) bool

	// list of inbound or outbound transactions for an address
	GetTransactions(address string) []Transaction
}

type EthParser struct {
	curBlock     int
	subsAddrs    map[string]bool
	fetchBlockCh chan int
	doneBlockCh  chan int
	trans        TransStorage
	tk           *time.Ticker
	ec           EthEndpointClientI
	mu           sync.RWMutex
	wg           sync.WaitGroup
	stopCh       chan struct{}

	relay       bool // if relay, will presist transactions to file
	relayBlocks []relayBlock
}

func NewEthParser() *EthParser {
	ethp := &EthParser{
		subsAddrs:    make(map[string]bool),
		trans:        NewInMemoryStorage(),
		ec:           NewEthEndpointClient("https://cloudflare-eth.com"),
		tk:           time.NewTicker(10 * time.Second),
		fetchBlockCh: make(chan int, 10),
		doneBlockCh:  make(chan int, 10),
		stopCh:       make(chan struct{}),

		relay: os.Getenv("RELAY") == "true",
	}

	if os.Getenv("MOCK") == "true" {
		ethp.ec = NewMockEthEndpointClient()
	}

	return ethp
}

func (ep *EthParser) GetCurrentBlock() int {
	ep.mu.RLock()
	defer ep.mu.RUnlock()
	return ep.curBlock
}

func (ep *EthParser) SetCurrentBlock(blockNumber int) {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	ep.curBlock = blockNumber
}

func (ep *EthParser) Subscribe(address string) bool {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	if _, exists := ep.subsAddrs[address]; exists {
		return false
	}
	ep.subsAddrs[address] = true
	return true
}

func (ep *EthParser) GetTransactions(address string) []Transaction {
	ts, err := ep.trans.GetAddressTrans(address)
	if err != nil {
		return nil
	}
	return ts
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

type Transaction struct {
	Hash        string `json:"hash"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	BlockNumber string `json:"blockNumber"`
}

func (ep *EthParser) fetchTrans(blockNumber int) {
	defer ep.wg.Done()
	block, err := ep.ec.GetBlockByNumber(blockNumber)
	if err != nil {
		return
	}

	if ep.relay {
		ep.relayBlocks = append(ep.relayBlocks, relayBlock{blockNumber, block})
	}

	transMap := map[string][]Transaction{}

	for _, tx := range block.Result.Transactions {
		// fmt.Printf("hash: %s, from: %s, to: %s\n", tx.Hash, tx.From, tx.To)
		if _, exists := ep.subsAddrs[tx.From]; exists {
			transMap[tx.From] = append(transMap[tx.From], tx)
		}

		if _, exists := ep.subsAddrs[tx.To]; exists {
			transMap[tx.To] = append(transMap[tx.To], tx)
		}
	}

	for address, txs := range transMap {
		ep.trans.BatchAdd(address, txs)
	}
}

func (ep *EthParser) Run() {
	var wg sync.WaitGroup

	wg.Add(3)

	go func() {
		defer wg.Done()
		for bn := range ep.doneBlockCh {
			ep.SetCurrentBlock(bn)
		}
	}()

	go func() {
		defer wg.Done()
		for bn := range ep.fetchBlockCh {
			ep.wg.Add(1)
			go ep.fetchTrans(bn)
			ep.doneBlockCh <- bn
		}
	}()

	initblock, err := ep.ec.GetBlockNumber()
	if err != nil {
		log.Fatalf("failed to get init block number: %v", err)
	}
	ep.SetCurrentBlock(initblock)

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ep.tk.C:
				blockNumber, err := ep.ec.GetBlockNumber()
				if err != nil {
					continue
				}
				log.Printf("current block: %d", blockNumber)
				if blockNumber > ep.curBlock {
					for i := ep.curBlock + 1; i <= blockNumber; i++ {
						ep.fetchBlockCh <- i
					}
				}
			case <-ep.stopCh:
				close(ep.fetchBlockCh)
				close(ep.doneBlockCh)
				ep.tk.Stop()
				return
			}
		}
	}()

	wg.Wait()
}

func (ep *EthParser) Stop() {
	close(ep.stopCh)
	ep.wg.Wait()
}

func (ep *EthParser) SaveRelay() {
	f, err := os.Create("relay.json")

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	json.NewEncoder(f).Encode(ep.relayBlocks)

	log.Printf("relay saved")
}

type relayBlock struct {
	BlockNumber int
	Block       *GetBlockByNumberResp
}
