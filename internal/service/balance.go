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
	ErrNoBalance = errors.New("balance not exists")
)

// BalanceStorage is user balance storage that acts as cache.
type BalanceStorage interface {
	// HasBalance checks if user balance present in cache.
	HasBalance(ctx context.Context, uid user.ID) (bool, error)

	// GetBalance returns user balance from cache storage.
	//
	// If balance was not cached, ErrNoBalance error will be returned.
	GetBalance(ctx context.Context, uid user.ID) ([]loan.Balance, error)

	// SetBalance implicitly sets user balance in cache storage.
	SetBalance(ctx context.Context, uid user.ID, balance ...loan.Balance) error

	// UpdateBalance updates user balance with specified delta.
	//
	// Method doesn't check if balance value exists, so HasBalance call is required.
	UpdateBalance(ctx context.Context, uid user.ID, deltas ...loan.Balance) error

	// ClearBalance removes balance record from storage.
	ClearBalance(ctx context.Context, uid user.ID) error
}

// LoansStorage is loans log storage.
type LoansStorage interface {
	// AddLoans adds loan records
	AddLoans(ctx context.Context, records []loan.Loan) error

	// GetUserBalance returns balance (saldo) for each user
	// that gave loan to a user or have dept.
	GetUserBalance(ctx context.Context, uid user.ID) ([]loan.Balance, error)
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

// GetUserBalance provides user balance status.
//
// Balance is total summary of all loans and debts.
//
// Method tries to lookup denormalized value in cache.
// If value is not cached or cache is borked, value will be calculated and stored in cache.
func (svc LoanService) GetUserBalance(ctx context.Context, uid user.ID) ([]loan.Balance, error) {
	balance, err := svc.cache.GetBalance(ctx, uid)
	if err == nil {
		// Return denormalized value if present
		svc.log.Debug("serving balance data from cache", zap.Any("uid", uid),
			zap.Any("balance", balance))
		return balance, nil
	}

	if err == ErrNoBalance {
		svc.log.Info("no user balance in cache, populating balance cache", zap.Any("uid", uid))
	} else {
		svc.log.Warn("failed to read cache, repopulating balance", zap.Error(err), zap.Any("uid", uid))
	}

	// User balance is populated to a cache only after first balance status was queried.
	balance, err = svc.loans.GetUserBalance(ctx, uid)
	if err != nil {
		svc.log.Error("failed to calculate user balance", zap.Error(err), zap.Any("uid", uid))
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}

	if err = svc.cache.SetBalance(ctx, uid, balance...); err != nil {
		svc.log.Error("failed to cache user balance", zap.Error(err), zap.Any("uid", uid))
	}
	return balance, nil
}

// AddLoan adds a loan for each debtor from lender by specified amount.
//
// Amount is not divided between debtors, but assigned to each debtor individually.
//
// Implements service.LoanAdder interface.
func (svc LoanService) AddLoan(ctx context.Context, lender user.ID, amount loan.Amount, debtors []user.ID) error {
	// append() to pre-allocated slice is still heavier
	// than index assign, but I want to make things a bit faster.
	// Someone might say that this is a premature optimisation, but anyway
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

	// update balance cache for affected users
	svc.commitBalanceChanges(amount, lender, debtors)
	return nil
}

// commitBalanceChanges updates user balance with specified delta in cache
func (svc LoanService) commitBalanceChanges(delta loan.Amount, lenderID user.ID, users []user.ID) {
	if err := svc.updateUserDebtorsBalance(lenderID, delta, users); err != nil {
		// we got an error for lender only, try luck with debtors...
		svc.log.Error("failed to update user balance in cache",
			zap.Error(err), zap.Any("uid", lenderID), zap.Int64("delta", delta))
	}

	// we're going to reuse this struct for all debtors.
	// For all debtors, balance is going to decrease by n(loan).
	balance := loan.Balance{
		UserID:  lenderID,
		Balance: -delta,
	}

	// update debtors balance status in cache
	for _, uid := range users {
		if err := svc.updateUserBalance(uid, balance); err != nil {
			svc.log.Error("failed to update debtor's balance in cache",
				zap.Error(err), zap.Any("lender_id", lenderID),
				zap.Any("debtor_id", uid), zap.Int64("delta", delta))
			continue
		}
	}

	svc.log.Debug("updated debtors balance in cache", zap.Any("lender_id", lenderID),
		zap.Any("debtors", users), zap.Int64("delta", delta))
}

// updateUserDebtorsBalance updates lender debtors balance registry.
//
// This is how user balance relation is kept in cache:
//
//	var UserBalance = map[user.ID]map[user.ID]loan.Amount
func (svc LoanService) updateUserDebtorsBalance(uid user.ID, delta loan.Amount, debtors []user.ID) error {
	exists, err := svc.cache.HasBalance(svc.rootCtx, uid)
	if err != nil {
		return fmt.Errorf("failed to check lender balance cache status: %w", err)
	}

	if !exists {
		svc.log.Info("lender balance cache not populated, skip update",
			zap.Any("uid", uid), zap.Int64("delta", delta), zap.Any("debtors", debtors))
		return nil
	}

	deltas := make([]loan.Balance, len(debtors))
	for i, debtorID := range debtors {
		deltas[i] = loan.Balance{
			UserID:  debtorID,
			Balance: delta,
		}
	}

	if err := svc.cache.UpdateBalance(svc.rootCtx, uid, deltas...); err != nil {
		// Loaner cache possibly borked, try to truncate it
		_ = svc.cache.ClearBalance(svc.rootCtx, uid)
		return fmt.Errorf("failed to commit updates loaner balance cache: %w", err)
	}
	return nil
}

// updateUserBalance commits user balance change when balance related to one of users is changed.
// For example when user loaned or took dept from other user.
func (svc LoanService) updateUserBalance(uid user.ID, balance loan.Balance) error {
	exists, err := svc.cache.HasBalance(svc.rootCtx, uid)
	if err != nil {
		return fmt.Errorf("failed to check user balance cache status: %w", err)
	}

	if !exists {
		svc.log.Info("user balance cache not populated, skip update",
			zap.Any("uid", uid), zap.Any("balance_uid", balance.UserID),
			zap.Int64("delta", balance.Balance))
		return nil
	}

	return svc.cache.UpdateBalance(svc.rootCtx, uid, balance)
}
