package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemorySubsStorage_Subscribe_IsSubscribed(t *testing.T) {
	storage := NewInMemorySubsStorage()

	address := "0x123456789"
	subscribed := storage.Subscribe(address)
	assert.True(t, subscribed)

	isSubscribed := storage.IsSubscribed(address)
	assert.True(t, isSubscribed)
}

func TestInMemoryTransStorage_Add_BatchAdd_GetAddressTrans(t *testing.T) {
	storage := NewInMemoryStorage()

	address := "0xabcdef123"
	tx := Transaction{
		From: "0x123456789abcdef123",
		To:   "0xabcdef123123456789",
	}
	err := storage.Add(address, tx)
	assert.NoError(t, err)

	txs := []Transaction{
		{
			From: "0x123456789abcdef123",
			To:   "0xabcdef123123456789",
		},
		{
			From: "0x123456789abcdef123",
			To:   "0xabcdef123123456789",
		},
		{
			From: "0x123456789abcdef123",
			To:   "0xabcdef123123456789",
		},
	}
	err = storage.BatchAdd(address, txs)
	assert.NoError(t, err)

	retrievedTxs, err := storage.GetAddressTrans(address)
	assert.NoError(t, err)
	assert.Len(t, retrievedTxs, 4)
}
