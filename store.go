package idempo

import (
	"context"
)

// Store defines the interface for persisting and retrieving idempotency records.
type Store interface {
	// Get retrieves an idempotency Record by its unique ID (idempotencyKey).
	Get(ctx context.Context, id string) (Record, error)
	// Save attempts to persist a new Record.
	Save(ctx context.Context, record Record) error
}
