package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/x1unix/sbda-ledger/internal/model/auth"
	"github.com/x1unix/sbda-ledger/internal/model/user"
	"github.com/x1unix/sbda-ledger/internal/web"
	"go.uber.org/zap"
)

var (
	ErrSessionNotExists = errors.New("session not exists")
	ErrCorruptedSession = errors.New("corrupted session")

	ErrInvalidCredentials = web.NewErrBadRequest("invalid username or password")
	ErrAuthRequired       = web.NewErrUnauthorized("authorization required")
)

const (
	defaultSessionTTL  = time.Hour * 8
	extendedSessionTTL = (time.Hour * 24) * 7
)

// SessionStore is auth session store
type SessionStore interface {
	// CreateSession creates a new auth session
	CreateSession(ctx context.Context, uid user.ID, ttl time.Duration) (*auth.Session, error)

	// GetSession retrieves session by id.
	//
	// Returns ErrSessionNotExists if session not exists, or ErrCorruptedSession if session is not readable.
	GetSession(ctx context.Context, ssid uuid.UUID) (*auth.Session, error)

	// RemoveSession revokes session by id
	RemoveSession(ctx context.Context, ssid uuid.UUID) error
}

// AuthService is authentication service
type AuthService struct {
	store SessionStore
	users *UsersService
	log   *zap.Logger
}

// NewAuthService is AuthService constructor
func NewAuthService(log *zap.Logger, usersSvc *UsersService, store SessionStore) *AuthService {
	return &AuthService{
		log:   log.Named("service.auth"),
		store: store,
		users: usersSvc,
	}
}

// Authenticate authenticates user with provided credentials and returns user info with session on success.
func (s AuthService) Authenticate(ctx context.Context, creds auth.Credentials) (*auth.LoginResult, error) {
	usr, err := s.users.UserByEmail(ctx, creds.Email)
	if err == ErrNotExists {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	passEqual, err := usr.ComparePassword(creds.Password)
	if err != nil {
		return nil, fmt.Errorf("cannot check password: %w", err)
	}

	if !passEqual {
		return nil, ErrInvalidCredentials
	}

	sess, err := s.CreateSession(ctx, usr.ID, creds.Remember)
	if err != nil {
		return nil, err
	}

	return &auth.LoginResult{
		Token:   sess.Token(),
		User:    usr,
		Session: sess,
	}, nil
}

// CreateSession implicitly creates user session by user ID.
//
// Remember argument adjusts session TTL.
func (s AuthService) CreateSession(ctx context.Context, uid user.ID, remember bool) (*auth.Session, error) {
	ttl := defaultSessionTTL
	if remember {
		ttl = extendedSessionTTL
	}

	sess, err := s.store.CreateSession(ctx, uid, ttl)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	return sess, nil
}

// GetSession retrieves session using provided token.
//
// Returns ErrAuthRequired if session is invalid.
func (s AuthService) GetSession(ctx context.Context, ssid uuid.UUID) (*auth.Session, error) {
	sess, err := s.store.GetSession(ctx, ssid)
	if err != nil {
		switch err {
		case ErrSessionNotExists:
			return nil, ErrAuthRequired
		case ErrCorruptedSession:
			s.dropCorruptedSession(ctx, ssid)
			return nil, ErrAuthRequired
		default:
			return nil, err
		}
	}

	return sess, nil
}

// ForgetSession removes session.
//
// Returns ErrAuthRequired if session not exists.
func (s AuthService) ForgetSession(ctx context.Context, ssid uuid.UUID) error {
	err := s.store.RemoveSession(ctx, ssid)
	if err == ErrSessionNotExists {
		return ErrAuthRequired
	}
	return err
}

func (s AuthService) dropCorruptedSession(ctx context.Context, ssid uuid.UUID) {
	if err := s.store.RemoveSession(ctx, ssid); err != nil {
		s.log.Error("failed to remove corrupted session",
			zap.String("ssid", ssid.String()),
			zap.Error(err))
		return
	}

	s.log.Warn("removed corrupted session", zap.String("ssid", ssid.String()))
}
