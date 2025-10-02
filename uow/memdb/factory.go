package memdb

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/ymz-ncnk/idempotency-go"
)

// RepositoryBundleFactory is a function that accepts a transaction context (Tx)
// and constructs the full application and idempotency repository bundle (T)
// for that specific transaction.
type RepositoryBundleFactory[T idempotency.UOWRepos] func(tx *memdb.Txn) T
