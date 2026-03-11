package dto

type TransferFailure struct {
	Reason string
}

func (e TransferFailure) Error() string {
	return e.Reason
}

func (e TransferFailure) IdempotentFailure() {}
