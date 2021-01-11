package request

import "github.com/x1unix/sbda-ledger/internal/model/user"

type GroupCreateRequest struct {
	Name string `json:"name" validate:"required,min=3,max=64"`
}

type GroupsResponse struct {
	Groups user.Groups `json:"groups"`
}
