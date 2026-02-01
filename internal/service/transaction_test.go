package service

import (
	"account-manager/internal/entity"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type mockAccountRepo struct {
	GetByIDForUpdateFn func(ctx context.Context, id uuid.UUID) (entity.Account, error)
	UpdateBalanceFn     func(ctx context.Context, id uuid.UUID, newBalance int64) error
}

func (m *mockAccountRepo) Create(ctx context.Context, account entity.Account) (entity.Account, error) {
	return entity.Account{}, nil
}

func (m *mockAccountRepo) GetByID(ctx context.Context, id uuid.UUID) (entity.Account, error) {
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
	CreateFn func(ctx context.Context, transaction entity.Transaction) (entity.Transaction, error)
}

func (m *mockTransactionRepo) Create(ctx context.Context, transaction entity.Transaction) (entity.Transaction, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, transaction)
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

type mockTransactor struct{}

type mockTransactionService struct {
	accountRepo     *mockAccountRepo
	transactionRepo *mockTransactionRepo
	outboxRepo      *mockOutboxRepo
}

func newMockTransactionService(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) *mockTransactionService {
	return &mockTransactionService{
		accountRepo:     ar,
		transactionRepo: tr,
		outboxRepo:      or,
	}
}

func (s *mockTransactionService) CreateTransaction(ctx context.Context, transaction entity.Transaction) error {
	op := "TransactionService.CreateTransaction"

	account, err := s.accountRepo.GetByIDForUpdate(ctx, transaction.AccountID)
	if err != nil {
		return errors.New(op + ": failde to get account: " + err.Error())
	}

	switch transaction.Type {
	case entity.TransactionTypeCredit:
		if err := account.Credit(transaction.Amount); err != nil {
			return errors.New(op + ": " + err.Error())
		}
	case entity.TransactionTypeDebit:
		if err := account.Debit(transaction.Amount); err != nil {
			return errors.New(op + ": " + err.Error())
		}
	default:
		return errors.New(op + ": " + entity.ErrInvalidTransactionType.Error())
	}

	err = s.accountRepo.UpdateBalance(ctx, account.ID, account.Balance)
	if err != nil {
		return errors.New(op + ": failed to update balance: " + err.Error())
	}

	_, err = s.transactionRepo.Create(ctx, transaction)
	if err != nil {
		return errors.New(op + ": failed to save transaction: " + err.Error())
	}

	err = s.outboxRepo.Create(ctx, "transaction.created", transaction)
	if err != nil {
		return errors.New(op + ": failed to save event in outbox: " + err.Error())
	}

	return nil
}

func TestCreateTransaction_BusinessLogic(t *testing.T) {
	accountID := uuid.New()
	clientID := uuid.New()

	tests := []struct {
		name      string
		amount    int64
		transactionType string
		account   entity.Account
		mockSetup func(*mockAccountRepo, *mockTransactionRepo, *mockOutboxRepo)
		wantErr   bool
		wantErrMsg string
	}{
		{
			name:      "Successful Credit",
			amount:    50,
			transactionType: entity.TransactionTypeCredit,
			account:   entity.Account{ID: accountID, ClientID: clientID, Balance: 100},
			mockSetup: func(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) {
				ar.GetByIDForUpdateFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{ID: accountID, ClientID: clientID, Balance: 100}, nil
				}
				ar.UpdateBalanceFn = func(ctx context.Context, id uuid.UUID, newBalance int64) error {
					if newBalance != 150 {
						t.Errorf("expected balance 150, got %d", newBalance)
					}
					return nil
				}
				tr.CreateFn = func(ctx context.Context, transaction entity.Transaction) (entity.Transaction, error) {
					transaction.ID = uuid.New()
					transaction.CreatedAt = 1234567890
					return transaction, nil
				}
				or.CreateFn = func(ctx context.Context, topic string, payload interface{}) error {
					if topic != "transaction.created" {
						t.Errorf("expected topic 'transaction.created', got '%s'", topic)
					}
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:      "Successful Debit",
			amount:    30,
			transactionType: entity.TransactionTypeDebit,
			account:   entity.Account{ID: accountID, ClientID: clientID, Balance: 100},
			mockSetup: func(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) {
				ar.GetByIDForUpdateFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{ID: accountID, ClientID: clientID, Balance: 100}, nil
				}
				ar.UpdateBalanceFn = func(ctx context.Context, id uuid.UUID, newBalance int64) error {
					if newBalance != 70 {
						t.Errorf("expected balance 70, got %d", newBalance)
					}
					return nil
				}
				tr.CreateFn = func(ctx context.Context, transaction entity.Transaction) (entity.Transaction, error) {
					transaction.ID = uuid.New()
					transaction.CreatedAt = 1234567890
					return transaction, nil
				}
				or.CreateFn = func(ctx context.Context, topic string, payload interface{}) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:      "Insufficient Funds Debit",
			amount:    100,
			transactionType: entity.TransactionTypeDebit,
			account:   entity.Account{ID: accountID, ClientID: clientID, Balance: 50},
			mockSetup: func(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) {
				ar.GetByIDForUpdateFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{ID: accountID, ClientID: clientID, Balance: 50}, nil
				}
			},
			wantErr:   true,
			wantErrMsg: "insufficient funds",
		},
		{
			name:      "Invalid Amount - Zero Credit",
			amount:    0,
			transactionType: entity.TransactionTypeCredit,
			account:   entity.Account{ID: accountID, ClientID: clientID, Balance: 100},
			mockSetup: func(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) {
				ar.GetByIDForUpdateFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{ID: accountID, ClientID: clientID, Balance: 100}, nil
				}
			},
			wantErr:   true,
			wantErrMsg: "amount must be positive",
		},
		{
			name:      "Invalid Amount - Negative Debit",
			amount:    -10,
			transactionType: entity.TransactionTypeDebit,
			account:   entity.Account{ID: accountID, ClientID: clientID, Balance: 100},
			mockSetup: func(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) {
				ar.GetByIDForUpdateFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{ID: accountID, ClientID: clientID, Balance: 100}, nil
				}
			},
			wantErr:   true,
			wantErrMsg: "amount must be positive",
		},
		{
			name:      "Invalid Transaction Type",
			amount:    50,
			transactionType: "transfer",
			account:   entity.Account{ID: accountID, ClientID: clientID, Balance: 100},
			mockSetup: func(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) {
				ar.GetByIDForUpdateFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{ID: accountID, ClientID: clientID, Balance: 100}, nil
				}
			},
			wantErr:   true,
			wantErrMsg: "invalid transaction type",
		},
		{
			name:      "Account Not Found",
			amount:    50,
			transactionType: entity.TransactionTypeCredit,
			account:   entity.Account{ID: accountID, ClientID: clientID, Balance: 100},
			mockSetup: func(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) {
				ar.GetByIDForUpdateFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{}, entity.ErrAccountNotFound
				}
			},
			wantErr:   true,
			wantErrMsg: "account not found",
		},
		{
			name:      "Update Balance Error",
			amount:    50,
			transactionType: entity.TransactionTypeCredit,
			account:   entity.Account{ID: accountID, ClientID: clientID, Balance: 100},
			mockSetup: func(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) {
				ar.GetByIDForUpdateFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{ID: accountID, ClientID: clientID, Balance: 100}, nil
				}
				ar.UpdateBalanceFn = func(ctx context.Context, id uuid.UUID, newBalance int64) error {
					return errors.New("database connection error")
				}
			},
			wantErr:   true,
			wantErrMsg: "failed to update balance",
		},
		{
			name:      "Create Transaction Error",
			amount:    50,
			transactionType: entity.TransactionTypeCredit,
			account:   entity.Account{ID: accountID, ClientID: clientID, Balance: 100},
			mockSetup: func(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) {
				ar.GetByIDForUpdateFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{ID: accountID, ClientID: clientID, Balance: 100}, nil
				}
				ar.UpdateBalanceFn = func(ctx context.Context, id uuid.UUID, newBalance int64) error {
					return nil
				}
				tr.CreateFn = func(ctx context.Context, transaction entity.Transaction) (entity.Transaction, error) {
					return entity.Transaction{}, errors.New("database connection error")
				}
			},
			wantErr:   true,
			wantErrMsg: "failed to save transaction",
		},
		{
			name:      "Outbox Create Error",
			amount:    50,
			transactionType: entity.TransactionTypeCredit,
			account:   entity.Account{ID: accountID, ClientID: clientID, Balance: 100},
			mockSetup: func(ar *mockAccountRepo, tr *mockTransactionRepo, or *mockOutboxRepo) {
				ar.GetByIDForUpdateFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{ID: accountID, ClientID: clientID, Balance: 100}, nil
				}
				ar.UpdateBalanceFn = func(ctx context.Context, id uuid.UUID, newBalance int64) error {
					return nil
				}
				tr.CreateFn = func(ctx context.Context, transaction entity.Transaction) (entity.Transaction, error) {
					transaction.ID = uuid.New()
					transaction.CreatedAt = 1234567890
					return transaction, nil
				}
				or.CreateFn = func(ctx context.Context, topic string, payload interface{}) error {
					return errors.New("rabbitmq connection error")
				}
			},
			wantErr:   true,
			wantErrMsg: "failed to save event in outbox",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAccRepo := &mockAccountRepo{}
			mockTxRepo := &mockTransactionRepo{}
			mockOutboxRepo := &mockOutboxRepo{}

			tt.mockSetup(mockAccRepo, mockTxRepo, mockOutboxRepo)

			svc := newMockTransactionService(mockAccRepo, mockTxRepo, mockOutboxRepo)

			transaction := entity.Transaction{
				AccountID: accountID,
				Amount:    tt.amount,
				Type:      tt.transactionType,
			}

			err := svc.CreateTransaction(context.Background(), transaction)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
					return
				}
				if tt.wantErrMsg != "" {
					errMsg := err.Error()
					found := false
					
					// Проверяем на специфические ошибки entity
					if tt.wantErrMsg == "insufficient funds" && errors.Is(err, entity.ErrInsufficientFunds) {
						found = true
					} else if tt.wantErrMsg == "amount must be positive" && errors.Is(err, entity.ErrInvalidAmount) {
						found = true
					} else if tt.wantErrMsg == "invalid transaction type" && errors.Is(err, entity.ErrInvalidTransactionType) {
						found = true
					} else if len(errMsg) > 0 && containsSubstring(errMsg, tt.wantErrMsg) {
						found = true
					}
					
					if !found {
						t.Errorf("expected error message to contain '%s', got '%v'", tt.wantErrMsg, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
