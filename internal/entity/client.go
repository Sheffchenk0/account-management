package entity

import "github.com/google/uuid"

type Client struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	Lastname  string    `db:"lastname"`
	CreatedAt int64     `db:"created_at"`
}
