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
	latestBlock  int
	fetchedBlock int // fetched block is the last block that was parsed

	subs         SubscribeStorage
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
		subs:  NewInMemorySubsStorage(),
		trans: NewInMemoryStorage(),
		ec:    NewEthEndpointClient("https://cloudflare-eth.com"),

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
	return ep.fetchedBlock
}

func (ep *EthParser) SetLatestBlock(bn int) {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	if bn <= ep.latestBlock {
		return
	}
	ep.latestBlock = bn
}

func (ep *EthParser) SetFetchedBlock(bn int) {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	if bn <= ep.fetchedBlock {
		return
	}
	ep.fetchedBlock = bn
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
		if exists := ep.subs.IsSubscribed(tx.From); exists {
			transMap[tx.From] = append(transMap[tx.From], tx)
		}

		if exists := ep.subs.IsSubscribed(tx.To); exists {
			transMap[tx.To] = append(transMap[tx.To], tx)
		}
	}

	for address, txs := range transMap {
		ep.trans.BatchAdd(address, txs)
	}
}

func (ep *EthParser) Run() {
	ep.wg.Add(1)
	go func() {
		defer ep.wg.Done()
		for bn := range ep.doneBlockCh {
			ep.SetFetchedBlock(bn)
		}
	}()

	ep.wg.Add(1)
	go func() {
		defer ep.wg.Done()
		for bn := range ep.fetchBlockCh {
			ep.wg.Add(1)
			log.Printf("[parser] Fetching block %d\n", bn)
			go ep.fetchTrans(bn)
			ep.doneBlockCh <- bn
		}
	}()

	initblock, err := ep.ec.GetBlockNumber()
	if err != nil {
		log.Fatalf("failed to get init block number: %v", err)
	}

	ep.SetLatestBlock(initblock)
	ep.SetFetchedBlock(initblock)

	for {
		select {
		case <-ep.tk.C:
			blockNumber, err := ep.ec.GetBlockNumber()
			if err != nil {
				continue
			}
			log.Printf("[parser] Latest: %d, Ethlast: %d, Parsed: %d\n", ep.latestBlock, blockNumber, ep.fetchedBlock)
			if blockNumber > ep.latestBlock {
				for i := ep.latestBlock + 1; i <= blockNumber; i++ {
					ep.fetchBlockCh <- i
				}
			}
			ep.SetLatestBlock(blockNumber)

		case <-ep.stopCh:
			return
		}
	}
}

func (ep *EthParser) Stop() {
	log.Println("[parser] Closing fetchBlockCh/doneBlockCh/ticker")
	close(ep.fetchBlockCh)
	close(ep.doneBlockCh)
	close(ep.stopCh)
	ep.tk.Stop()

	log.Println("[parser] Waiting for workers...")
	ep.wg.Wait()
	log.Println("[parser] Stopped")
}

func (ep *EthParser) SaveRelay() {
	f, err := os.Create("testdata/relay.json")

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	json.NewEncoder(f).Encode(ep.relayBlocks)

	log.Printf("Relay saved to testdata/relay.json")
}

type relayBlock struct {
	BlockNumber int
	Block       *GetBlockByNumberResp
}
