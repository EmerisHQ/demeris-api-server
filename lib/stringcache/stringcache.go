package stringcache

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

var (
	ErrCacheMiss = fmt.Errorf("cache miss")
)

//go:generate mockery --name CacheBackend
type CacheBackend interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, expiration time.Duration) error
}

type StringCache struct {
	l              *zap.SugaredLogger
	backend        CacheBackend
	cacheDuration  time.Duration
	storeKeyPrefix string
	handler        Handler
}

//go:generate mockery --name Handler
type Handler interface {
	Handle(ctx context.Context, key string) (string, error)
}

type HandlerFunc func(ctx context.Context, key string) (string, error)

func (h HandlerFunc) Handle(ctx context.Context, key string) (string, error) {
	return h(ctx, key)
}

func NewStringCache(
	logger *zap.SugaredLogger,
	backend CacheBackend,
	cacheDuration time.Duration,
	storeKeyPrefix string,
	handler Handler,
) *StringCache {
	return &StringCache{
		l:              logger,
		backend:        backend,
		cacheDuration:  cacheDuration,
		storeKeyPrefix: storeKeyPrefix,
		handler:        handler,
	}
}

func (c *StringCache) Get(ctx context.Context, key string) (string, error) {
	cacheKey := c.avatarCacheKey(key)

	res, err := c.backend.Get(ctx, cacheKey)
	if err != nil && err != ErrCacheMiss {
		return "", fmt.Errorf("reading cache: %w", err)
	}

	if err == ErrCacheMiss {
		// cache miss, update it
		c.l.Debugw(
			"cache miss",
			"key", key,
		)
		res, err := c.handler.Handle(ctx, key)
		if err != nil {
			setErr := c.backend.Set(ctx, cacheKey, res, c.cacheDuration)
			if setErr != nil {
				c.l.Errorw(
					"updating cache, proceeding anyway",
					"err", err,
				)
			}
		}
		return res, err
	}

	return res, nil
}

func (c *StringCache) avatarCacheKey(validatorId string) string {
	return fmt.Sprintf("%s/%s", c.storeKeyPrefix, validatorId)
}
