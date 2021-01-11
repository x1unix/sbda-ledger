package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/x1unix/sbda-ledger/internal/model/auth"
	"github.com/x1unix/sbda-ledger/internal/model/user"
	"github.com/x1unix/sbda-ledger/internal/web"
	"go.uber.org/zap"
)

var (
	ErrSessionNotExists = errors.New("session not exists")
	ErrCorruptedSession = errors.New("corrupted session")

	errInvalidCredentials = web.NewBadRequestError("invalid username or password")
)

const (
	defaultSessionTTL  = time.Hour * 8
	extendedSessionTTL = (time.Hour * 24) * 7
)

// SessionStore is auth session store
type SessionStore interface {
	// CreateSession creates a new auth session
	CreateSession(ctx context.Context, uid user.ID, ttl time.Duration) (*auth.Session, error)

	// GetSession retrieves session by auth token.
	//
	// Returns ErrSessionNotExists if session not exists, or ErrCorruptedSession if session is not readable.
	GetSession(ctx context.Context, token auth.Token) (*auth.Session, error)

	// RemoveSession revokes session by token
	RemoveSession(ctx context.Context, token auth.Token) error
}

// AuthService is authentication service
type AuthService struct {
	store SessionStore
	users UsersService
	log   *zap.Logger
}

// NewAuthService is AuthService constructor
func NewAuthService(log *zap.Logger, store SessionStore) *AuthService {
	return &AuthService{
		log:   log,
		store: store,
	}
}

// Login authenticates user with provided credentials and returns user info with session on success.
func (s AuthService) Login(ctx context.Context, creds auth.Credentials) (*auth.LoginResult, error) {
	usr, err := s.users.UserByEmail(ctx, creds.Email)
	if err == ErrNotExists {
		return nil, errInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	passEqual, err := usr.ComparePassword(creds.Password)
	if err != nil {
		return nil, fmt.Errorf("cannot check password: %w", err)
	}

	if !passEqual {
		return nil, errInvalidCredentials
	}

	sess, err := s.CreateSession(ctx, usr.ID, creds.Remember)
	if err != nil {
		return nil, err
	}

	return &auth.LoginResult{
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
