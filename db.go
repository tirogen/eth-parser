package ethparser

type Database interface {
	GetTransactions(address string) []Transaction
	AddTransaction(address string, tx Transaction)
}

type database struct {
	transactions map[string][]Transaction
}

func NewDatabase() Database {
	return &database{
		transactions: make(map[string][]Transaction),
	}
}

func (d *database) GetTransactions(address string) []Transaction {
	return d.transactions[address]
}

func (d *database) AddTransaction(address string, tx Transaction) {
	d.transactions[address] = append(d.transactions[address], tx)
}
