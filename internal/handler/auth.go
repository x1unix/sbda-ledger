package handler

import (
	"net/http"

	"github.com/x1unix/sbda-ledger/internal/model"
	"github.com/x1unix/sbda-ledger/internal/web"
)

type AuthHandler struct {
}

func (h AuthHandler) Auth(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	var payload model.LoginRequest
	if err := web.UnmarshalJSON(r, &payload); err != nil {
		return nil, err
	}

	if err := model.Validate(payload); err != nil {
		return nil, err
	}
	return payload, nil
}
