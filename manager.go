package idempotency

import (
	"context"
)

// FailToError defines the function that converts a stored failure output ('F')
// back into a Go 'error' object.
// Used by the Manager during AlreadyProcessed to recreate the original error.
type FailToError[F any] func(faildOutput F) error

// NewManager creates a new instance of the Manager, initializing it with
// the necessary serializers and the function required to reconstruct a
// stored failure object back into an active Go error.
func NewManager[S, F any](successSer Serializer[S], failSer Serializer[F],
	failToError FailToError[F],
) Manager[S, F] {
	return manager[S, F]{
		successSer:  successSer,
		failSer:     failSer,
		failToError: failToError,
	}
}

// Manager is the core component responsible for interacting with the idempotency
// store (Store). It handles the serialization and deserialization of the
// operation's success output (S) and failure output (F), and converts stored
// failure data back into an application error.
type Manager[S, F any] interface {
	// AlreadyProcessed checks the Store for a record associated with the given
	// idempotency key.
	//
	// If a record is found (ok=true):
	//  1. It reconstructs the original result (either successOutput or an error).
	//  2. If the record is a success, it deserializes and returns the successOutput.
	//  3. If the record is a failure, it deserializes the failure output (F) and
	//     uses the internal failToError function to return the original error.
	//
	// Returns (false, nil, nil) if no record is found.
	AlreadyProcessed(ctx context.Context, idempotencyKey string, inputHash string,
		store Store) (ok bool, successOutput S, err error)
	// SaveSuccessOutput serializes the successful output (S) and persists it
	// to the Store. The inputHash is included to detect non-idempotent re-attempts.
	SaveSuccessOutput(ctx context.Context, idempotencyKey, inputHash string,
		successOutput S, store Store) (err error)
	// SaveFailOutput serializes the failure output (F) and persists it to the
	// Store. This allows the client to receive the same failure error upon retry.
	SaveFailOutput(ctx context.Context, idempotencyKey, inputHash string,
		failOutput F, store Store) (err error)
}

type manager[S, F any] struct {
	successSer  Serializer[S]
	failSer     Serializer[F]
	failToError func(faildOutput F) error
}

func (m manager[S, F]) AlreadyProcessed(ctx context.Context,
	idempotencyKey string,
	inputHash string,
	store Store,
) (ok bool, successOutput S, err error) {
	record, err := store.Get(ctx, idempotencyKey)
	if err != nil {
		if err == ErrIdempotencyRecordNotFound {
			err = nil
		}
		return
	}
	if record.InputHash != inputHash {
		err = ErrHashMismatch
		return
	}
	ok = true
	if record.SuccessOutput {
		successOutput, err = m.successSer.Unmarshal(record.Output)
		if err != nil {
			err = NewSuccessOutputUnmarshalError(err)
		}
		return
	}
	failOutput, err := m.failSer.Unmarshal(record.Output)
	if err != nil {
		err = NewFailureOutputUnmarshalError(err)
		return
	}
	err = m.failToError(failOutput)
	return
}

func (m manager[S, F]) SaveSuccessOutput(ctx context.Context,
	idempotencyKey, inputHash string,
	successOutput S,
	store Store,
) (err error) {
	output, err := m.successSer.Marshal(successOutput)
	if err != nil {
		err = NewSuccessOutputMarshalError(err)
		return
	}
	record := Record{
		ID:            idempotencyKey,
		InputHash:     inputHash,
		SuccessOutput: true,
		Output:        output,
	}
	return store.Save(ctx, record)
}

func (m manager[S, F]) SaveFailOutput(ctx context.Context,
	idempotencyKey, inputHash string,
	failOutput F,
	store Store,
) (err error) {
	output, err := m.failSer.Marshal(failOutput)
	if err != nil {
		// TODO
		err = NewFailureOutputMarshalError(err)
		return
	}
	record := Record{
		ID:            idempotencyKey,
		InputHash:     inputHash,
		SuccessOutput: false,
		Output:        output,
	}
	return store.Save(ctx, record)
}
