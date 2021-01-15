package service

import (
	"context"
	"errors"
	"github.com/x1unix/sbda-ledger/internal/model/loan"
	"github.com/x1unix/sbda-ledger/internal/model/user"
)

var (
	ErrNoBalance      = errors.New("balance not exists")
	ErrInvalidBalance = errors.New("invalid balance value")
)

// BalanceStorage is user balance storage that acts as cache.
type BalanceStorage interface {
	// HasBalance checks if user balance present in cache.
	HasBalance(ctx context.Context, uid user.ID) (bool, error)

	// GetBalance returns user balance from storage.
	//
	// If balance not present, ErrNoBalance error will be returned.
	//
	// If balance value is corrupted, ErrInvalidBalance error will be returned.
	GetBalance(ctx context.Context, uid user.ID) (loan.Amount, error)

	// SetBalance implicitly sets user balance in storage.
	SetBalance(ctx context.Context, uid user.ID, amount loan.Amount) error

	// UpdateBalance updates user balance with specified delta.
	//
	// Method doesn't check if balance value exists, so HasBalance call is required.
	UpdateBalance(ctx context.Context, uid user.ID, delta loan.Amount) error

	// ClearBalance removes balance record from storage.
	ClearBalance(ctx context.Context, uid user.ID) error
}

type BalanceService struct {
}
