package service

import (
	"context"
	"fmt"

	"github.com/x1unix/sbda-ledger/internal/model"
	"github.com/x1unix/sbda-ledger/internal/web"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotExists = web.NewBadRequestError("record not found")
	ErrExists    = web.NewBadRequestError("record already exists")
)

type UsersStorage interface {
	AddUser(ctx context.Context, u model.User) error
	Exists(email string) (bool, error)
}

type UsersService struct {
	log   *zap.Logger
	store UsersStorage
}

func NewUsersService(log *zap.Logger, store UsersStorage) *UsersService {
	return &UsersService{
		log:   log.Named("users"),
		store: store,
	}
}

// AddUser registers a new user
func (s UsersService) AddUser(ctx context.Context, usrProps model.UserProps, passwd string) (err error) {
	// TODO: should I do this in single isolated transaction?
	exists, err := s.store.Exists(usrProps.Email)
	if err != nil {
		return fmt.Errorf("can't check if user exists: %w", err)
	}

	if exists {
		return ErrExists
	}

	usr := model.User{UserProps: usrProps}
	usr.Password, err = s.EncryptPassword(passwd)
	if err != nil {
		return err
	}

	if err = s.store.AddUser(ctx, usr); err != nil {
		return fmt.Errorf("failed to create new user %q: %w", usrProps.Email, err)
	}
	return nil
}

// EncryptPassword encrypts password string
func (_ UsersService) EncryptPassword(passwd string) (encryptedPwd string, err error) {
	// bcrypt already embeds random salt to hashed pass
	hash, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), err
}
