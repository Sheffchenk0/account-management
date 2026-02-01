package entity

import "errors"

var (
	ErrInsufficientFunds      = errors.New("insufficient funds")
	ErrAccountNotFound        = errors.New("account not found")
	ErrDuplicateTransaction   = errors.New("transaction already exists")
	ErrInvalidAmount          = errors.New("amount must be positive")
	ErrInvalidTransactionType = errors.New("invalid transaction type")
)
