package handler

import (
	"net/http"

	"github.com/x1unix/sbda-ledger/internal/web"
)

type AuthHandler struct {
}

func (h AuthHandler) Auth(rw http.ResponseWriter, req *http.Request) (interface{}, error) {
	return nil, web.NewBadRequestError("fuck")
}
