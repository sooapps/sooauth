package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Limiter struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Limiter {
	return &Limiter{rdb: rdb}
}

func (l *Limiter) Allow(ctx context.Context, scope, id string, limit int, window time.Duration) (bool, error) {
	if l.rdb == nil {
		return true, nil
	}
	key := fmt.Sprintf("rl:%s:%s", scope, id)
	count, err := l.rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		_ = l.rdb.Expire(ctx, key, window).Err()
	}
	return count <= int64(limit), nil
}
