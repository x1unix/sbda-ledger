package request

import "github.com/x1unix/sbda-ledger/internal/model/user"

type GroupCreateRequest struct {
	Name string `json:"name" validate:"required,min=3,max=64"`
}

type UserIDs struct {
	IDs []user.ID `json:"ids" validate:"required,min=1"`
}

type GroupsResponse struct {
	Groups user.Groups `json:"groups"`
}
