package idempotency

import (
	"context"
	"fmt"
)

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
type Action[T, I, S any] func(ctx context.Context, repos T, input I) (S, error)

// ErrorToFail defines the function that converts a Go 'error' into the
// storable failure output ('F').
type ErrorToFail[F any] func(err error) (F, bool)

// NewWrapper creates a new instance of the execution Wrapper.
//
// It configures the Wrapper with the necessary components: the UnitOfWork
// for transactional data access, the Manager for handling persistence details,
// and the error mapping function to correctly categorize execution failures.
func NewWrapper[T UOWRepos, I Hasher, S, F any](uow UnitOfWork[T],
	manager Manager[S, F],
	errorToFail ErrorToFail[F],
) Wrapper[T, I, S, F] {
	return Wrapper[T, I, S, F]{
		uow:         uow,
		manager:     manager,
		errorToFail: errorToFail,
	}
}

// Wrapper is the core type that enforces idempotency for a protected Action.
//
// It manages the execution flow: checking the store for prior results,
// running the Action, and atomically persisting the final result (success or
// failure) alongside the Action's execution within a single Unit of Work.
//
// T is the type representing the repository bundle accessible within the
// UnitOfWork.
// I is the Action input type, which must implement the Hasher interface.
// S is the type of the successful output.
// F is the type of the failure output.
type Wrapper[T UOWRepos, I Hasher, S, F any] struct {
	uow         UnitOfWork[T]
	manager     Manager[S, F]
	errorToFail ErrorToFail[F]
}

// Wrap executes the provided Action idempotently.
//
//  1. It calculates a hash of the input (I).
//  2. Executes the UnitOfWork (UOW):
//     a. Checks the Store for a record associated with idempotencyKey. If
//     found, and its hash is equal to the hash of the input (I) returns the
//     stored result.
//     b. If no record is found, executes the core Action.
//     c. If the Action succeeds, saves the success output.
//     d. If the Action fails, with errorToFail it tries to get and persist a
//     failure output.
//  3. The UOW ensures the Action's side effects and the idempotency record
//     persistence are completed together or roll back completely.
func (w Wrapper[T, I, S, F]) Wrap(ctx context.Context, idempotencyKey string,
	input I,
	action Action[T, I, S],
) (successOutput S, err error) {
	hash, err := input.Hash()
	if err != nil {
		err = fmt.Errorf("idempotency wrapper failed to calculate input hash: %w", err)
		return
	}
	execErr := w.uow.Execute(func(repos T) (fnErr error) {
		// Idempotency Check
		var ok bool
		ok, successOutput, fnErr = w.manager.AlreadyProcessed(ctx, idempotencyKey,
			hash, repos.IdempotentStore())
		if ok || fnErr != nil {
			return
		}
		// Execute Action
		successOutput, fnErr = action(ctx, repos, input)
		if fnErr != nil {
			// Handle Failure: Business or System Error
			failOutput, isBusinessError := w.errorToFail(fnErr)
			if isBusinessError {
				// Business logic failure (e.g., OCC failed, Stock unavailable). Save
				// the fail record.
				if storeErr := w.manager.SaveFailOutput(ctx, idempotencyKey, hash,
					failOutput, repos.IdempotentStore()); storeErr != nil {
					fnErr = NewFailureOutputStoreError(storeErr, fnErr)
				} else {
					err = fnErr
					fnErr = nil
				}
			}
			return
		}
		// Action SUCCEEDED. Save the success record.
		if storeErr := w.manager.SaveSuccessOutput(ctx, idempotencyKey, hash,
			successOutput, repos.IdempotentStore()); storeErr != nil {
			fnErr = NewSuccessOutputStoreError(storeErr)
		}
		return
	})
	if execErr != nil {
		err = execErr
	}
	return
}
