package entity

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccount_Debit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		initialBalance int64
		amount         int64
		wantBalance    int64
		wantErr        error
	}{
		{
			name:           "successful debit",
			initialBalance: 1000,
			amount:         500,
			wantBalance:    500,
			wantErr:        nil,
		},
		{
			name:           "insufficient funds",
			initialBalance: 100,
			amount:         200,
			wantBalance:    100,
			wantErr:        ErrInsufficientFunds,
		},
		{
			name:           "insufficient funds exact balance",
			initialBalance: 100,
			amount:         100,
			wantBalance:    100,
			wantErr:        ErrInsufficientFunds,
		},
		{
			name:           "negative amount",
			initialBalance: 1000,
			amount:         -100,
			wantBalance:    1000,
			wantErr:        ErrInvalidAmount,
		},
		{
			name:           "zero amount",
			initialBalance: 1000,
			amount:         0,
			wantBalance:    1000,
			wantErr:        ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			acc := &Account{Balance: tt.initialBalance}
			err := acc.Debit(tt.amount)

			if tt.wantErr != nil {
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantBalance, acc.Balance)
		})
	}
}

func TestAccount_Credit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		initialBalance int64
		amount         int64
		wantBalance    int64
		wantErr        error
	}{
		{
			name:           "successful credit",
			initialBalance: 1000,
			amount:         500,
			wantBalance:    1500,
			wantErr:        nil,
		},
		{
			name:           "credit from zero balance",
			initialBalance: 0,
			amount:         100,
			wantBalance:    100,
			wantErr:        nil,
		},
		{
			name:           "negative amount",
			initialBalance: 1000,
			amount:         -100,
			wantBalance:    1000,
			wantErr:        ErrInvalidAmount,
		},
		{
			name:           "zero amount",
			initialBalance: 1000,
			amount:         0,
			wantBalance:    1000,
			wantErr:        ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			acc := &Account{Balance: tt.initialBalance}
			err := acc.Credit(tt.amount)

			if tt.wantErr != nil {
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantBalance, acc.Balance)
		})
	}
}
