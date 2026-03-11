package dto

import "fmt"

type TransferInput struct {
	FromAccount string
	ToAccount   string
	Amount      int64
}

func (in TransferInput) Hash() (string, error) {
	return fmt.Sprintf("%s:%s:%d", in.FromAccount, in.ToAccount, in.Amount), nil
}
