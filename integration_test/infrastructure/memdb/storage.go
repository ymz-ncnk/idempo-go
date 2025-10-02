package memdb

import (
	"fmt"

	memdb "github.com/hashicorp/go-memdb"
)

// --- MemDB Setup (Schema) ---

const (
	MemDBAccountsTableName    = "accounts"
	MemDBIdempotencyTableName = "idempotency_records"
)

// NewMemDB creates and initializes the go-memdb database instance.
func NewMemDB() (*memdb.MemDB, error) {
	db, err := memdb.NewMemDB(DBConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create memdb: %w", err)
	}
	return db, nil
}

// DBConfig defines the initial database schema for the entire application.
var DBConfig = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		MemDBAccountsTableName:    AccountSchema,
		MemDBIdempotencyTableName: IdempotencyRecordSchema,
	},
}

// AccountSchema defines the structure and indexes for the Transfer aggregates
// in MemDB.
var AccountSchema = &memdb.TableSchema{
	Name: MemDBAccountsTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
	},
}

// IdempotencyRecordSchema defines the structure and indexes for
// IdempotencyRecords in MemDB.
var IdempotencyRecordSchema = &memdb.TableSchema{
	Name: MemDBIdempotencyTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"}, // Index by the Idempotency Key
		},
	},
}
