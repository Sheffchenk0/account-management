package postgres

import (
	"account-manager/internal/entity"
	"account-manager/pkg/postgres"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type TransactionRepo struct {
	txMgr *postgres.Manager
}

func NewTransactionRepo(txMgr *postgres.Manager) *TransactionRepo {
	return &TransactionRepo{txMgr}
}

func (r *TransactionRepo) Create(ctx context.Context, transaction entity.Transaction) (entity.Transaction, error) {
	op := "TransactionRepo.Create"
	executor := r.txMgr.GetExecutor(ctx)

	sql := `
		INSERT INTO transactions(id, account_id, amount, type)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`

	row := executor.QueryRow(ctx, sql, transaction.ID, transaction.AccountID, transaction.Amount, transaction.Type)
	err := row.Scan(&transaction.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return entity.Transaction{}, fmt.Errorf("%s: %w", op, entity.ErrDuplicateTransaction)
		}

		return entity.Transaction{}, fmt.Errorf("%s: %w", op, err)
	}

	return transaction, nil
}
