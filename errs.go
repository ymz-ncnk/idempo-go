package idempo

import (
	"errors"
	"fmt"
)

// ErrorPrefix is a common prefix for all idempotency errors.
const ErrorPrefix = "idempotency error: "

var (
	// ErrIdempotencyRecordNotFound is returned when the idempotency record is not
	// found.
	ErrIdempotencyRecordNotFound = errors.New(ErrorPrefix + "record not found")
	// ErrHashMismatch is returned when an attempt is made to reuse an execution
	// ID with input parameters that are different from those used in the
	// original, completed execution.
	// This indicates a misuse of the idempotency key.
	ErrHashMismatch = errors.New(ErrorPrefix + "idempotency key already used with different input data")
)

// NewSuccessOutputMarshalError wraps a low-level marshalling error.
//
// This error is returned by the StoreAdapter when it fails to marshal the
// success output.
func NewSuccessOutputMarshalError(marshalErr error) error {
	return fmt.Errorf(ErrorPrefix+"success output marshal error: %w", marshalErr)
}

// NewFailureOutputMarshalError wraps a low-level marshalling error.
//
// This error is returned by the StoreAdapter when it fails to marshal the
// failure output.
func NewFailureOutputMarshalError(marshalErr error) error {
	return fmt.Errorf(ErrorPrefix+"fail output marshal error: %w", marshalErr)
}

// NewSuccessOutputUnmarshalError wraps a low-level unmarshalling error.
//
// This error is returned by the StoreAdapter when it fails to unmarshal
// the persisted success output.
func NewSuccessOutputUnmarshalError(unmarshalErr error) error {
	return fmt.Errorf(ErrorPrefix+"success output unmarshal error: %w", unmarshalErr)
}

// NewFailureOutputUnmarshalError wraps a low-level unmarshalling error.
//
// This error is returned by the StoreAdapter when it fails to unmarshal
// the persisted failure output.
func NewFailureOutputUnmarshalError(unmarshalErr error) error {
	return fmt.Errorf(ErrorPrefix+"fail output unmarshal error: %w", unmarshalErr)
}

// NewSuccessOutputStoreError constructs a new error instance indicating a
// failure to persist an output of the successful Action.
//
// This error signals a critical system failure within the UnitOfWork after
// the execution of the protected Action.
// It leads to the rollback of the entire UnitOfWork.
func NewSuccessOutputStoreError(storeErr error) *SuccessOutputStoreError {
	return &SuccessOutputStoreError{
		storeErr: storeErr,
	}
}

// SuccessOutputStoreError represents a critical failure to save the output of
// the successful Action.
//
// This is a system-level infrastructure error (e.g., database connection loss)
// and means the Wrapper cannot guarantee idempotency for the completed Action.
// It leads to the rollback of the entire UnitOfWork.
type SuccessOutputStoreError struct {
	// storeErr is the critical error from the idempotency store (system error).
	storeErr error
}

func (e *SuccessOutputStoreError) Error() string {
	return fmt.Sprintf(
		ErrorPrefix+"failed to save success output: %s",
		e.storeErr,
	)
}

func (e *SuccessOutputStoreError) Unwrap() error {
	return e.storeErr
}

// NewFailureOutputStoreError constructs a new error instance indicating a
// failure to persist an output of the failed Action.
//
// This error signals a critical system failure within the UnitOfWork that
// occurred while attempting to save the record of a previously failed Action.
func NewFailureOutputStoreError(storeErr error,
	originalExecErr error,
) *FailureOutputStoreError {
	return &FailureOutputStoreError{
		storeErr:        storeErr,
		originalExecErr: originalExecErr,
	}
}

// FailureOutputStoreError represents a critical failure to save the output
// of the failed Action.
//
// This is a system-level infrastructure error (e.g., database connection loss)
// and means the Wrapper cannot guarantee idempotency for the completed Action.
// It leads to the rollback of the entire UnitOfWork.
type FailureOutputStoreError struct {
	// storeErr is the critical error from the idempotency store (system error).
	storeErr error
	// originalExecErr is the initial error from the action execution.
	originalExecErr error
}

func (e *FailureOutputStoreError) Error() string {
	return fmt.Sprintf(
		ErrorPrefix+"failed to save failure output: %s (original error: %s)",
		e.storeErr,
		e.originalExecErr,
	)
}

func (e *FailureOutputStoreError) Unwrap() error {
	return e.storeErr
}
