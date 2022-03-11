package stringcache

import (
	"context"
	"testing"
	"time"

	"github.com/emerishq/emeris-utils/store"

	"github.com/alicebob/miniredis/v2"

	"github.com/stretchr/testify/assert"
)

// NewStore returns an initialized test *Store connected to a miniredis instance.
// The miniredis instance is also returned for convenience and further customizing the test environment.
func NewStore(t *testing.T) (*store.Store, *miniredis.Miniredis) {
	s := miniredis.RunT(t)
	storeClient, err := store.NewClient(s.Addr())
	if err != nil {
		t.Fatalf("creating a new Store: %v", err)
	}
	return storeClient, s
}

func TestStoreBackend_SetNewKey(t *testing.T) {
	assert := assert.New(t)
	s, _ := NewStore(t)

	back := NewStoreBackend(s)
	err := back.Set(context.Background(), "key", "value", 10*time.Minute)
	assert.NoError(err)
}

func TestStoreBackend_ReplaceKey(t *testing.T) {
	assert := assert.New(t)
	s, miniredis := NewStore(t)
	_ = miniredis.Set("key", "something")

	back := NewStoreBackend(s)
	err := back.Set(context.Background(), "key", "value", 10*time.Minute)

	assert.NoError(err)

	val, _ := miniredis.Get("key")
	assert.Equal("value", val)
}

func TestStoreBackend_GetKey(t *testing.T) {
	assert := assert.New(t)
	s, miniredis := NewStore(t)
	_ = miniredis.Set("key", "something")

	back := NewStoreBackend(s)
	res, err := back.Get(context.Background(), "key")

	assert.NoError(err)
	assert.Equal("something", res)
}

func TestStoreBackend_GetUnsetKey(t *testing.T) {
	assert := assert.New(t)
	s, _ := NewStore(t)

	back := NewStoreBackend(s)
	_, err := back.Get(context.Background(), "key")

	assert.ErrorIs(ErrCacheMiss, err)
}

func TestStoreBackend_GetAfterExpiration(t *testing.T) {
	assert := assert.New(t)
	s, miniredis := NewStore(t)

	back := NewStoreBackend(s)
	err := back.Set(context.Background(), "key", "value", 60*time.Second)
	assert.NoError(err)

	miniredis.FastForward(10 * time.Minute)

	res, err := back.Get(context.Background(), "key")
	assert.ErrorIsf(ErrCacheMiss, err, "expected cache miss, actual cache value was: %s", res)
}
