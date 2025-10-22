package app

import (
	"github.com/ymz-ncnk/idempo-go"

	"github.com/ymz-ncnk/idempo-go/integration_test/domain"
)

// NewRepositoryBundle creates a RepositoryBundle containing the
// Idempotency store. Other repositories (e.g., AccountRepo) can
// be added after creation.
func NewRepositoryBundle(idempotencyStore idempo.Store) RepositoryBundle {
	return RepositoryBundle{
		idempotencyStore: idempotencyStore,
	}
}

// RepositoryBundle groups all repositories needed by a UnitOfWork.
// It provides access to the domain repositories (like AccountRepo)
// and the Idempotency store required by the wrapper.
type RepositoryBundle struct {
	AccountRepo      domain.AccountRepository
	idempotencyStore idempo.Store
}

func (b RepositoryBundle) IdempotencyStore() idempo.Store {
	return b.idempotencyStore
}
