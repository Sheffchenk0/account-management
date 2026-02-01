package entity

import (
	"time"

	"github.com/google/uuid"
)

const (
	TransactionTypeDebit  = "debit"
	TransactionTypeCredit = "credit"
)

type Transaction struct {
	ID        uuid.UUID `db:"id"`
	AccountID uuid.UUID `db:"account_id"`
	Amount    int64     `db:"amount"`
	Type      string    `db:"type"`
	CreatedAt time.Time `db:"created_at"`
}
