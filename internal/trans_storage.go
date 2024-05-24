package internal

import (
	"sync"
)

type TransStorage interface {
	Add(address string, tx Transaction) error
	BatchAdd(address string, txs []Transaction) error
	GetAddressTrans(address string) ([]Transaction, error)
}

type InMemoryStorage struct {
	trans map[string][]Transaction // address -> transactions

	mu sync.RWMutex
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		trans: make(map[string][]Transaction),
	}
}

func (i *InMemoryStorage) Add(address string, tx Transaction) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.trans[address] = append(i.trans[address], tx)
	return nil
}

func (i *InMemoryStorage) BatchAdd(address string, txs []Transaction) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.trans[address] = append(i.trans[address], txs...)
	return nil
}

func (i *InMemoryStorage) GetAddressTrans(address string) ([]Transaction, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.trans[address], nil
}
