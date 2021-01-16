package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/x1unix/sbda-ledger/internal/model/loan"
	"github.com/x1unix/sbda-ledger/internal/model/user"
	"go.uber.org/zap"
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

// LoansStorage is loans log storage.
type LoansStorage interface {
	// AddLoans adds loan records
	AddLoans(ctx context.Context, records []loan.Loan) error

	// UserBalance calculates user balance using transaction log
	UserBalance(ctx context.Context, uid user.ID) (loan.Amount, error)
}

// LoanService manages user dept balance and transactions history
type LoanService struct {
	log     *zap.Logger
	rootCtx context.Context
	cache   BalanceStorage
	loans   LoansStorage
}

// NewLoanService is LoanService constructor.
func NewLoanService(ctx context.Context, log *zap.Logger, cache BalanceStorage, loans LoansStorage) *LoanService {
	return &LoanService{rootCtx: ctx, log: log.Named("service.loans"), cache: cache, loans: loans}
}

// AddLoan adds a loan for each debtor from lender by specified amount.
//
// Amount is not divided between debtors, but assigned to each debtor individually.
//
// Implements service.LoanAdder interface.
func (svc LoanService) AddLoan(ctx context.Context, lender user.ID, amount loan.Amount, debtors []user.ID) error {
	// append() to pre-allocated slice is still heavier
	// than index assign, but I want to make things a bit faster.
	// Someone might say that premature optimisation, but anyway
	// I don't see a purpose for append() here.
	transactions := make([]loan.Loan, len(debtors))
	for i, debtor := range debtors {
		transactions[i] = loan.Loan{
			LenderID: lender,
			DebtorID: debtor,
			Amount:   amount,
		}
	}

	// log all transactions
	if err := svc.loans.AddLoans(ctx, transactions); err != nil {
		return fmt.Errorf("failed to save loan transactions: %w", err)
	}

	// update balance cache for affected users in background
	go svc.commitBalanceChanges(-amount, debtors)
	return nil
}

// commitBalanceChanges updates user balance with specified delta in cache
func (svc LoanService) commitBalanceChanges(delta loan.Amount, users []user.ID) {
	for _, uid := range users {
		if err := svc.updateUserBalance(uid, delta); err != nil {
			svc.log.Error("failed to update user balance in cache",
				zap.Error(err), zap.Any("uid", uid), zap.Int64("delta", delta))
			continue
		}

		svc.log.Debug("updated user balance in cache",
			zap.Any("uid", uid), zap.Int64("delta", delta))
	}
}

func (svc LoanService) updateUserBalance(uid user.ID, delta loan.Amount) error {
	exists, err := svc.cache.HasBalance(svc.rootCtx, uid)
	if err != nil {
		return fmt.Errorf("failed to check user balance cache status: %w", err)
	}

	if !exists {
		svc.log.Info("user balance cache not populated, skip update",
			zap.Any("uid", uid), zap.Int64("delta", delta))
		return nil
	}

	return svc.cache.UpdateBalance(svc.rootCtx, uid, delta)
}
