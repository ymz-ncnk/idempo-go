package memdb

import (
	"context"
	"errors"
	"fmt"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/ymz-ncnk/idempotency-go"
)

// MemDBIdempotencyTableName is the table name for idempotency records.
const MemDBIdempotencyTableName = "idempotency_records"

// NewIdempotencyStore returns a new MemDB idempotency store.
func NewIdempotencyStore(tx *memdb.Txn) idempotency.Store {
	return &IdempotentStore{tx}
}

// IdempotentStore implements the app.IdempotencyStore interface.
type IdempotentStore struct {
	tx *memdb.Txn
}

// Get retrieves an IdempotencyRecord by key.
func (s *IdempotentStore) Get(ctx context.Context, id string) (
	record idempotency.Record, err error,
) {
	raw, err := s.tx.First(MemDBIdempotencyTableName, "id", id)
	if err != nil {
		err = fmt.Errorf(idempotency.ErrorPrefix+"memdb get error: %w", err)
		return
	}
	if raw == nil {
		err = idempotency.ErrIdempotencyRecordNotFound
		return
	}
	record, ok := raw.(idempotency.Record)
	if !ok {
		err = errors.New(idempotency.ErrorPrefix + "memdb internal error: stored value is not IdempotencyRecord")
		return
	}
	return
}

// Save creates a new record.
func (s *IdempotentStore) Save(ctx context.Context,
	record idempotency.Record,
) (err error) {
	if err := s.tx.Insert(MemDBIdempotencyTableName, record); err != nil {
		return fmt.Errorf(idempotency.ErrorPrefix+"memdb insert error: %w", err)
	}
	return
}
