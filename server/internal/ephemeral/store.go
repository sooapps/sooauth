package ephemeral

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type memoryItem struct {
	raw       []byte
	expiresAt time.Time
}

type Store struct {
	rdb *redis.Client
	mem sync.Map
}

func New(rdb *redis.Client) *Store {
	return &Store{rdb: rdb}
}

func (s *Store) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if s.rdb != nil {
		return s.rdb.Set(ctx, "ephemeral:"+key, raw, ttl).Err()
	}
	s.mem.Store(key, memoryItem{
		raw:       raw,
		expiresAt: time.Now().Add(ttl),
	})
	return nil
}

func (s *Store) Get(ctx context.Context, key string, dest any) (bool, error) {
	if s.rdb != nil {
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

	val, ok := s.mem.Load(key)
	if !ok {
		return false, nil
	}
	item := val.(memoryItem)
	if time.Now().After(item.expiresAt) {
		s.mem.Delete(key)
		return false, nil
	}
	if err := json.Unmarshal(item.raw, dest); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) Delete(ctx context.Context, key string) error {
	if s.rdb != nil {
		return s.rdb.Del(ctx, "ephemeral:"+key).Err()
	}
	s.mem.Delete(key)
	return nil
}
