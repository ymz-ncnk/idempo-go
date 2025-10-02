package app

// UnitOfWork defines a single unit of work for a business transaction.
type UnitOfWork interface {
	// Execute runs the provided function (fn) within a single database transaction.
	// It automatically handles BEGIN, COMMIT, and ROLLBACK.
	Execute(fn func(repos RepositoryBundle) error) error
}
