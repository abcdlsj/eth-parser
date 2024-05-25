package internal

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"
)

type MockEthEndpointClient struct {
	blocks []relayBlock
	curIdx int
}

func NewMockEthEndpointClient() *MockEthEndpointClient {
	f, err := os.Open(RELAY_FILE)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	mc := &MockEthEndpointClient{
		curIdx: -1,
	}

	json.NewDecoder(f).Decode(&mc.blocks)

	return mc
}

func (e *MockEthEndpointClient) GetBlockByNumber(blockNumber int) (*GetBlockByNumberResp, error) {
	if e.curIdx < 0 || e.curIdx >= len(e.blocks) {
		return nil, errors.New("no more blocks")
	}

	time.Sleep(1 * time.Second)

	return e.blocks[e.curIdx].Block, nil
}

func (e *MockEthEndpointClient) GetBlockNumber() (int, error) {
	n := e.curIdx

	if n >= len(e.blocks) {
		return 0, errors.New("no more blocks")
	}

	if n == -1 { // this is a init block, return fist block number - 1
		e.curIdx = 0
		return e.blocks[0].BlockNumber - 1, nil
	}

	e.curIdx++

	time.Sleep(1 * time.Second)
	return e.blocks[n].BlockNumber, nil
}
