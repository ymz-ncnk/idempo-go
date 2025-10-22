package idempo

import (
	"context"
	"fmt"
)

// ErrorToFail defines the function that converts a Go 'error' into the
// storable failure output ('F').
type ErrorToFail[F any] func(err error) (bool, F)

// NewWrapper creates a new instance of the Wrapper.
func NewWrapper[T UOWRepos, I Hasher, S, F any](
	conf Config[T, S, F],
) Wrapper[T, I, S, F] {
	storeAdapter := NewStoreAdapter(conf.SuccessSer, conf.FailureSer,
		conf.FailToError)
	return Wrapper[T, I, S, F]{conf.UnitOfWork, storeAdapter, conf.ErrorToFail}
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
	unitOfWork   UnitOfWork[T]
	storeAdapter StoreAdapter[S, F]
	errorToFail  ErrorToFail[F]
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
	execErr := w.unitOfWork.Execute(func(repos T) (fnErr error) {
		// Idempotency Check
		var ok bool
		ok, successOutput, fnErr = w.storeAdapter.AlreadyProcessed(ctx, idempotencyKey,
			hash, repos.IdempotencyStore())
		if ok || fnErr != nil {
			return
		}
		// Execute Action
		successOutput, fnErr = action(ctx, repos, idempotencyKey, input)
		if fnErr != nil {
			// Handle Failure: Business or System Error
			isBusinessError, failOutput := w.errorToFail(fnErr)
			if isBusinessError {
				// Business logic failure (e.g., OCC failed, Stock unavailable). Save
				// the fail record.
				if storeErr := w.storeAdapter.SaveFailOutput(ctx, idempotencyKey, hash,
					failOutput, repos.IdempotencyStore()); storeErr != nil {
					fnErr = NewFailureOutputStoreError(storeErr, fnErr)
				} else {
					err = fnErr
					fnErr = nil
				}
			}
			return
		}
		// Action SUCCEEDED. Save the success record.
		if storeErr := w.storeAdapter.SaveSuccessOutput(ctx, idempotencyKey, hash,
			successOutput, repos.IdempotencyStore()); storeErr != nil {
			fnErr = NewSuccessOutputStoreError(storeErr)
		}
		return
	})
	if execErr != nil {
		err = execErr
	}
	return
}
