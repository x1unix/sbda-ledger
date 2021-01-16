package loan

import "github.com/x1unix/sbda-ledger/internal/model/user"

// Balance is dept balance (saldo) for specific user.
//
// Basically is summary of loans given to specific user and debts of that user.
type Balance struct {
	// UserID is ID of related user
	UserID user.ID `json:"user_id" db:"user_id"`

	// Balance is summary of loans given to specific user and debts of that user.
	Balance Amount `json:"balance" db:"amount"`
}
