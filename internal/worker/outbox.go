package worker

import (
	"account-manager/internal/repo"
	"account-manager/pkg/postgres"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type EventPublisher interface {
	Publish(ctx context.Context, topic string, payload []byte) error
}

type OutboxProcessor struct {
	txMgr     *postgres.Manager
	repo      repo.OutboxRepo
	publisher EventPublisher
	interval  time.Duration
	batchSize int
}

func NewOutboxProcessor(
	txMgr *postgres.Manager,
	repo repo.OutboxRepo,
	publisher EventPublisher,
	interval time.Duration,
	batchSize int,
) *OutboxProcessor {
	return &OutboxProcessor{
		txMgr,
		repo,
		publisher,
		interval,
		batchSize,
	}
}

func (p *OutboxProcessor) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.processBatch(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (p *OutboxProcessor) processBatch(ctx context.Context) error {
	return p.txMgr.Run(ctx, func(ctxTX context.Context) error {
		events, err := p.repo.GetBatch(ctxTX, p.batchSize)
		if err != nil {
			return fmt.Errorf("get batch: %w", err)
		}

		var successIDs []uuid.UUID
		for _, event := range events {
			pubCtx, cancel := context.WithTimeout(ctxTX, 2*time.Second)
			err := p.publisher.Publish(pubCtx, event.Topic, event.Payload)
			cancel()

			if err != nil {
				// TODO: add logs
				break
			}

			successIDs = append(successIDs, event.ID)
		}

		if len(successIDs) > 0 {
			err := p.repo.DeleteBatch(ctxTX, successIDs)
			if err != nil {
				return fmt.Errorf("delete batch: %w", err)
			}
		}

		return nil
	})
}
