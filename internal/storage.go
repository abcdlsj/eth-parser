package internal

import (
	"sync"
)

type TransStorage interface {
	Add(address string, tx Transaction) error
	BatchAdd(address string, txs []Transaction) error
	GetAddressTrans(address string) ([]Transaction, error)
}

type SubscribeStorage interface {
	Subscribe(address string) bool
	IsSubscribed(address string) bool
}

type InMemorySubsStorage struct {
	subsAddrs map[string]bool
	mu        sync.RWMutex
}

func NewInMemorySubsStorage() *InMemorySubsStorage {
	return &InMemorySubsStorage{
		subsAddrs: make(map[string]bool),
	}
}

func (i *InMemorySubsStorage) Subscribe(address string) bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	if _, exists := i.subsAddrs[address]; exists {
		return false
	}
	i.subsAddrs[address] = true
	return true
}

func (i *InMemorySubsStorage) IsSubscribed(address string) bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	_, exists := i.subsAddrs[address]
	return exists
}

type InMemoryTransStorage struct {
	trans map[string][]Transaction // address -> transactions

	mu sync.RWMutex
}

func NewInMemoryStorage() *InMemoryTransStorage {
	return &InMemoryTransStorage{
		trans: make(map[string][]Transaction),
	}
}

func (i *InMemoryTransStorage) Add(address string, tx Transaction) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.trans[address] = append(i.trans[address], tx)
	return nil
}

func (i *InMemoryTransStorage) BatchAdd(address string, txs []Transaction) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.trans[address] = append(i.trans[address], txs...)
	return nil
}

func (i *InMemoryTransStorage) GetAddressTrans(address string) ([]Transaction, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.trans[address], nil
}
