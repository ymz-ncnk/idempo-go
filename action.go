package idempotency

import "context"

// Action defines the function signature for the core idempotent operation
// to be executed by the Wrapper.
//
// An Action takes the current execution context, a dependency bundle (T)
// for data access, and the call input (I). It is expected to return
// the successful operation output (S) or a standard Go error.
//
// The Action's failure errors must be mappable by the Wrapper to a failure
// output so they can be persisted and replayed to the caller on subsequent
// attempts.
type Action[T, I, S any] func(ctx context.Context, repos T,
	idempotencyKey string, input I) (S, error)
