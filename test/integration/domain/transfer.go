package domain

import (
	"errors"
)

// ErrInsufficientFunds is returned by Transfer when the source account
// does not have enough balance.
var ErrInsufficientFunds = errors.New("insufficient funds")

// Transfer moves the specified amount from one account to another.
// Both accounts are updated in place. Returns ErrInsufficientFunds
// if the source account does not have enough balance.
func Transfer(from *Account, to *Account, amount int64) error {
	if from.Balance < amount {
		return ErrInsufficientFunds
	}
	from.Balance -= amount
	to.Balance += amount
	return nil
}
