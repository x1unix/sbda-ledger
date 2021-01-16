package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/x1unix/sbda-ledger/internal/model/loan"
)

const (
	tableLoans = "loans"

	colLenderID  = "lender_id"
	colDebtorID  = "debtor_id"
	colAmount    = "amount"
	colCreatedAt = "created_at"
)

// LoansRepository stores loan log in database
type LoansRepository struct {
	db *sqlx.DB
}

// NewLoansRepository is LoansRepository constructor
func NewLoansRepository(db *sqlx.DB) *LoansRepository {
	return &LoansRepository{db: db}
}

// AddLoans implements service.LoansStorage
func (r LoansRepository) AddLoans(ctx context.Context, records []loan.Loan) error {
	q := psql.Insert(tableLoans).Columns(colLenderID, colDebtorID, colAmount)
	for _, record := range records {
		q = q.Values(record.LenderID, record.DebtorID, record.Amount)
	}

	_, err := q.RunWith(r.db).ExecContext(ctx)
	return err
}
