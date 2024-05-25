package internal

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
)

type EthEndpointClientI interface {
	GetBlockByNumber(blockNumber int) (*GetBlockByNumberResp, error)
	GetBlockNumber() (int, error)
}

const (
	APIMethodGetBlockByNumber = "eth_getBlockByNumber"
	APIMethodGetBlockNumber   = "eth_blockNumber"
)

type JsonAPIPayload struct {
	Params  interface{} `json:"params"`
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	ID      int         `json:"id"`
}

func (j *JsonAPIPayload) Marshal() []byte {
	data, _ := json.Marshal(j) // ignore error
	return data
}

type EthEndpointClient struct {
	httpc    *http.Client
	endpoint string
}

func NewEthEndpointClient(endpoint string) *EthEndpointClient {
	return &EthEndpointClient{http.DefaultClient, endpoint}
}

func (e *EthEndpointClient) GetBlockByNumber(blockNumber int) (*GetBlockByNumberResp, error) {
	payload := &JsonAPIPayload{
		JSONRPC: "2.0",
		Method:  APIMethodGetBlockByNumber,
		Params: []interface{}{
			"0x" + strconv.FormatInt(int64(blockNumber), 16),
			true,
		},
		ID: 1,
	}

	resp, err := e.httpc.Post(e.endpoint, "application/json", bytes.NewReader(payload.Marshal()))
	if err != nil {
		return nil, err
	}

	//nolint:errcheck
	defer resp.Body.Close()

	ret := &GetBlockByNumberResp{}
	if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func (e *EthEndpointClient) GetBlockNumber() (int, error) {
	payload := &JsonAPIPayload{
		JSONRPC: "2.0",
		Method:  APIMethodGetBlockNumber,
		Params:  []interface{}{},
		ID:      1,
	}

	resp, err := e.httpc.Post(e.endpoint, "application/json", bytes.NewReader(payload.Marshal()))
	if err != nil {
		return 0, err
	}

	//nolint:errcheck
	defer resp.Body.Close()

	var result GetBlockNumberResp
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	hexBlockNumber := result.Result
	blockNumber, err := strconv.ParseInt(hexBlockNumber[2:], 16, 64)
	if err != nil {
		return 0, err
	}

	return int(blockNumber), nil
}
