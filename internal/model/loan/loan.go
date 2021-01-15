package loan

import (
	"time"

	"github.com/jackc/pgtype"
	"github.com/x1unix/sbda-ledger/internal/model/user"
)

// Amount is amount of cents in balance or loan log
type Amount = int64

// Loan describes loan amount given by lender to debtor.
type Loan struct {
	// LenderID is ID of user which lent money.
	LenderID user.ID `json:"lender_id" db:"lender_id"`

	// DebtorID is ID of user which borrowed money.
	DebtorID user.ID `json:"debtor_id" db:"debtor_id"`

	// Amount is loan amount in cents.
	Amount Amount `json:"amount" db:"amount"`
}

// Record is loan record in log with record ID and creation date.
type Record struct {
	Loan

	// ID is record ID in loan log.
	ID pgtype.UUID `json:"id" db:"id"`

	// CreatedAt is record creation date and time.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
