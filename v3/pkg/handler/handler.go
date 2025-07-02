package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/abhishek2966/pismo/v3/pkg/model"
	"github.com/abhishek2966/pismo/v3/pkg/store"
)

type handler struct {
	store  store.Storer
	logger io.Writer
}

// InitHandler returns a handler with its store and logger initialized.
// It sets the logger to os.Stdout if nil logger argument is provided.
func InitHandler(s store.Storer, logger io.Writer) *handler {
	h := &handler{
		store:  s,
		logger: logger,
	}
	if h.logger == nil {
		h.logger = os.Stdout
	}
	return h
}

// CreateAccount handles creation of a new account.
// Responds 400 status for bad payload.
// Responds 500 for all other error scenario like store not initialized.
// Responds 200 with account info for successfully created acocount.
func (h *handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	payload, _ := io.ReadAll(r.Body)
	var acct model.Account
	err := json.Unmarshal(payload, &acct)
	if err != nil {
		fmt.Fprintf(h.logger, "unmarshal error: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	acct, err = h.store.CreateAccount(acct.DocNum)

	if err != nil {
		fmt.Fprintf(h.logger, "account create error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, _ := json.Marshal(acct)

	w.Write(resp)
}

// FetchAccount handles retrieval of account information.
// Responds 400 status for bad path parameter.
// Responds 500 for all other error scenario like account not present.
// Responds 200 with account info for successfully retrieved account.
func (h *handler) FetchAccount(w http.ResponseWriter, r *http.Request) {
	accountIDString := r.PathValue("accountId")

	accountID, err := strconv.ParseUint(accountIDString, 10, 64)
	if err != nil {
		fmt.Fprintf(h.logger, "parse error: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	acct, err := h.store.FetchAccount(accountID)
	if err != nil {
		fmt.Fprintf(h.logger, "account retrieval error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, _ := json.Marshal(acct)
	w.Write(resp)
}

// Transact handles the transaction.
// Responds 400 status for bad payload.
// Responds 500 for all other error scenario like account not present.
// Responds 200 with transaction info for successful transaction.
func (h *handler) Transact(w http.ResponseWriter, r *http.Request) {
	payload, _ := io.ReadAll(r.Body)
	var txn model.Transaction
	err := json.Unmarshal(payload, &txn)
	if err != nil {
		fmt.Fprintf(h.logger, "unmarshal error: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	txn, err = h.store.Transact(txn.AccountID, txn.OpsID, txn.Amount)

	if err != nil {
		fmt.Fprintf(h.logger, "transaction error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, _ := json.Marshal(txn)
	w.Write(resp)
}
