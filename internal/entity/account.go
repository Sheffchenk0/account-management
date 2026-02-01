package entity

import (
	"github.com/google/uuid"
)

type Account struct {
	ID        uuid.UUID `db:"id"`
	Balance   int64     `db:"balance"`
	ClientID  uuid.UUID `db:"client_id"`
	CreatedAt string    `db:"created_at"`
}

func (a *Account) Debit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if a.Balance <= amount {
		return ErrInsufficientFunds
	}

	a.Balance -= amount
	return nil
}

func (a *Account) Credit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	a.Balance += amount
	return nil
}
