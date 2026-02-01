package entity

import "github.com/google/uuid"

const (
	TransactionTypeDebit  = "debit"
	TransactionTypeCredit = "credit"
)

type Transaction struct {
	ID        uuid.UUID `db:"id"`
	AccountID uuid.UUID `db:"account_id"`
	Amount    int64     `db:"amount"`
	Type      string    `db:"type"`
	CreatedAt int64     `db:"created_at"`
}
