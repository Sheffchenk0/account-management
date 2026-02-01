package service

import (
	"account-manager/internal/entity"
	"account-manager/internal/repo"
	"account-manager/pkg/postgres"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type AccountService struct {
	txMgr      *postgres.Manager
	accountRepo repo.AccountRepo
}

func NewAccountService(txMgr *postgres.Manager, accountRepo repo.AccountRepo) *AccountService {
	return &AccountService{
		txMgr:      txMgr,
		accountRepo: accountRepo,
	}
}

func (s *AccountService) CreateAccount(ctx context.Context, account entity.Account) (entity.Account, error) {
	const op = "AccountService.CreateAccount"

	acc, err := s.accountRepo.Create(ctx, account)
	if err != nil {
		return entity.Account{}, fmt.Errorf("%s: %w", op, err)
	}

	return acc, nil
}

func (s *AccountService) GetAccount(ctx context.Context, id uuid.UUID) (entity.Account, error) {
	const op = "AccountService.GetAccount"

	account, err := s.accountRepo.GetByID(ctx, id)
	if err != nil {
		return entity.Account{}, fmt.Errorf("%s: %w", op, err)
	}

	return account, nil
}
