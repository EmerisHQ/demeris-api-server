package stringcache

import (
	"context"
	"fmt"
	"time"

	"github.com/emerishq/emeris-utils/store"
	"github.com/go-redis/redis/v8"
)

type StoreBackend struct {
	*store.Store
}

func NewStoreBackend(s *store.Store) *StoreBackend {
	return &StoreBackend{Store: s}
}

func (s *StoreBackend) Get(ctx context.Context, key string) (string, error) {
	cmd := s.Client.Get(ctx, key)
	err := cmd.Err()

	if err == redis.Nil {
		return "", ErrCacheMiss
	} else if err != nil {
		return "", fmt.Errorf("reading store: %w", err)
	}

	return cmd.Val(), nil
}

func (s *StoreBackend) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	return s.Client.Set(ctx, key, value, expiration).Err()
}
