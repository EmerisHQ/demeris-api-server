package stringcache

import (
	"context"
	"testing"
	"time"

	"github.com/emerishq/emeris-utils/store"

	"github.com/alicebob/miniredis/v2"

	"github.com/stretchr/testify/require"
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
	require := require.New(t)
	s, _ := NewStore(t)

	back := NewStoreBackend(s)
	err := back.Set(context.Background(), "key", "value", 10*time.Minute)
	require.NoError(err)
}

func TestStoreBackend_ReplaceKey(t *testing.T) {
	require := require.New(t)
	s, miniredis := NewStore(t)
	_ = miniredis.Set("key", "something")

	back := NewStoreBackend(s)
	err := back.Set(context.Background(), "key", "value", 10*time.Minute)

	require.NoError(err)

	val, _ := miniredis.Get("key")
	require.Equal("value", val)
}

func TestStoreBackend_GetKey(t *testing.T) {
	require := require.New(t)
	s, miniredis := NewStore(t)
	_ = miniredis.Set("key", "something")

	back := NewStoreBackend(s)
	res, err := back.Get(context.Background(), "key")

	require.NoError(err)
	require.Equal("something", res)
}

func TestStoreBackend_GetUnsetKey(t *testing.T) {
	require := require.New(t)
	s, _ := NewStore(t)

	back := NewStoreBackend(s)
	_, err := back.Get(context.Background(), "key")

	require.ErrorIs(ErrCacheMiss, err)
}

func TestStoreBackend_GetAfterExpiration(t *testing.T) {
	require := require.New(t)
	s, miniredis := NewStore(t)

	back := NewStoreBackend(s)
	err := back.Set(context.Background(), "key", "value", 60*time.Second)
	require.NoError(err)

	miniredis.FastForward(10 * time.Minute)

	res, err := back.Get(context.Background(), "key")
	require.ErrorIsf(ErrCacheMiss, err, "expected cache miss, actual cache value was: %s", res)
}
