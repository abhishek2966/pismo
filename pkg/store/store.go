package store

import "github.com/abhishek2966/pismo/pkg/model"

// Storer interface can be implemented by various store systems like
// inmemory, postgres, mongodb.
type Storer interface {
	CreateAccount(docNum string) (account model.Account, err error)
	FetchAccount(accountID uint64) (account model.Account, err error)
	Transact(accountID uint64, Ops int, amount float64) (txn model.Transaction, err error)
}
