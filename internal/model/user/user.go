package user

import (
	"errors"
	"fmt"

	"github.com/jackc/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type ID = pgtype.UUID

// IDToString converts user.ID to string
func IDToString(uid ID) string {
	var out string

	// error occurs only when invalid type passed,
	// so in this case it won't occur
	_ = uid.AssignTo(&out)
	return out
}

type Registration struct {
	Props
	Password string `json:"password" validate:"required,min=6"`
}

type Props struct {
	Email string `json:"email" db:"email" validate:"required,email,max=254"`
	Name  string `json:"name" db:"name" validate:"required,min=3,max=64,name"`
}

type Users = []User

type User struct {
	Props

	// ID is unique user ID
	ID pgtype.UUID `json:"id" db:"id"`

	// PasswordHash contains encrypted password and salt in bcrypt format
	PasswordHash string `json:"-" db:"password"`
}

// SetPassword encrypts and updates user password
func (u *User) SetPassword(newPassword string) error {
	// bcrypt already embeds random salt to hashed pass
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	u.PasswordHash = string(hash)
	return nil
}

// ComparePassword compares password with hashed in user
func (u User) ComparePassword(pwd string) (bool, error) {
	if u.PasswordHash == "" {
		return false, errors.New("origin password not available")
	}

	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(pwd))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
