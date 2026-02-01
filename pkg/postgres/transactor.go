package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type key struct{}

type Transactor interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}

type Manager struct {
	pool *pgxpool.Pool
}

func NewManager(pool *pgxpool.Pool) *Manager {
	return &Manager{pool: pool}
}

func (tm *Manager) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	opts := pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	}

	tx, err := tm.pool.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	ctxWithTx := context.WithValue(ctx, key{}, tx)
	if err := fn(ctxWithTx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

type Executor interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (tm *Manager) GetExecutor(ctx context.Context) Executor {
	if tx, ok := ctx.Value(key{}).(pgx.Tx); ok {
		return tx
	}
	return tm.pool
}
