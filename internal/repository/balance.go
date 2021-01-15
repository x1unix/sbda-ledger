package repository

import (
	"context"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/x1unix/sbda-ledger/internal/model/loan"
	"github.com/x1unix/sbda-ledger/internal/model/user"
	"github.com/x1unix/sbda-ledger/internal/service"
)

const keyPrefixBalance = "balance:"

// BalanceRepository keeps user balance in Redis cache.
type BalanceRepository struct {
	redis redis.Cmdable
}

// NewBalanceRepository is BalanceRepository constructor
func NewBalanceRepository(r redis.Cmdable) *BalanceRepository {
	return &BalanceRepository{redis: r}
}

func (r BalanceRepository) HasBalance(ctx context.Context, uid user.ID) (bool, error) {
	key := formatBalanceKey(uid)
	v, err := r.redis.Exists(ctx, key).Result()
	return v > 0, err
}

func (r BalanceRepository) GetBalance(ctx context.Context, uid user.ID) (loan.Amount, error) {
	key := formatBalanceKey(uid)
	v, err := r.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, service.ErrNoBalance
	}
	if err != nil {
		return 0, err
	}

	amount, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, service.ErrInvalidBalance
	}
	return amount, nil
}

func (r BalanceRepository) SetBalance(ctx context.Context, uid user.ID, amount loan.Amount) error {
	key := formatBalanceKey(uid)

	return r.redis.Set(ctx, key, amount, 0).Err()
}

func (r BalanceRepository) UpdateBalance(ctx context.Context, uid user.ID, delta loan.Amount) error {
	key := formatBalanceKey(uid)
	return r.redis.IncrBy(ctx, key, delta).Err()
}

func (r BalanceRepository) ClearBalance(ctx context.Context, uid user.ID) error {
	key := formatBalanceKey(uid)
	return r.redis.Del(ctx, key).Err()
}

func formatBalanceKey(uid user.ID) string {
	return keyPrefixBalance + user.IDToString(uid)
}
