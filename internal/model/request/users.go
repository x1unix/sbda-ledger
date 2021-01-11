package request

import "github.com/x1unix/sbda-ledger/internal/model/user"

type UserIDs struct {
	IDs []user.ID `json:"ids" validate:"required,min=1"`
}

type UsersList struct {
	Users user.Users `json:"users"`
}
