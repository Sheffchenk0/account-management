package entity

import (
	"github.com/google/uuid"
)

type Outbox struct {
	ID        uuid.UUID `db:"id"`
	Topic     string    `db:"topic"`
	Payload   []byte    `db:"payload"`
	CreatedAt int64     `db:"created_at"`
}
