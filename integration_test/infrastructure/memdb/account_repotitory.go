package memdb

import (
	"github.com/hashicorp/go-memdb"
	"github.com/ymz-ncnk/idempotency-go/integration_test/domain"
)

// NewAccountRepository creates a new AccountRepository backed by the given
// in-memory transaction.
func NewAccountRepository(tx *memdb.Txn) domain.AccountRepository {
	return AccountRepository{tx}
}

// AccountRepository provides access to Account data in the in-memory DB.
// All operations are performed within the provided transaction.
type AccountRepository struct {
	tx *memdb.Txn
}

// Add inserts a new Account into the in-memory DB.
func (r AccountRepository) Add(account domain.Account) (err error) {
	return r.tx.Insert(MemDBAccountsTableName, account)
}

// Get retrieves an Account by its ID from the in-memory DB.
func (r AccountRepository) Get(id string) (account domain.Account, err error) {
	acc, err := r.tx.First(MemDBAccountsTableName, "id", id)
	if err != nil {
		return
	}
	account = acc.(domain.Account)
	return
}

// Update inserts or overwrites an existing Account in the in-memory DB.
func (r AccountRepository) Update(account domain.Account) (err error) {
	return r.tx.Insert(MemDBAccountsTableName, account)
}
