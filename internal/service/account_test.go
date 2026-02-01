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

func TestAccountService_CreateAccount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     entity.Account
		mockSetup func(*mockAccountRepo)
		wantErr   bool
		checkErr  func(*testing.T, error)
	}{
		{
			name: "successful create",
			input: entity.Account{
				Balance: 1000,
			},
			mockSetup: func(m *mockAccountRepo) {
				m.CreateFn = func(ctx context.Context, acc entity.Account) (entity.Account, error) {
					acc.ID = uuid.New()
					return acc, nil
				}
			},
			wantErr: false,
		},
		{
			name: "repository error",
			input: entity.Account{
				Balance: 1000,
			},
			mockSetup: func(m *mockAccountRepo) {
				m.CreateFn = func(ctx context.Context, acc entity.Account) (entity.Account, error) {
					return entity.Account{}, errors.New("database error")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mockAccountRepo{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			service := NewAccountService(nil, mockRepo)
			acc, err := service.CreateAccount(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.checkErr != nil {
					tt.checkErr(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, acc.ID)
			}
		})
	}
}

func TestAccountService_GetAccount(t *testing.T) {
	t.Parallel()

	accountID := uuid.New()

	tests := []struct {
		name      string
		accountID uuid.UUID
		mockSetup func(*mockAccountRepo)
		wantErr   bool
		checkErr  func(*testing.T, error)
	}{
		{
			name:      "successful get",
			accountID: accountID,
			mockSetup: func(m *mockAccountRepo) {
				m.GetByIDFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{
						ID:      id,
						Balance: 1000,
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name:      "account not found",
			accountID: accountID,
			mockSetup: func(m *mockAccountRepo) {
				m.GetByIDFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{}, entity.ErrAccountNotFound
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, errors.Is(err, entity.ErrAccountNotFound))
			},
		},
		{
			name:      "repository error",
			accountID: accountID,
			mockSetup: func(m *mockAccountRepo) {
				m.GetByIDFn = func(ctx context.Context, id uuid.UUID) (entity.Account, error) {
					return entity.Account{}, errors.New("database error")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mockAccountRepo{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			service := NewAccountService(nil, mockRepo)
			acc, err := service.GetAccount(context.Background(), tt.accountID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.checkErr != nil {
					tt.checkErr(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.accountID, acc.ID)
				assert.Equal(t, int64(1000), acc.Balance)
			}
		})
	}
}
