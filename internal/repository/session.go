package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/x1unix/sbda-ledger/internal/model/auth"
	"github.com/x1unix/sbda-ledger/internal/model/user"
	"github.com/x1unix/sbda-ledger/internal/service"
)

type SessionRepository struct {
	redis redis.Cmdable
}

func NewSessionRepository(r redis.Cmdable) *SessionRepository {
	return &SessionRepository{redis: r}
}

// CreateSession implements service.SessionStore
func (r SessionRepository) CreateSession(ctx context.Context, uid user.ID, ttl time.Duration) (*auth.Session, error) {
	sess := auth.NewSession(uid, ttl)
	key := r.redisKeyFromToken(sess.Token())
	data, err := json.Marshal(sess)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	return sess, r.redis.Set(ctx, key, data, sess.TTL).Err()
}

// GetSession implements service.SessionStore
func (r SessionRepository) GetSession(ctx context.Context, token auth.Token) (*auth.Session, error) {
	key := r.redisKeyFromToken(token)
	val, err := r.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, service.ErrSessionNotExists
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	sess := new(auth.Session)
	if err = json.Unmarshal([]byte(val), sess); err != nil {
		return nil, service.ErrCorruptedSession
	}

	return sess, nil
}

// RemoveSession implements service.SessionStore
func (r SessionRepository) RemoveSession(ctx context.Context, token auth.Token) error {
	key := r.redisKeyFromToken(token)
	nAffected, err := r.redis.Del(ctx, key).Result()
	if err == redis.Nil {
		return service.ErrNotExists
	}
	if err != nil {
		return err
	}

	if nAffected == 0 {
		return service.ErrNotExists
	}
	return nil
}

func (_ SessionRepository) redisKeyFromToken(token auth.Token) string {
	return "sess:" + string(token)
}
