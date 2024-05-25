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
	subs        SubscribeStorage
	trans       TransStorage
	ec          EthEndpointClientI
	tk          *time.Ticker
	stopCh      chan struct{}
	relayBlocks []relayBlock
	latestBlock int
	mu          sync.RWMutex
	relay       bool
}

func NewEthParser() *EthParser {
	ethp := &EthParser{
		subs:  NewInMemorySubsStorage(),
		trans: NewInMemoryStorage(),
		ec:    NewEthEndpointClient("https://cloudflare-eth.com"),

		tk:     time.NewTicker(12 * time.Second),
		stopCh: make(chan struct{}),

		relay: RELAY_FLAG == "true",
	}

	if MOCK_FLAG == "true" {
		ethp.ec = NewMockEthEndpointClient()
	}

	return ethp
}

func (ep *EthParser) GetCurrentBlock() int {
	ep.mu.RLock()
	defer ep.mu.RUnlock()
	return ep.latestBlock
}

func (ep *EthParser) SetLatestBlock(bn int) {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	if bn <= ep.latestBlock {
		return
	}
	ep.latestBlock = bn
}

func (ep *EthParser) Subscribe(addr string) bool {
	return ep.subs.Subscribe(addr)
}

func (ep *EthParser) GetTransactions(addr string) []Transaction {
	ts, err := ep.trans.GetAddressTrans(addr)
	if err != nil {
		return nil
	}
	return ts
}

type GetBlockByNumberResp struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		Number       string        `json:"number"`
		Hash         string        `json:"hash"`
		Transactions []Transaction `json:"transactions"`
	} `json:"result"`
	ID int `json:"id"`
}

type GetBlockNumberResp struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  string `json:"result"`
	ID      int    `json:"id"`
}

type Transaction struct {
	Hash        string `json:"hash"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	BlockNumber string `json:"blockNumber"`
}

func (ep *EthParser) fetchTrans(blockNumber int) {
	log.Printf("[parser] Fetching block %d", blockNumber)
	block, err := ep.ec.GetBlockByNumber(blockNumber)
	if err != nil {
		log.Printf("Error fetching block %d: %v", blockNumber, err)
		return
	}

	if ep.relay {
		ep.relayBlocks = append(ep.relayBlocks, relayBlock{block, blockNumber})
	}

	transMap := make(map[string][]Transaction, len(block.Result.Transactions)*2) // init a length of txs, avoid re-alloc

	for _, tx := range block.Result.Transactions {
		// fmt.Printf("hash: %s, from: %s, to: %s\n", tx.Hash, tx.From, tx.To)
		if exists := ep.subs.IsSubscribed(tx.From); exists {
			transMap[tx.From] = append(transMap[tx.From], tx)
		}

		if exists := ep.subs.IsSubscribed(tx.To); exists {
			transMap[tx.To] = append(transMap[tx.To], tx)
		}
	}

	for address, txs := range transMap {
		if err := ep.trans.BatchAdd(address, txs); err != nil {
			log.Printf("Error saving txs for %s: %v", address, err)
		}
	}
}

func (ep *EthParser) Run() {
	initblock, err := ep.ec.GetBlockNumber()
	if err != nil {
		log.Fatalf("failed to get init block number: %v", err)
	}

	log.Printf("[parser] Init block: %d", initblock)
	ep.SetLatestBlock(initblock)

	for {
		select {
		case <-ep.tk.C:
			blockNumber, err := ep.ec.GetBlockNumber()
			if err != nil {
				continue
			}
			log.Printf("[parser] Parsed: %d, Ethlast: %d", ep.latestBlock, blockNumber)
			if blockNumber > ep.latestBlock {
				for i := ep.latestBlock + 1; i <= blockNumber; i++ {
					ep.fetchTrans(i)
				}
			}
			ep.SetLatestBlock(blockNumber)
		case <-ep.stopCh:
			return
		}
	}
}

func (ep *EthParser) Stop() {
	log.Println("[parser] Closing ticker")
	close(ep.stopCh)
	ep.tk.Stop()
	log.Println("[parser] Stopped")
}

func (ep *EthParser) SaveRelay() {
	f, err := os.Create(RELAY_FILE)

	if err != nil {
		log.Printf("Error creating relay file: %v", err)
		return
	}

	//nolint:errcheck
	defer f.Close()

	if err := json.NewEncoder(f).Encode(ep.relayBlocks); err != nil {
		log.Printf("Error saving relay: %v", err)
		return
	}

	log.Printf("Relay saved to %s", RELAY_FILE)
}

type relayBlock struct {
	Block       *GetBlockByNumberResp
	BlockNumber int
}
