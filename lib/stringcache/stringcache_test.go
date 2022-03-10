package stringcache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/emerishq/demeris-api-server/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

var l = zap.New(nil).Sugar()

func TestStringCache_CacheMiss(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	// backend doesn't contain key
	backend := new(mocks.CacheBackend)
	backend.EXPECT().Get(mock.Anything, "pref/key12345").Return("", ErrCacheMiss)
	backend.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	defer backend.AssertExpectations(t)

	// handler return a value
	handler := new(mocks.Handler)
	handler.EXPECT().Handle(ctx, "key12345").Return("result", nil)
	defer handler.AssertExpectations(t)

	var cache = NewStringCache(
		l,
		backend,
		time.Microsecond,
		"pref",
		handler,
	)

	res, err := cache.Get(ctx, "key12345")
	assert.NoError(err)
	assert.Equal("result", res)
}

func TestStringCache_UseCache(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	// backend contains key
	backend := new(mocks.CacheBackend)
	res := "result"
	backend.EXPECT().Get(mock.Anything, "pref/key12345").Return(res, nil)
	defer backend.AssertExpectations(t)

	handler := new(mocks.Handler)
	defer handler.AssertExpectations(t)

	var cache = NewStringCache(
		l,
		backend,
		time.Microsecond,
		"pref",
		handler,
	)

	res, err := cache.Get(ctx, "key12345")
	assert.NoError(err)
	assert.Equal("result", res)
}

func TestStringCache_HandlerErrorSetsCache(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	// backend contains key
	backend := new(mocks.CacheBackend)
	backend.EXPECT().Get(mock.Anything, "pref/key12345").Return("", ErrCacheMiss)
	// assert that Set() is called anyway, even if the handler returns an error
	backend.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	defer backend.AssertExpectations(t)

	handler := new(mocks.Handler)
	handler.EXPECT().Handle(ctx, "key12345").Return("", fmt.Errorf("error from handler"))
	defer handler.AssertExpectations(t)

	var cache = NewStringCache(
		l,
		backend,
		time.Microsecond,
		"pref",
		handler,
	)

	_, err := cache.Get(ctx, "key12345")
	assert.Error(err)
}

func TestStringCache_CacheSetError(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	backend := new(mocks.CacheBackend)
	backend.EXPECT().Get(mock.Anything, "pref/key12345").Return("", ErrCacheMiss)
	backend.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error setting cache"))
	defer backend.AssertExpectations(t)

	handler := new(mocks.Handler)
	handler.EXPECT().Handle(ctx, "key12345").Return("result", nil)
	defer handler.AssertExpectations(t)

	var cache = NewStringCache(
		l,
		backend,
		time.Microsecond,
		"pref",
		handler,
	)

	res, err := cache.Get(ctx, "key12345")

	assert.NoError(err)
	assert.Equal("result", res)
}

func TestStringCache_CacheGetError(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	backend := new(mocks.CacheBackend)
	backend.EXPECT().Get(mock.Anything, "pref/key12345").Return("", fmt.Errorf("error getting cache"))
	defer backend.AssertExpectations(t)

	handler := new(mocks.Handler)
	defer handler.AssertExpectations(t)

	var cache = NewStringCache(
		l,
		backend,
		time.Microsecond,
		"pref",
		handler,
	)

	_, err := cache.Get(ctx, "key12345")
	assert.Error(err)
}
