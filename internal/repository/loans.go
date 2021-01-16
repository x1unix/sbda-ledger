package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/x1unix/sbda-ledger/internal/model/loan"
	"github.com/x1unix/sbda-ledger/internal/model/user"
)

const (
	tableLoans = "loans"

	colLenderID = "lender_id"
	colDebtorID = "debtor_id"
	colAmount   = "amount"
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

// GetUserBalance implements service.LoansStorage
func (r LoansRepository) GetUserBalance(ctx context.Context, uid user.ID) ([]loan.Balance, error) {
	var out []loan.Balance
	const query = "SELECT user_id, SUM(amount) FROM (" +
		"SELECT debtor_id AS user_id, amount FROM loans WHERE lender_id = $1" +
		" UNION ALL " +
		"SELECT lender_id AS user_id, amount * -1 FROM loans WHERE debtor_id = $1" +
		") balance GROUP BY user_id"
	err := r.db.SelectContext(ctx, &out, query, uid)
	return out, err
}

//
//// GetUserBalance implements service.LoansStorage
//func (r LoansRepository) GetUserBalance(ctx context.Context, uid user.ID) (out loan.Amount, err error) {
//	// TODO: find a better way using LEFT JOIN
//	const query = "SELECT SUM(amount) as balance FROM ((" +
//		"SELECT amount FROM loans WHERE lender_id = $1" +
//		") UNION ALL (" +
//		"SELECT amount * -1 AS amount FROM loans WHERE debtor_id = $1" +
//		")) AS balance"
//
//	err = r.db.GetContext(ctx, &out, query, uid)
//	return out, err
//}
