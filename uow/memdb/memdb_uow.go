package memdb

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/ymz-ncnk/idempotency-go"
)

// NewMemDBUnitOfWork is the constructor for the MemDBUnitOfWork.
func NewMemDBUnitOfWork[T idempotency.UOWRepos](db *memdb.MemDB,
	factory RepositoryBundleFactory[T],
) *MemDBUnitOfWork[T] {
	return &MemDBUnitOfWork[T]{
		db:      db,
		factory: factory,
	}
}

// MemDBUnitOfWork manages the transaction lifecycle for the MemDB.
// It is generic over the Repository Bundle type (T).
type MemDBUnitOfWork[T idempotency.UOWRepos] struct {
	db *memdb.MemDB
	// factory is the external function used to construct the bundle (T)
	// for a specific transaction (tx).
	factory RepositoryBundleFactory[T]
}

// Execute starts a transaction, executes the work function, and handles
// commit/rollback.
func (u *MemDBUnitOfWork[T]) Execute(fn func(repos T) error) error {
	tx := u.db.Txn(true)
	defer tx.Abort()
	repos := u.factory(tx)
	if err := fn(repos); err != nil {
		return err
	}
	tx.Commit()
	return nil
}
