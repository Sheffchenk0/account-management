package postgres

import (
	"account-manager/internal/entity"
	"account-manager/pkg/postgres"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type AccountRepo struct {
	mgr *postgres.Manager
}

func New(m *postgres.Manager) *AccountRepo {
	return &AccountRepo{m}
}

func (r *AccountRepo) Create(ctx context.Context, account entity.Account) (entity.Account, error) {
	const op = "AccountRepo.Create"
	executor := r.mgr.GetExecutor(ctx)

	sql := `
		INSERT INTO accounts (id, client_id) VALUES ($1, $2)
		RETURNING id, balance, created_at
	`

	err := executor.QueryRow(ctx, sql, account.ID, account.ClientID).Scan(
		&account.ID, &account.Balance, &account.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return entity.Account{}, fmt.Errorf("%s: %w", op, entity.ErrDuplicateTransaction)
		}
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
			return entity.Account{}, fmt.Errorf("%s: %w", op, entity.ErrClientNotFound)
		}
		return entity.Account{}, fmt.Errorf("%s: %w", op, err)
	}

	return account, nil
}

func (r *AccountRepo) GetByID(ctx context.Context, id uuid.UUID) (entity.Account, error) {
	const op = "AccountRepo.GetByID"
	executor := r.mgr.GetExecutor(ctx)

	sql := `
		SELECT id, balance, client_id, created_at FROM accounts
		WHERE id = $1
	`
	var account entity.Account

	err := executor.QueryRow(ctx, sql, id).Scan(
		&account.ID, &account.Balance, &account.ClientID, &account.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Account{}, fmt.Errorf("%s: %w", op, entity.ErrAccountNotFound)
		}
		return entity.Account{}, fmt.Errorf("%s: %w", op, err)
	}

	return account, nil
}

func (r *AccountRepo) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (entity.Account, error) {
	const op = "AccountRepo.GetByIDForUpdate"
	executor := r.mgr.GetExecutor(ctx)

	sql := `
		SELECT id, balance, client_id, created_at FROM accounts
		WHERE id = $1
		FOR UPDATE
	`
	var account entity.Account

	err := executor.QueryRow(ctx, sql, id).Scan(
		&account.ID, &account.Balance, &account.ClientID, &account.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Account{}, fmt.Errorf("%s: %w", op, entity.ErrAccountNotFound)
		}
		return entity.Account{}, fmt.Errorf("%s: %w", op, err)
	}

	return account, nil
}
