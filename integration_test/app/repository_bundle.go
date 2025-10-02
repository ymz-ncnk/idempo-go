package app

import (
	"github.com/ymz-ncnk/idempotency-go"

	"github.com/ymz-ncnk/idempotency-go/integration_test/domain"
)

// NewRepositoryBundle creates a RepositoryBundle containing the
// Idempotency store. Other repositories (e.g., AccountRepo) can
// be added after creation.
func NewRepositoryBundle(idempotencyStore idempotency.Store) RepositoryBundle {
	return RepositoryBundle{
		idempotencyStore: idempotencyStore,
	}
}

// RepositoryBundle groups all repositories needed by a UnitOfWork.
// It provides access to the domain repositories (like AccountRepo)
// and the Idempotency store required by the wrapper.
type RepositoryBundle struct {
	AccountRepo      domain.AccountRepository
	idempotencyStore idempotency.Store
}

func (b RepositoryBundle) IdempotencyStore() idempotency.Store {
	return b.idempotencyStore
}
