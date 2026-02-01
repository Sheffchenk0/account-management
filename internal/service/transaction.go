package service

import (
	"account-manager/internal/entity"
	"account-manager/internal/repo"
	"account-manager/pkg/postgres"
	"context"
	"fmt"
)

type TransactionService struct {
	txMgr *postgres.Manager

	accountRepo     repo.AccountRepo
	transactionRepo repo.TransactionRepo
	outboxRepo      repo.OutboxRepo
}

func NewTransactionService(txMgr *postgres.Manager, ar repo.AccountRepo, tr repo.TransactionRepo, or repo.OutboxRepo) *TransactionService {
	return &TransactionService{
		txMgr: txMgr,

		accountRepo:     ar,
		transactionRepo: tr,
		outboxRepo:      or,
	}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, transaction entity.Transaction) error {
	op := "TransactionService.CreateTransaction"

	return s.txMgr.Run(ctx, func(ctxTX context.Context) error {
		account, err := s.accountRepo.GetByIDForUpdate(ctxTX, transaction.AccountID)
		if err != nil {
			return fmt.Errorf("%s: failde to get account: %w", op, err)
		}

		switch transaction.Type {
		case entity.TransactionTypeCredit:
			if err := account.Credit(transaction.Amount); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		case entity.TransactionTypeDebit:
			if err := account.Debit(transaction.Amount); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		default:
			return fmt.Errorf("%s: %w", op, entity.ErrInvalidTransactionType)
		}

		err = s.accountRepo.UpdateBalance(ctxTX, account.ID, account.Balance)
		if err != nil {
			return fmt.Errorf("%s: failed to update balance: %w", op, err)
		}

		t, err := s.transactionRepo.Create(ctxTX, transaction)
		if err != nil {
			return fmt.Errorf("%s: failed to save transaction: %w", op, err)
		}

		err = s.outboxRepo.Create(ctxTX, "transaction.created", t)
		if err != nil {
			return fmt.Errorf("%s: failed to save event in outbox: %w", op, err)
		}

		return nil
	})
}
