package handler

import (
	"net/http"

	"github.com/x1unix/sbda-ledger/internal/model/auth"
	"github.com/x1unix/sbda-ledger/internal/model/user"
	"github.com/x1unix/sbda-ledger/internal/service"
)

type AuthHandler struct {
	userService *service.UsersService
	authService *service.AuthService
}

// NewAuthHandler is AuthHandler constructor
func NewAuthHandler(userSvc *service.UsersService, authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{
		userService: userSvc,
		authService: authSvc,
	}
}

func (h AuthHandler) Register(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	var reg user.Registration
	if err := UnmarshalAndValidate(r.Body, &reg); err != nil {
		return nil, err
	}

	ctx := r.Context()
	usr, err := h.userService.AddUser(ctx, reg)
	if err != nil {
		return nil, err
	}

	// perform login after registration and return session info
	sess, err := h.authService.CreateSession(ctx, usr.ID, false)
	if err != nil {
		return nil, err
	}

	return &auth.LoginResult{
		Token:   sess.Token(),
		User:    usr,
		Session: sess,
	}, nil
}

func (h AuthHandler) Login(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	var creds auth.Credentials
	if err := UnmarshalAndValidate(r.Body, &creds); err != nil {
		return nil, err
	}

	return h.authService.Authenticate(r.Context(), creds)
}

func (h AuthHandler) GetSession(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	sess := auth.SessionFromContext(r.Context())
	if sess == nil {
		return nil, service.ErrAuthRequired
	}

	return sess, nil
}

func (h AuthHandler) Logout(w http.ResponseWriter, r *http.Request) error {
	sess := auth.SessionFromContext(r.Context())
	if sess == nil {
		return service.ErrAuthRequired
	}

	if err := h.authService.ForgetSession(r.Context(), sess.ID); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
