package idempo

// UOWRepos is the constraint interface required by the generic UnitOfWork.
// Any type T passed to UnitOfWork must implement this method.
type UOWRepos interface {
	IdempotencyStore() Store
}

// UnitOfWork defines a single unit of work for an application transaction.
type UnitOfWork[T UOWRepos] interface {
	// Execute runs the provided function (fn) within a single database
	// transaction. It automatically handles BEGIN, COMMIT, and ROLLBACK.
	Execute(fn func(repos T) error) error
}
