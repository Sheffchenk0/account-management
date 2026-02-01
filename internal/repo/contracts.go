package repo

import (
	"account-manager/internal/entity"
	"context"

	"github.com/google/uuid"
)

type (
	AccountRepo interface {
		Create(context.Context, entity.Account) (entity.Account, error)
		GetByIDForUpdate(ctx context.Context, id uuid.UUID) (entity.Account, error)
		UpdateBalance(ctx context.Context, id uuid.UUID, newBalance int64) (entity.Account, error)
	}

	TransactionRepo interface {
		Create(context.Context, entity.Transaction) (entity.Transaction, error)
	}

	OutboxRepo interface {
		Create(ctx context.Context, topic string, payload interface{}) error
	}
)
