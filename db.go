package ethparser

import "sync"

type Database interface {
	GetTransactions(address string) []Transaction
	AddTransaction(address string, tx Transaction)
}

type database struct {
	mu           sync.RWMutex
	transactions map[string][]Transaction
}

func NewDatabase() Database {
	return &database{
		mu:           sync.RWMutex{},
		transactions: make(map[string][]Transaction),
	}
}

func (d *database) GetTransactions(address string) []Transaction {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.transactions[address]
}

func (d *database) AddTransaction(address string, tx Transaction) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.transactions[address] = append(d.transactions[address], tx)
}
