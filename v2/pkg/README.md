## Steps to run
go run main.go -p 80 -n 10000
- Flag p denotes the port number and n is for the cluster size.
## Design
1. App Memory is used for accounts/transactions storage. 
2. Accounts are grouped in Clusters in serial fashion.
3. The fully occupied clusters remain always available/safe for concurrent retrieval.
4. The cluster where an account is added is synchronised through write lock.
5. The Transactions are grouped based on account ids.
6. A particular transaction, when being added to store, only blocks the group corresponding to that account.
7. All other transaction groups remain available for concurrent writes.
## Implementation
1. A configurable cluster size is taken via flag input that defaults to 10.
1. The InMemory store is initialized with just 1 cluster.
1. The moment first account entry is added in a cluster, a next cluster is created in the store.
1. This ensures cluster addition remains race safe.
1. Before adding an account to a cluster a write lock is obtained on the cluster. Even before obtaining the write lock, store's latestAccountID counter is atomically incremented.
1. This enables 2 concurrent addition when one is the last account for current cluster and other should go to the next cluster.
1. This also demands a fine tuned value for cluster size.
1. Otherwise request is dropped in a situation wherein say cluster size is 10 -> there are 11 concurrent accounts add request -> current cluster is empty -> even before one account is added in current cluster i.e. the next cluster is yet to be created 11th account attempts to add -> 11th account should go to the next cluster, but next cluster is not yet created -> 11th account add request is dropped to ensure consistency.
