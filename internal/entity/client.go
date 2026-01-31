package entity

import "github.com/google/uuid"

type Client struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	Lastname  string    `db:"lastname"`
	CreatedAt int64     `db:"created_at"`
}

type Account struct {
	ID        uuid.UUID `db:"id"`
	Balance   int64     `db:"balance"`
	ClientID  uuid.UUID `db:"client_id"`
	CreatedAt string    `db:"created_at"`
}

type Transaction struct {
	ID        uuid.UUID `db:"id"`
	AccountID int64     `db:"account_id"`
	Amount    string    `db:"amount"`
	Type      string    `db:"type"`
	CreatedAt string    `db:"created_at"`
}
