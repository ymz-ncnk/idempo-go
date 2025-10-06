package app

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ymz-ncnk/idempotency-go"
	"github.com/ymz-ncnk/idempotency-go/integration_test/domain"
	"github.com/ymz-ncnk/idempotency-go/integration_test/dto"
	serializer "github.com/ymz-ncnk/idempotency-go/serializer/json"
)

// NewTransferService constructs a TransferService that executes
// transfers atomically and idempotently using the provided UnitOfWork.
func NewTransferService(unitOfWork UnitOfWork) TransferService {
	return TransferService{
		makeTransferWrapper(unitOfWork),
	}
}

// TransferService wraps money transfer logic with idempotency support.
// The wrapper ensures that repeated requests with the same idempotency key
// return the same result without reapplying side effects.
type TransferService struct {
	wrapper idempotency.Wrapper[RepositoryBundle, dto.TransferInput,
		dto.TransferResult, dto.TransferFailure]
}

// Transfer executes a money transfer wrapped in idempotency handling.
// If the same idempotency key and input are seen again, the cached
// result (success or failure) is returned instead of re-executing.
func (s TransferService) Transfer(ctx context.Context, idempotencyKey string,
	input dto.TransferInput,
) (result dto.TransferResult, err error) {
	return s.wrapper.Wrap(ctx, idempotencyKey, input, s.doTransfer)
}

// doTransfer executes a money transfer.
func (s TransferService) doTransfer(ctx context.Context,
	repos RepositoryBundle,
	idempotencyKey string,
	input dto.TransferInput,
) (result dto.TransferResult, err error) {
	from, err := repos.AccountRepo.Get(input.FromAccount)
	if err != nil {
		return
	}
	to, err := repos.AccountRepo.Get(input.ToAccount)
	if err != nil {
		return
	}
	err = domain.Transfer(&from, &to, input.Amount)
	if err != nil {
		return
	}
	err = repos.AccountRepo.Update(from)
	if err != nil {
		return
	}
	err = repos.AccountRepo.Update(to)
	if err != nil {
		return
	}
	result = dto.TransferResult{
		TransactionID: uuid.New().String(),
	}
	return
}

// makeTransferWrapper creates an idempotency.Wrapper specifically configured
// for the fund transfer business logic.
func makeTransferWrapper(
	uow UnitOfWork,
) idempotency.Wrapper[RepositoryBundle, dto.TransferInput, dto.TransferResult, dto.TransferFailure] {
	var (
		// failToError converts the stored failure output (TransferFailure) back
		// into a Go error (ErrInsufficientFunds) for the client on subsequent
		// retries.
		failToError = func(failureOutput dto.TransferFailure) error {
			return domain.ErrInsufficientFunds
		}
		// errorToFail determines which error should be saved as an idempotent
		// failure output.
		errorToFail idempotency.ErrorToFail[dto.TransferFailure] = func(err error) (
			output dto.TransferFailure, ok bool,
		) {
			if errors.Is(err, domain.ErrInsufficientFunds) {
				return dto.TransferFailure{Reason: err.Error()}, true
			}
			// All other errors (e.g., context.DeadlineExceeded, DB errors) are not
			// stored (ok=false),
			return
		}
		manager = idempotency.NewStoreAdapter(
			serializer.JSONSerializer[dto.TransferResult]{},
			serializer.JSONSerializer[dto.TransferFailure]{},
			failToError,
		)
	)
	return idempotency.NewWrapper[RepositoryBundle, dto.TransferInput](
		uow, manager, errorToFail)
}
