package entity

import "github.com/google/uuid"

const (
	TransactionTypeDebit  = "debit"
	TransactionTypeCredit = "credit"
)

type Transaction struct {
	ID        uuid.UUID `db:"id"`
	AccountID int64     `db:"account_id"`
	Amount    string    `db:"amount"`
	Type      string    `db:"type"`
	CreatedAt string    `db:"created_at"`
}
