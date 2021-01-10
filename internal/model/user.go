package model

import (
	"github.com/jackc/pgtype"
)

type UserProps struct {
	Email string `json:"email" db:"email" validate:"required,email"`
	Name  string `json:"name" db:"name" validate:"required,min=3,max=64,name"`
}

type User struct {
	UserProps
	ID       pgtype.UUID `json:"id" db:"id"`
	Password string      `json:"-" db:"password"`
}
