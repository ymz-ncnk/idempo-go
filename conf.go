package idempo

// Config holds all necessary external dependencies and serialization/error
// conversion logic required to initialize the Wrapper.
type Config[T UOWRepos, S any, F Failure] struct {
	// UnitOfWork manages the transactional boundary for idempotency key check
	// and business logic execution.
	UnitOfWork UnitOfWork[T]
	// SuccessSer serializes successful results (S) for storage.
	SuccessSer Serializer[S]
	// FailureSer serializes failure results (F) for storage.
	FailureSer Serializer[F]
}
