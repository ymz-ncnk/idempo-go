package idempo

// Hasher defines the interface for calculating a unique, deterministic string
// signature for the Action input.
//
// The Wrapper uses the Hash to verify that a retry attempt, using the same
// idempotency key, is being made with the identical input data. This check
// prevents misuse of the idempotency key.
type Hasher interface {
	Hash() (string, error)
}
