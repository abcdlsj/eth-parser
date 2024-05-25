package internal

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEthEndpointClient_GetBlockNumber_GetBlockByNumber(t *testing.T) {
	httpClient := http.DefaultClient

	endpoint := "https://cloudflare-eth.com"
	client := &EthEndpointClient{httpc: httpClient, endpoint: endpoint}

	lastBn, err := client.GetBlockNumber()
	assert.NoError(t, err)
	assert.Greater(t, lastBn, 0)

	resp, err := client.GetBlockByNumber(lastBn)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	bn, err := strconv.ParseInt(resp.Result.Number[2:], 16, 64)
	assert.NoError(t, err)
	assert.Equal(t, lastBn, int(bn))
}
