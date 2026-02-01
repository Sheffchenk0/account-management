package postgres

import (
	"account-manager/internal/entity"
	"account-manager/pkg/postgres"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type OutboxRepo struct {
	mgr *postgres.Manager
}

func NewOutboxRepo(mgr *postgres.Manager) *OutboxRepo {
	return &OutboxRepo{mgr}
}

func (r *OutboxRepo) SaveEvent(ctx context.Context, topic string, payload interface{}) error {
	const op = "OutboxRepo.SaveEvent"
	executor := r.mgr.GetExecutor(ctx)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("%s: error while marshalling payload: %w", op, err)
	}

	sql := `
		INSERT INTO outbox (topic, payload) VALUES ($1, $2)
	`

	_, err = executor.Exec(ctx, sql, topic, payloadBytes)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *OutboxRepo) GetBatch(ctx context.Context, batchSize int) ([]entity.Outbox, error) {
	const op = "OutboxRepo.GetBatch"
	executor := r.mgr.GetExecutor(ctx)

	sql := `
		SELECT id, topic, payload, created_at
		FROM outbox
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`

	rows, err := executor.Query(ctx, sql, batchSize)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var events []entity.Outbox
	for rows.Next() {
		var event entity.Outbox
		if err := rows.Scan(&event.ID, &event.Topic, &event.Payload, &event.CreatedAt); err != nil {
			return nil, fmt.Errorf("%s: scan error: %w", op, err)
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *OutboxRepo) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	const op = "OutboxRepo.DeleteBatch"
	executor := r.mgr.GetExecutor(ctx)

	sql := `
		DELETE FROM outbox WHERE id = ANY($1)
	`

	ct, err := executor.Exec(ctx, sql, ids)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if ct.RowsAffected() != int64(len(ids)) {
		// TODO: log warn
	}

	return nil
}
