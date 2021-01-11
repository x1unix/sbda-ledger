package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/x1unix/sbda-ledger/internal/model/user"
)

const (
	ctxSessionKey = "session"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

// Token is user auth token
type Token string

// ParseToken parses session id from token string
func ParseToken(t string) (uuid.UUID, error) {
	return Token(t).SessionID()
}

// SessionID returns session id from token
func (t Token) SessionID() (uuid.UUID, error) {
	rawid, err := base64.StdEncoding.DecodeString(string(t))
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	ssid, err := uuid.FromBytes(rawid)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	return ssid, nil
}

// Session contains user auth session
type Session struct {
	ID       uuid.UUID     `json:"id"`
	UserID   user.ID       `json:"user_id"`
	LoggedAt time.Time     `json:"logged_at"`
	TTL      time.Duration `json:"ttl"`
}

// Token returns auth token for a session
func (s Session) Token() Token {
	return Token(base64.StdEncoding.EncodeToString(s.ID[:]))
}

// ContextWithSession wraps context with session value
func ContextWithSession(ctx context.Context, sess *Session) context.Context {
	return context.WithValue(ctx, ctxSessionKey, sess)
}

// SessionFromContext returns session from context
func SessionFromContext(ctx context.Context) *Session {
	v := ctx.Value(ctxSessionKey)
	if v == nil {
		return nil
	}

	ss, ok := v.(*Session)
	if !ok {
		return nil
	}
	return ss
}

// NewSession returns new session
func NewSession(uid user.ID, ttl time.Duration) *Session {
	return &Session{
		ID:       uuid.New(),
		UserID:   uid,
		LoggedAt: time.Now(),
		TTL:      ttl,
	}
}
