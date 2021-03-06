package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/x1unix/sbda-ledger/internal/model"
	"github.com/x1unix/sbda-ledger/internal/model/auth"
	"github.com/x1unix/sbda-ledger/internal/model/request"
	"github.com/x1unix/sbda-ledger/internal/service"
)

type UserHandler struct {
	usersSvc *service.UsersService
	loanSvc  *service.LoanService
}

// NewUserHandler is UserHandler constructor
func NewUserHandler(usersSvc *service.UsersService, loanSvc *service.LoanService) *UserHandler {
	return &UserHandler{usersSvc: usersSvc, loanSvc: loanSvc}
}

func (h UserHandler) GetUsersList(r *http.Request) (interface{}, error) {
	list, err := h.usersSvc.GetAll(r.Context())
	if err != nil {
		return nil, err
	}

	return request.UsersList{Users: list}, nil
}

func (h UserHandler) GetByID(r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	gid, err := model.DecodeUUID(vars["userId"])
	if err != nil {
		return nil, err
	}

	return h.usersSvc.UserByID(r.Context(), *gid)
}

func (h UserHandler) GetCurrentUser(r *http.Request) (interface{}, error) {
	ctx := r.Context()
	sess := auth.SessionFromContext(ctx)
	if sess == nil {
		return nil, service.ErrAuthRequired
	}

	return h.usersSvc.UserByID(ctx, sess.UserID)
}

func (h UserHandler) GetBalance(r *http.Request) (interface{}, error) {
	ctx := r.Context()
	sess := auth.SessionFromContext(ctx)
	if sess == nil {
		return nil, service.ErrAuthRequired
	}

	val, err := h.loanSvc.GetUserBalance(r.Context(), sess.UserID)
	if err != nil {
		return nil, err
	}

	return request.BalanceStatus{Status: val}, nil
}
