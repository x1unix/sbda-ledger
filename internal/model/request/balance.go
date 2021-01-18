package request

import "github.com/x1unix/sbda-ledger/internal/model/loan"

type BalanceStatus struct {
	Status []loan.Balance `json:"status"`
}
