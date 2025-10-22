package app

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ymz-ncnk/idempo-go"
	"github.com/ymz-ncnk/idempo-go/integration_test/domain"
	"github.com/ymz-ncnk/idempo-go/integration_test/dto"
	serializer "github.com/ymz-ncnk/idempo-go/serializer/json"
)

// NewTransferService constructs a TransferService that executes
// transfers atomically and idempotently using the provided UnitOfWork.
func NewTransferService(unitOfWork idempo.UnitOfWork[RepositoryBundle]) TransferService {
	conf := idempo.Config[RepositoryBundle, dto.TransferSuccess, dto.TransferFailure]{
		UnitOfWork: unitOfWork,
		SuccessSer: serializer.JSONSerializer[dto.TransferSuccess]{},
		FailureSer: serializer.JSONSerializer[dto.TransferFailure]{},
		FailureToError: func(failure dto.TransferFailure) error {
			return domain.ErrInsufficientFunds
		},
		ErrorToFailure: func(err error) (ok bool, failure dto.TransferFailure) {
			if errors.Is(err, domain.ErrInsufficientFunds) {
				return true, dto.TransferFailure{Reason: err.Error()}
			}
			// All other errors (e.g., context.DeadlineExceeded, DB errors) are not
			// stored (ok=false),
			return
		},
	}
	return TransferService{
		wrapper: idempo.NewWrapper[RepositoryBundle, dto.TransferInput](conf),
	}
}

// TransferService wraps money transfer logic with idempotency support.
// The wrapper ensures that repeated requests with the same idempotency key
// return the same result without reapplying side effects.
type TransferService struct {
	wrapper idempo.Wrapper[RepositoryBundle, dto.TransferInput,
		dto.TransferSuccess, dto.TransferFailure]
}

// Transfer executes a money transfer wrapped in idempotency handling.
// If the same idempotency key and input are seen again, the cached
// result (success or failure) is returned instead of re-executing.
func (s TransferService) Transfer(ctx context.Context, idempotencyKey string,
	input dto.TransferInput,
) (result dto.TransferSuccess, err error) {
	return s.wrapper.Wrap(ctx, idempotencyKey, input, s.doTransfer)
}

// doTransfer executes a money transfer.
func (s TransferService) doTransfer(ctx context.Context,
	repos RepositoryBundle,
	idempotencyKey string,
	input dto.TransferInput,
) (result dto.TransferSuccess, err error) {
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
	result = dto.TransferSuccess{
		TransactionID: uuid.New().String(),
	}
	return
}
