package repository

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/x1unix/sbda-ledger/internal/model"
	"github.com/x1unix/sbda-ledger/internal/model/loan"
	"github.com/x1unix/sbda-ledger/internal/model/user"
	"github.com/x1unix/sbda-ledger/internal/service"
	"go.uber.org/zap"
)

const (
	keyPrefixBalance = "balance:"
	keyPrefixCached  = "cached:"
)

// BalanceRepository keeps user balance in Redis cache.
type BalanceRepository struct {
	log   *zap.Logger
	redis redis.Cmdable
}

// NewBalanceRepository is BalanceRepository constructor
func NewBalanceRepository(log *zap.Logger, r redis.Cmdable) *BalanceRepository {
	return &BalanceRepository{redis: r, log: log.Named("balance_cache")}
}

// HasBalance implements service.BalanceStore
func (r BalanceRepository) HasBalance(ctx context.Context, uid user.ID) (bool, error) {
	key := formatBalanceKey(uid)
	v, err := r.redis.Exists(ctx, key).Result()
	return v > 0, err
}

// GetBalance implements service.BalanceStore
func (r BalanceRepository) GetBalance(ctx context.Context, uid user.ID) ([]loan.Balance, error) {
	key := formatBalanceKey(uid)
	items, err := r.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("HGetAll: %w", err)
	}

	if len(items) > 0 {
		balance, err := allResultToBalance(items)
		if err != nil {
			// User cache is corrupted and should be truncated.
			// Clean cache and ask service to repopulate it.
			r.mustClearBalance(ctx, uid)
			return nil, service.ErrNoBalance
		}
		return balance, nil
	}

	// if map is empty, check if cache flag is set.
	v, err := r.redis.Exists(ctx, formatCachedKey(uid)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to check if cache flag exists: %w", err)
	}

	if v == 0 {
		// Return ErrNoBalance it user cache mark not present.
		// This used to determine if user balance was never cached
		// or user just has no any balance data for not.
		return nil, service.ErrNoBalance
	}
	return nil, nil
}

// SetBalance implements service.BalanceStore
func (r BalanceRepository) SetBalance(ctx context.Context, uid user.ID, balance ...loan.Balance) error {
	key := formatBalanceKey(uid)
	if err := r.setUserCacheFlag(ctx, uid, true); err != nil {
		return fmt.Errorf("failed to set cache flag: %w", err)
	}

	kvargs := make([]interface{}, 0, len(balance)*2)
	for _, v := range balance {
		kvargs = append(kvargs, user.IDToString(v.UserID), v.Balance)
	}

	if err := r.redis.HMSet(ctx, key, kvargs).Err(); err != nil {
		r.mustClearBalance(ctx, uid)
		return fmt.Errorf("failed to set user balance: %w", err)
	}

	return nil
}

func (r BalanceRepository) setUserCacheFlag(ctx context.Context, uid user.ID, val bool) error {
	key := formatCachedKey(uid)
	if val {
		if err := r.redis.Set(ctx, key, true, 0).Err(); err != nil {
			return fmt.Errorf("failed to set user cache flag: %w", err)
		}
		r.log.Debug("marked user balance as cached", zap.Any("uid", uid))
		return nil
	}

	if err := r.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete user cache flag: %w", err)
	}

	r.log.Debug("removed user cache flag", zap.Any("uid", uid))
	return nil
}

// UpdateBalance implements service.BalanceStore
func (r BalanceRepository) UpdateBalance(ctx context.Context, uid user.ID, deltas ...loan.Balance) error {
	key := formatBalanceKey(uid)
	pipe := r.redis.Pipeline()

	for _, delta := range deltas {
		pipe.HIncrBy(ctx, key, user.IDToString(delta.UserID), delta.Balance)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to commit balance delta for user %q to cache: %w", uid, err)
	}

	return nil
}

// ClearBalance implements service.BalanceStore
func (r BalanceRepository) ClearBalance(ctx context.Context, uid user.ID) error {
	if err := r.redis.Del(ctx, formatCachedKey(uid), formatBalanceKey(uid)).Err(); err != nil {
		return fmt.Errorf("failed to clear user balance from cache: %w", err)
	}
	return nil
}

func (r BalanceRepository) mustClearBalance(ctx context.Context, uid user.ID) {
	if err := r.ClearBalance(ctx, uid); err != nil {
		r.log.Error("failed to drop user balance cache", zap.Error(err), zap.Any("uid", uid))
		return
	}
	r.log.Info("dropped user balance cache", zap.Any("uid", uid))
}

func formatBalanceKey(uid user.ID) string {
	return keyPrefixBalance + user.IDToString(uid)
}

func formatCachedKey(uid user.ID) string {
	return keyPrefixCached + user.IDToString(uid)
}

func allResultToBalance(items map[string]string) ([]loan.Balance, error) {
	out := make([]loan.Balance, 0, len(items))
	for actorID, balance := range items {
		id, err := model.DecodeUUID(actorID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse actor ID %q: %w", id, err)
		}

		balanceVal, err := strconv.ParseInt(balance, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse balance %q for actor ID %q: %w", balance, actorID, err)
		}

		out = append(out, loan.Balance{
			UserID:  *id,
			Balance: balanceVal,
		})
	}

	return out, nil
}
