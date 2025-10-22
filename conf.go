package idempo

// Config holds all necessary external dependencies and serialization/error
// conversion logic required to initialize the Wrapper.
type Config[T UOWRepos, S, F any] struct {
	// UnitOfWork manages the transactional boundary for idempotency key check
	// and business logic execution.
	UnitOfWork UnitOfWork[T]
	// SuccessSer serializes successful results (S) for storage.
	SuccessSer Serializer[S]
	// FailureSer serializes failure results (F) for storage.
	FailureSer Serializer[F]
	// ErrorToFail maps a runtime error to a storable failure (F).
	// Returns ok=false if the error should not be persisted.
	ErrorToFail func(err error) (ok bool, failure F)
	// FailToError converts a stored failure (F) back into a Go error.
	FailToError func(failure F) error
}
