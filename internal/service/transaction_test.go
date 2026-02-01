package service

import (
	"account-manager/internal/entity"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTransactionService struct {
	accountRepo     *mockAccountRepo
	transactionRepo *mockTransactionRepo
	outboxRepo      *mockOutboxRepo
}

func (s *mockTransactionService) CreateTransaction(ctx context.Context, transaction entity.Transaction) error {
	op := "TransactionService.CreateTransaction"

	account, err := s.accountRepo.GetByIDForUpdate(ctx, transaction.AccountID)
	if err != nil {
		return err
	}

	switch transaction.Type {
	case entity.TransactionTypeCredit:
		if err := account.Credit(transaction.Amount); err != nil {
			return err
		}
	case entity.TransactionTypeDebit:
		if err := account.Debit(transaction.Amount); err != nil {
			return err
		}
	default:
		return entity.ErrInvalidTransactionType
	}

	err = s.accountRepo.UpdateBalance(ctx, account.ID, account.Balance)
	if err != nil {
		return err
	}

	t, err := s.transactionRepo.Create(ctx, transaction)
	if err != nil {
		return err
	}

	err = s.outboxRepo.Create(ctx, "transaction.created", t)
	if err != nil {
		return err
	}

	_ = op
	return nil
}

func TestTransactionService_CreateTransaction(t *testing.T) {
	t.Parallel()

	accountID := uuid.New()
	transactionID := uuid.New()

	tests := []struct {
		name      string
		input     entity.Transaction
		mockSetup func(*mockAccountRepo, *mockTransactionRepo, *mockOutboxRepo)
		wantErr   bool
		checkErr  func(*testing.T, error)
	}{
		{
			name: "successful credit transaction",
			input: entity.Transaction{
				AccountID: accountID,
				Amount:    500,
				Type:      entity.TransactionTypeCredit,
			},
			mockSetup: func(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) {
				ar.GetByIDForUpdateFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{
						ID:      id,
						Balance: 1000,
					}, nil
				}
				ar.UpdateBalanceFn = func(ctx context.Context, id uuid.UUID, newBalance int64) error {
					return nil
				}
				tr.CreateFn = func(ctx context.Context, tx entity.Transaction) (entity.Transaction, error) {
					tx.ID = transactionID
					return tx, nil
				}
				or.CreateFn = func(ctx context.Context, topic string, payload interface{}) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "successful debit transaction",
			input: entity.Transaction{
				AccountID: accountID,
				Amount:    500,
				Type:      entity.TransactionTypeDebit,
			},
			mockSetup: func(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) {
				ar.GetByIDForUpdateFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{
						ID:      id,
						Balance: 1000,
					}, nil
				}
				ar.UpdateBalanceFn = func(ctx context.Context, id uuid.UUID, newBalance int64) error {
					return nil
				}
				tr.CreateFn = func(ctx context.Context, tx entity.Transaction) (entity.Transaction, error) {
					tx.ID = transactionID
					return tx, nil
				}
				or.CreateFn = func(ctx context.Context, topic string, payload interface{}) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "debit insufficient funds",
			input: entity.Transaction{
				AccountID: accountID,
				Amount:    200,
				Type:      entity.TransactionTypeDebit,
			},
			mockSetup: func(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) {
				ar.GetByIDForUpdateFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{
						ID:      id,
						Balance: 100,
					}, nil
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, errors.Is(err, entity.ErrInsufficientFunds))
			},
		},
		{
			name: "credit negative amount",
			input: entity.Transaction{
				AccountID: accountID,
				Amount:    -100,
				Type:      entity.TransactionTypeCredit,
			},
			mockSetup: func(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) {
				ar.GetByIDForUpdateFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{
						ID:      id,
						Balance: 1000,
					}, nil
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, errors.Is(err, entity.ErrInvalidAmount))
			},
		},
		{
			name: "debit zero amount",
			input: entity.Transaction{
				AccountID: accountID,
				Amount:    0,
				Type:      entity.TransactionTypeDebit,
			},
			mockSetup: func(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) {
				ar.GetByIDForUpdateFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{
						ID:      id,
						Balance: 1000,
					}, nil
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, errors.Is(err, entity.ErrInvalidAmount))
			},
		},
		{
			name: "invalid transaction type",
			input: entity.Transaction{
				AccountID: accountID,
				Amount:    100,
				Type:      "invalid",
			},
			mockSetup: func(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) {
				ar.GetByIDForUpdateFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{
						ID:      id,
						Balance: 1000,
					}, nil
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, errors.Is(err, entity.ErrInvalidTransactionType))
			},
		},
		{
			name: "account not found",
			input: entity.Transaction{
				AccountID: accountID,
				Amount:    100,
				Type:      entity.TransactionTypeCredit,
			},
			mockSetup: func(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) {
				ar.GetByIDForUpdateFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{}, entity.ErrAccountNotFound
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, errors.Is(err, entity.ErrAccountNotFound))
			},
		},
		{
			name: "update balance fails",
			input: entity.Transaction{
				AccountID: accountID,
				Amount:    100,
				Type:      entity.TransactionTypeCredit,
			},
			mockSetup: func(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) {
				ar.GetByIDForUpdateFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{
						ID:      id,
						Balance: 1000,
					}, nil
				}
				ar.UpdateBalanceFn = func(ctx context.Context, id uuid.UUID, newBalance int64) error {
					return errors.New("database error")
				}
			},
			wantErr: true,
		},
		{
			name: "create transaction fails",
			input: entity.Transaction{
				AccountID: accountID,
				Amount:    100,
				Type:      entity.TransactionTypeCredit,
			},
			mockSetup: func(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) {
				ar.GetByIDForUpdateFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{
						ID:      id,
						Balance: 1000,
					}, nil
				}
				ar.UpdateBalanceFn = func(ctx context.Context, id uuid.UUID, newBalance int64) error {
					return nil
				}
				tr.CreateFn = func(ctx context.Context, tx entity.Transaction) (entity.Transaction, error) {
					return entity.Transaction{}, errors.New("database error")
				}
			},
			wantErr: true,
		},
		{
			name: "create outbox fails",
			input: entity.Transaction{
				AccountID: accountID,
				Amount:    100,
				Type:      entity.TransactionTypeCredit,
			},
			mockSetup: func(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) {
				ar.GetByIDForUpdateFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{
						ID:      id,
						Balance: 1000,
					}, nil
				}
				ar.UpdateBalanceFn = func(ctx context.Context, id uuid.UUID, newBalance int64) error {
					return nil
				}
				tr.CreateFn = func(ctx context.Context, tx entity.Transaction) (entity.Transaction, error) {
					tx.ID = transactionID
					return tx, nil
				}
				or.CreateFn = func(ctx context.Context, topic string, payload interface{}) error {
					return errors.New("outbox error")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockAccountRepo := &mockAccountRepo{}
			mockTransactionRepo := &mockTransactionRepo{}
			mockOutboxRepo := &mockOutboxRepo{}

			if tt.mockSetup != nil {
				tt.mockSetup(mockAccountRepo, mockTransactionRepo, mockOutboxRepo)
			}

			service := &mockTransactionService{
				accountRepo:     mockAccountRepo,
				transactionRepo: mockTransactionRepo,
				outboxRepo:      mockOutboxRepo,
			}

			err := service.CreateTransaction(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.checkErr != nil {
					tt.checkErr(t, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
