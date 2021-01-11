package auth

import (
	"encoding/base64"
	"time"

	"github.com/google/uuid"
	"github.com/x1unix/sbda-ledger/internal/model/user"
)

// Token is user auth token
type Token string

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

// NewSession returns new session
func NewSession(uid user.ID, ttl time.Duration) *Session {
	return &Session{
		ID:       uuid.New(),
		UserID:   uid,
		LoggedAt: time.Now(),
		TTL:      ttl,
	}
}
