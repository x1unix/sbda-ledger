package service

import (
	"context"
	"fmt"

	"github.com/x1unix/sbda-ledger/internal/model"
	"github.com/x1unix/sbda-ledger/internal/model/user"
	"github.com/x1unix/sbda-ledger/internal/web"
	"go.uber.org/zap"
)

var (
	ErrNotExists = web.NewBadRequestError("record not found")
	ErrExists    = web.NewBadRequestError("record already exists")
)

// UserStorage provides user storage
type UserStorage interface {
	// AddUser adds a new user to a storage
	AddUser(ctx context.Context, u user.User) error

	// UserByEmail finds user by email
	UserByEmail(email string) (*user.User, error)

	// Exists checks if user with specified email exists
	Exists(email string) (bool, error)
}

type UsersService struct {
	log   *zap.Logger
	store UserStorage
}

func NewUsersService(log *zap.Logger, store UserStorage) *UsersService {
	return &UsersService{
		log:   log.Named("users"),
		store: store,
	}
}

// UserByEmail finds user by email
func (s UsersService) UserByEmail(ctx context.Context, email string) (*user.User, error) {
	usr, err := s.UserByEmail(ctx, email)
	if err == ErrNotExists {
		return nil, err
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return usr, nil
}

// AddUser registers a new user
func (s UsersService) AddUser(ctx context.Context, usrReg user.Registration) (err error) {
	if err = model.Validate(usrReg); err != nil {
		return err
	}

	// TODO: should I do this in single isolated transaction?
	exists, err := s.store.Exists(usrReg.Email)
	if err != nil {
		return fmt.Errorf("can't check if user exists: %w", err)
	}

	if exists {
		return ErrExists
	}

	usr := user.User{Props: usrReg.Props}
	if err = usr.SetPassword(usrReg.Password); err != nil {
		return err
	}

	if err = s.store.AddUser(ctx, usr); err != nil {
		return fmt.Errorf("failed to create new user %q: %w", usrReg.Email, err)
	}
	return nil
}
