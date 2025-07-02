package model

import (
	"fmt"
)

type Account struct {
	ID     uint64 `json:"account_id"`
	DocNum string `json:"document_number"`
}

type Operations map[int]struct {
	Mult  int
	Descr string
}

var OpsTypes = Operations{
	1: {-1, "Normal Purchase"},
	2: {-1, "Purchase with installments"},
	3: {-1, "Withdrawal"},
	4: {1, "Credit Voucher"},
}

type Transaction struct {
	ID        uint64  `json:"transaction_id"`
	AccountID uint64  `json:"account_id"`
	OpsID     int     `json:"operation_type_id"`
	Amount    float64 `json:"amount"`
	EventDate string  `json:"event_date"`
}

var ErrorStoreNotInitialized = fmt.Errorf("store not initialized")
var ErrorAccountNotPresent = fmt.Errorf("account not present in store")
var ErrorOperationNotAllowed = fmt.Errorf("operation not allowed")
var ErrorAccountAdditionRetry = fmt.Errorf("retry adding this account")

const ClustersCapacity = 1000
