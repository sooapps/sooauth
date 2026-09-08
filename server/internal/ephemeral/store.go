package ephemeral

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Store {
	return &Store{rdb: rdb}
}

func (s *Store) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if s.rdb == nil {
		return fmt.Errorf("redis required for ephemeral state")
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, "ephemeral:"+key, raw, ttl).Err()
}

func (s *Store) Get(ctx context.Context, key string, dest any) (bool, error) {
	if s.rdb == nil {
		return false, fmt.Errorf("redis required for ephemeral state")
	}
	raw, err := s.rdb.Get(ctx, "ephemeral:"+key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) Delete(ctx context.Context, key string) error {
	if s.rdb == nil {
		return nil
	}
	return s.rdb.Del(ctx, "ephemeral:"+key).Err()
}
