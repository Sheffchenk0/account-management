package entity

import "github.com/google/uuid"

type Account struct {
	ID        uuid.UUID `db:"id"`
	Balance   int64     `db:"balance"`
	ClientID  uuid.UUID `db:"client_id"`
	CreatedAt string    `db:"created_at"`
}

func (a *Account) CanDebit(amount int64) bool {
	return a.Balance >= amount
}
