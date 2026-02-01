package repo

import (
	"account-manager/internal/entity"
	"context"

	"github.com/google/uuid"
)

type (
	AccountRepo interface {
		Create(context.Context, entity.Account) (entity.Account, error)
		GetByID(ctx context.Context, id uuid.UUID) (entity.Account, error)
		GetBalanceById(ctx context.Context, id uuid.UUID) (int64, error)
		GetByIDForUpdate(ctx context.Context, id uuid.UUID) (entity.Account, error)
		UpdateBalance(ctx context.Context, id uuid.UUID) (entity.Account, error)
	}

	TransactionRepo interface {
		Create(context.Context, entity.Account) error
		GetByClientID(ctx context.Context, id uuid.UUID) (entity.Account, error)
	}
)
