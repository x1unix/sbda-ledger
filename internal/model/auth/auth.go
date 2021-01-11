package auth

import "github.com/x1unix/sbda-ledger/internal/model/user"

type Credentials struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Remember bool   `json:"remember"`
}

type LoginResult struct {
	User    *user.User `json:"user"`
	Session *Session   `json:"session"`
}
