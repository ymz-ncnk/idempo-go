package intest

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	assertfatal "github.com/ymz-ncnk/assert/fatal"
	"github.com/ymz-ncnk/idempotency-go/integration_test/app"
	"github.com/ymz-ncnk/idempotency-go/integration_test/domain"
	"github.com/ymz-ncnk/idempotency-go/integration_test/dto"
	infra "github.com/ymz-ncnk/idempotency-go/integration_test/infrastructure/memdb"
	uow "github.com/ymz-ncnk/idempotency-go/uow/memdb"
)

// TestIdempotency demonstrates how to use the idempotency wrapper
// in a simple money transfer domain.
//
// It shows how retries behave consistently thanks to the idempotency
// guarantees.
func TestIdempotency(t *testing.T) {
	db, err := infra.NewMemDB()
	if err != nil {
		panic(err)
	}
	// Fills the DB with initial account balances for the test scenario.
	fillDB(db)
	service := makeService(db)
	var (
		idempotencyKey = "transfer-123"
		input          = dto.TransferInput{
			FromAccount: "A",
			ToAccount:   "B",
			Amount:      500,
		}
	)

	// First execution succeeds: money moves from A → B.
	t.Run("Should complete transfer successfully", func(t *testing.T) {
		result, err := service.Transfer(context.TODO(), idempotencyKey, input)
		assertfatal.EqualError(err, nil, t)
		assertfatal.Equal(isUUID(result.TransactionID), true, t)

		assertfatal.Equal(getAccount(db, input.FromAccount).Balance, 500, t)
		assertfatal.Equal(getAccount(db, input.ToAccount).Balance, 1500, t)
	})

	// Retrying the same transfer with the same idempotency key
	// returns the cached result without running the action again.
	t.Run("Should return cached result when successfully rerunning same transfer",
		func(t *testing.T) {
			result, err := service.Transfer(context.TODO(), idempotencyKey, input)
			assertfatal.EqualError(err, nil, t)
			assertfatal.Equal(isUUID(result.TransactionID), true, t)

			assertfatal.Equal(getAccount(db, input.FromAccount).Balance, 500, t)
			assertfatal.Equal(getAccount(db, input.ToAccount).Balance, 1500, t)
		})

	// A new transfer request that exceeds available balance fails.
	idempotencyKey = "transfer-456"
	input = dto.TransferInput{
		FromAccount: "A",
		ToAccount:   "B",
		Amount:      1000,
	}
	t.Run("Should fail with insufficient funds error and prevent state change",
		func(t *testing.T) {
			result, err := service.Transfer(context.TODO(), idempotencyKey, input)
			assertfatal.EqualError(err, domain.ErrInsufficientFunds, t)
			assertfatal.Equal(result.TransactionID, "", t)

			assertfatal.Equal(getAccount(db, input.FromAccount).Balance, 500, t)
			assertfatal.Equal(getAccount(db, input.ToAccount).Balance, 1500, t)
		})

	// Transfer in the opposite direction works fine.
	t.Run("Should be able to move fonds back", func(t *testing.T) {
		var (
			idempotencyKey = "transfer-789"
			input          = dto.TransferInput{
				FromAccount: "B",
				ToAccount:   "A",
				Amount:      500,
			}
		)
		result, err := service.Transfer(context.TODO(), idempotencyKey, input)
		assertfatal.EqualError(err, nil, t)
		assertfatal.Equal(isUUID(result.TransactionID), true, t)

		assertfatal.Equal(getAccount(db, input.FromAccount).Balance, 1000, t)
		assertfatal.Equal(getAccount(db, input.ToAccount).Balance, 1000, t)
	})

	// Retrying a failed transfer also returns the cached error instead of
	// executing the action again.
	t.Run("Should return cached error when rerunning failed transfer",
		func(t *testing.T) {
			result, err := service.Transfer(context.TODO(), idempotencyKey, input)
			assertfatal.EqualError(err, domain.ErrInsufficientFunds, t)
			assertfatal.Equal(result.TransactionID, "", t)

			assertfatal.Equal(getAccount(db, input.FromAccount).Balance, 1000, t)
			assertfatal.Equal(getAccount(db, input.ToAccount).Balance, 1000, t)
		})
}

// makeService constructs a TransferService wired with an in-memory UnitOfWork.
//
// The UnitOfWork ensures that both the business action (money transfer)
// and the idempotency record are executed atomically in the same transaction.
func makeService(db *memdb.MemDB) app.TransferService {
	var (
		// The factory creates a new RepositoryBundle for each transaction,
		// containing both the IdempotencyStore and the AccountRepository.
		factory = func(tx *memdb.Txn) app.RepositoryBundle {
			bundle := app.NewRepositoryBundle(uow.NewIdempotencyStore(tx))
			bundle.AccountRepo = infra.NewAccountRepository(tx)
			return bundle
		}
		unitOfWork = uow.NewUnitOfWork(db, factory)
	)
	return app.NewTransferService(unitOfWork)
}

func fillDB(db *memdb.MemDB) {
	tx := db.Txn(true)
	defer tx.Abort()
	repo := infra.NewAccountRepository(tx)
	if err := repo.Add(domain.Account{ID: "A", Balance: 1000}); err != nil {
		panic(err)
	}
	if err := repo.Add(domain.Account{ID: "B", Balance: 1000}); err != nil {
		panic(err)
	}
	tx.Commit()
}

func getAccount(db *memdb.MemDB, id string) domain.Account {
	tx := db.Txn(false)
	defer tx.Abort()
	acc, err := tx.First(infra.AccountsTableName, "id", id)
	if err != nil {
		panic(err)
	}
	return acc.(domain.Account)
}

func isUUID(str string) bool {
	_, err := uuid.Parse(str)
	return err == nil
}
