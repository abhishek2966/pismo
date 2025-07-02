package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/abhishek2966/pismo/v2/pkg/model"
)

// mockStrore implements store.Storer
type mockStore struct {
	account model.Account
	err     error
	txn     model.Transaction
}

func (m mockStore) CreateAccount(docNum string) (account model.Account, err error) {
	return m.account, m.err
}

func (m mockStore) FetchAccount(accountID uint64) (account model.Account, err error) {
	return m.account, m.err
}
func (m mockStore) Transact(accountID uint64, opsType int, amount float64) (txn model.Transaction, err error) {
	return m.txn, m.err
}

func TestCreateAccount(t *testing.T) {

	testCases := []struct {
		name       string
		payload    string
		acct       model.Account
		acctErr    error
		statusWant int
		respWant   string
	}{
		{
			name:       "happy path",
			payload:    `{"document_number": "12345678900"}`,
			acct:       model.Account{ID: 1, DocNum: "12345678900"},
			statusWant: http.StatusOK,
			respWant:   `{"account_id":1,"document_number":"12345678900"}`,
		},
		{
			name:       "wrong payload",
			payload:    `{"document_number": "12345678900"`,
			statusWant: http.StatusBadRequest,
			respWant:   "unexpected end of JSON input\n",
		},
	}
	url := "http://test.com/accounts"
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, url, strings.NewReader(tc.payload))
			s := mockStore{tc.acct, tc.acctErr, model.Transaction{}}
			h := InitHandler(s, nil)
			h.CreateAccount(w, r)
			response := w.Result()
			if response.StatusCode != tc.statusWant {
				t.Errorf("status got: %v, want: %v", response.StatusCode, tc.statusWant)
			}
			body, _ := io.ReadAll(response.Body)
			if string(body) != tc.respWant {
				t.Errorf("response got: %s, want: %v", body, tc.respWant)
			}
		})
	}
}
