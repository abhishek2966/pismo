package inmemory

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/abhishek2966/pismo/v2/pkg/model"
	"github.com/abhishek2966/pismo/v2/pkg/store"
)

// db type is used to store accounts and transactions.
// Accounts are grouped in clusters.
// Each cluster to have max clusterSize no. of accounts.
// Number of clusters grow as accounts grow.
// ops field store the informataion of operation types.
// txns store the transactions.
// latestAcctID field is used to uniquely assign an id to each account.
type db struct {
	clusters     []*cluster
	ops          model.Operations
	txns         transactions
	clusterSize  uint64
	latestAcctID atomic.Uint64
}

// cluster type is used to store a max db.clusterSize number of accounts.
// m field is RW mutex.
// Its read lock is obtained before retieving any account info of the cluster.
// Its write lock is obtained before adding a new account to the cluster.
// acctMap field holds information mapped from account
type cluster struct {
	m       *sync.RWMutex
	acctMap map[uint64]model.Account
}

// transactions type is used to store transactions.
// txnID field ensures every transaction gets a unique ID.
// acctTxnMap maps account ID to safeTxn which is the list of its transactions.
// m is RW mutex and its write lock is obtained
// only before adding first transaction of an account ID.
// Each safeTxn has its own mutex whose lock is obtained when a transaction
// corresponding to an account is added to safeTxn.
type transactions struct {
	m           *sync.RWMutex
	txnID       atomic.Uint64
	acctTxnsMap map[uint64]safeTxn
}

// safeTxn type has a field data which holds all the transactions for an account.
type safeTxn struct {
	m    *sync.Mutex
	data *[]model.Transaction
}

// InitDB initializes a db instance and returns it as a store.Storer.
func InitDB(clusterSize uint64) store.Storer {
	s := db{
		clusters: make([]*cluster, 1, model.ClustersCapacity),
		ops:      model.OpsTypes,
		txns: transactions{
			m:           new(sync.RWMutex),
			acctTxnsMap: map[uint64]safeTxn{},
		},
		clusterSize: clusterSize,
	}
	s.clusters[0] = &cluster{
		m:       new(sync.RWMutex),
		acctMap: make(map[uint64]model.Account, s.clusterSize),
	}
	return &s
}

// CreateAccount creates an account and adds it to the db.
// db.lstestAcctID is atomically incremented to get the
// account id for this new account.
// Using account id and clustersize, right cluster is decided for account addition.
// A next(new) cluster is also created if this account is the first account in this cluster.
// An account is added to the cluster only after obtaining write lock on the cluster.
func (s *db) CreateAccount(docNum string) (account model.Account, err error) {
	if len(s.clusters) == 0 || s.clusterSize == 0 {
		err = model.ErrorStoreNotInitialized
		return
	}

	acctID := s.latestAcctID.Add(1)
	clusterIndex := int(acctID / s.clusterSize)
	if len(s.clusters) <= clusterIndex || s.clusters[clusterIndex] == nil {
		err = model.ErrorAccountAdditionRetry
	}
	currCluster := s.clusters[clusterIndex]

	currCluster.m.Lock()
	defer currCluster.m.Unlock()
	// create a new cluster as this account is the first account in this cluster
	if len(currCluster.acctMap) == 0 {
		nextCluster := &cluster{
			m:       new(sync.RWMutex),
			acctMap: make(map[uint64]model.Account, s.clusterSize),
		}
		s.clusters = append(s.clusters, nextCluster)
	}

	account.ID = acctID
	account.DocNum = docNum
	currCluster.acctMap[acctID] = account
	return
}

// FetchAccount fetches info of an account from db.
// Using the account ID and cluster size, the correct cluster is decided.
// The account is retrieved from that cluster after obtaining read lock on it.
func (s *db) FetchAccount(accountID uint64) (account model.Account, err error) {
	if s == nil {
		err = model.ErrorStoreNotInitialized
		return
	}
	clusterIndex := int(accountID / s.clusterSize)
	if len(s.clusters) <= clusterIndex || s.clusters[clusterIndex] == nil {
		err = model.ErrorAccountNotPresent
		return
	}
	currCluster := s.clusters[clusterIndex]
	currCluster.m.RLock()
	defer currCluster.m.RUnlock()
	var ok bool
	if account, ok = currCluster.acctMap[accountID]; !ok {
		err = model.ErrorAccountNotPresent
		return
	}
	return
}

// Transact adds a transaction to the db.
// Using the account ID and cluster size, the correct cluster is decided.
// The account presence is ascertained after obtaining read lock on it.
// A write lock is obtained on the list corresponding to the account id before adding this transaction.
// If this is the first transaction for the given account, then first a list is created
// after obtaining write lock on all the transactions.
func (s *db) Transact(accountID uint64, opsType int, amount float64) (txn model.Transaction, err error) {
	if s == nil {
		err = model.ErrorStoreNotInitialized
		return
	}
	clusterIndex := int(accountID / s.clusterSize)
	if len(s.clusters) <= clusterIndex || s.clusters[clusterIndex] == nil {
		err = model.ErrorAccountNotPresent
	}
	currCluster := s.clusters[clusterIndex]
	err = func() error {
		currCluster.m.RLock()
		defer currCluster.m.RUnlock()
		if _, ok := currCluster.acctMap[accountID]; !ok {
			return model.ErrorAccountNotPresent
		}
		return nil
	}()
	if err != nil {
		return
	}

	txn.ID = s.txns.txnID.Add(1)
	s.txns.m.RLock()
	_, ok := s.txns.acctTxnsMap[accountID]
	s.txns.m.RUnlock()
	// this is a new account id; needs to be inserted in the map.
	if !ok {
		s.txns.m.Lock()
		s.txns.acctTxnsMap[accountID] = safeTxn{
			m:    new(sync.Mutex),
			data: new([]model.Transaction),
		}
		s.txns.m.Unlock()
	}

	txn.Amount = amount * float64(s.ops[opsType].Mult)
	txn.AccountID = accountID
	txn.OpsID = opsType
	txn.EventDate = time.Now().Format(time.RFC3339Nano)

	// obtain the lock only on that part of transactions list which
	// correspond to incoming account id.
	s.txns.acctTxnsMap[accountID].m.Lock()
	*s.txns.acctTxnsMap[accountID].data = append(*s.txns.acctTxnsMap[accountID].data, txn)
	s.txns.acctTxnsMap[accountID].m.Unlock()
	return
}
