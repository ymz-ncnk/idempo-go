package memdb

import (
	"context"
	"errors"
	"fmt"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/ymz-ncnk/idempo-go"
)

// MemDBIdempotencyTableName is the table name for idempotency records.
const MemDBIdempotencyTableName = "idempotency_records"

// NewIdempotencyStore returns a new MemDB idempotency store.
func NewIdempotencyStore(tx *memdb.Txn) idempo.Store {
	return &IdempotencyStore{tx}
}

// IdempotencyStore implements the app.IdempotencyStore interface.
type IdempotencyStore struct {
	tx *memdb.Txn
}

// Get retrieves an IdempotencyRecord by key.
func (s *IdempotencyStore) Get(ctx context.Context, id string) (
	record idempo.Record, err error,
) {
	raw, err := s.tx.First(MemDBIdempotencyTableName, "id", id)
	if err != nil {
		err = fmt.Errorf(idempo.ErrorPrefix+"memdb get error: %w", err)
		return
	}
	if raw == nil {
		err = idempo.ErrIdempotencyRecordNotFound
		return
	}
	record, ok := raw.(idempo.Record)
	if !ok {
		err = errors.New(idempo.ErrorPrefix + "memdb internal error: stored value is not IdempotencyRecord")
		return
	}
	return
}

// Save creates a new record.
func (s *IdempotencyStore) Save(ctx context.Context,
	record idempo.Record,
) (err error) {
	if err := s.tx.Insert(MemDBIdempotencyTableName, record); err != nil {
		return fmt.Errorf(idempo.ErrorPrefix+"memdb insert error: %w", err)
	}
	return
}
