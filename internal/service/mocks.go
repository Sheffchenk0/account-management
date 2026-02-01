package service

import (
	"account-manager/internal/entity"
	"account-manager/pkg/postgres"
	"context"

	"github.com/google/uuid"
)

type mockAccountRepo struct {
	CreateFn           func(context.Context, entity.Account) (entity.Account, error)
	GetByIDFn          func(ctx context.Context, id uuid.UUID) (entity.Account, error)
	GetByIDForUpdateFn func(ctx context.Context, id uuid.UUID) (entity.Account, error)
	UpdateBalanceFn    func(ctx context.Context, id uuid.UUID, newBalance int64) error
}

func (m *mockAccountRepo) Create(ctx context.Context, acc entity.Account) (entity.Account, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, acc)
	}
	return entity.Account{}, nil
}

func (m *mockAccountRepo) GetByID(ctx context.Context, id uuid.UUID) (entity.Account, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return entity.Account{}, nil
}

func (m *mockAccountRepo) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (entity.Account, error) {
	if m.GetByIDForUpdateFn != nil {
		return m.GetByIDForUpdateFn(ctx, id)
	}
	return entity.Account{}, nil
}

func (m *mockAccountRepo) UpdateBalance(ctx context.Context, id uuid.UUID, newBalance int64) error {
	if m.UpdateBalanceFn != nil {
		return m.UpdateBalanceFn(ctx, id, newBalance)
	}
	return nil
}

type mockTransactionRepo struct {
	CreateFn func(context.Context, entity.Transaction) (entity.Transaction, error)
}

func (m *mockTransactionRepo) Create(ctx context.Context, tx entity.Transaction) (entity.Transaction, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, tx)
	}
	return entity.Transaction{}, nil
}

type mockOutboxRepo struct {
	CreateFn func(ctx context.Context, topic string, payload interface{}) error
}

func (m *mockOutboxRepo) Create(ctx context.Context, topic string, payload interface{}) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, topic, payload)
	}
	return nil
}

func (m *mockOutboxRepo) GetBatch(ctx context.Context, batchSize int) ([]entity.Outbox, error) {
	return nil, nil
}

func (m *mockOutboxRepo) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	return nil
}

type mockTransactor struct {
	RunFn func(context.Context, func(context.Context) error) error
}

func (m *mockTransactor) Run(ctx context.Context, fn func(context.Context) error) error {
	if m.RunFn != nil {
		return m.RunFn(ctx, fn)
	}
	return fn(ctx)
}

func newMockManager(t postgres.Transactor) *postgres.Manager {
	return &postgres.Manager{}
}
