package inmemory

import (
	"testing"

	"github.com/abhishek2966/pismo/pkg/model"
	"github.com/abhishek2966/pismo/pkg/store"
)

func TestCreateAccount(t *testing.T) {
	testCases := []struct {
		name     string
		doc      string
		str      store.Storer
		acctWant model.Account
		errWant  error
	}{
		{
			name:    "store not initialized",
			doc:     "1234",
			str:     InitDB(0),
			errWant: model.ErrorStoreNotInitialized,
		},
		{
			name:     "happy path",
			doc:      "1234",
			str:      InitDB(2),
			acctWant: model.Account{ID: 1, DocNum: "1234"},
			errWant:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			acctGot, errGot := tc.str.CreateAccount(tc.doc)
			if acctGot != tc.acctWant {
				t.Errorf("got:%v, want:%v", acctGot, tc.acctWant)
			}
			if errGot != tc.errWant {
				t.Errorf("got:%v, want:%v", errGot, tc.errWant)
			}
		})
	}
}
