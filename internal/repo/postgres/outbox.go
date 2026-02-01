package postgres

import (
	"account-manager/pkg/postgres"
	"context"
	"encoding/json"
	"fmt"
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
