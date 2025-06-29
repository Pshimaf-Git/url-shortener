package redis

import (
	"context"
	"testing"
	"time"

	"github.com/Pshimaf-Git/url-shortener/api/internal/cache"
	"github.com/Pshimaf-Git/url-shortener/api/internal/config"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) (*redisClient, *miniredis.Miniredis) {
	t.Helper()

	// Create miniredis instance
	mr, err := miniredis.Run()
	require.NoError(t, err)

	// Create config with miniredis address
	cfg := &config.RedisCongig{
		Host:     "localhost",
		Port:     mr.Port(),
		Password: "",
		DB:       0,
		TTL:      10 * time.Minute,
	}

	// Create client
	client, err := New(context.Background(), cfg)
	require.NoError(t, err)

	return client, mr
}

func TestNew(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		client, mr := setupTestRedis(t)
		defer mr.Close()
		defer client.Close()

		assert.NotNil(t, client.rdb)
		assert.Equal(t, client.cfg.TTL, 10*time.Minute)
	})

	t.Run("connection failure", func(t *testing.T) {
		cfg := &config.RedisCongig{
			Host:     "invalid",
			Port:     "1234",
			Password: "",
			DB:       0,
			TTL:      10 * time.Minute,
		}

		client, err := New(context.Background(), cfg)
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestSet(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()

	t.Run("successful set", func(t *testing.T) {
		err := client.Set(ctx, "test-key", "test-value")
		assert.NoError(t, err)

		val, err := mr.Get("test-key")
		assert.NoError(t, err)
		assert.Equal(t, "test-value", val)

		ttl := mr.TTL("test-key")
		assert.Greater(t, ttl, time.Duration(0))
	})

	t.Run("empty key", func(t *testing.T) {
		err := client.Set(ctx, "", "value")
		assert.ErrorIs(t, err, cache.ErrEmptyKey)
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := client.Set(ctx, "cancelled-key", "value")
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
		assert.False(t, mr.Exists("cancelled-key"))
	})
}

func TestGet(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()

	t.Run("successful get", func(t *testing.T) {
		mr.Set("existing-key", "existing-value")
		mr.SetTTL("existing-key", 10*time.Minute)

		val, err := client.Get(ctx, "existing-key")
		assert.NoError(t, err)
		assert.Equal(t, "existing-value", val)
	})

	t.Run("key not found", func(t *testing.T) {
		_, err := client.Get(ctx, "non-existent-key")
		assert.ErrorIs(t, err, cache.ErrKeyNotExist)
	})

	t.Run("empty key", func(t *testing.T) {
		_, err := client.Get(ctx, "")
		assert.ErrorIs(t, err, cache.ErrEmptyKey)
	})

	t.Run("cancelled context", func(t *testing.T) {
		mr.Set("ctx-key", "value")

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := client.Get(ctx, "ctx-key")
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestExpire(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()

	t.Run("successful expire", func(t *testing.T) {
		mr.Set("test-key", "test-value")
		initialTTL := time.Second
		mr.SetTTL("test-key", initialTTL)

		err := client.Expire(ctx, "test-key")
		assert.NoError(t, err)

		newTTL := mr.TTL("test-key")
		assert.Equal(t, client.cfg.TTL, newTTL)
	})

	t.Run("key not found", func(t *testing.T) {
		err := client.Expire(ctx, "non-existent-key")
		assert.ErrorIs(t, err, cache.ErrKeyNotExist)
	})

	t.Run("empty key", func(t *testing.T) {
		err := client.Expire(ctx, "")
		assert.ErrorIs(t, err, cache.ErrEmptyKey)
	})

	t.Run("cancelled context", func(t *testing.T) {
		mr.Set("ctx-key", "value")

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := client.Expire(ctx, "ctx-key")
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestDelete(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		k := "to-delete"
		err := mr.Set(k, "value")
		require.NoError(t, err)
		assert.True(t, mr.Exists(k))

		err = client.Delete(ctx, k)
		assert.NoError(t, err)
		assert.False(t, mr.Exists(k))
	})

	t.Run("key not found", func(t *testing.T) {
		err := client.Delete(ctx, "non-existent-key")
		assert.Error(t, err)
		assert.ErrorIs(t, err, cache.ErrKeyNotExist)
	})

	t.Run("empty key", func(t *testing.T) {
		err := client.Delete(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, cache.ErrEmptyKey)
	})

	t.Run("cancelled context", func(t *testing.T) {
		k := "ctx-delete"
		err := mr.Set(k, "value")
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = client.Delete(ctx, k)
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestClose(t *testing.T) {
	t.Run("good_rdb", func(t *testing.T) {
		client, mr := setupTestRedis(t)
		defer mr.Close()

		err := client.Close()
		assert.NoError(t, err)

		// Verify connection is closed by trying to perform an operation
		err = client.Set(context.Background(), "test", "value")
		assert.Error(t, err)
	})

	t.Run("nil_rdb", func(t *testing.T) {
		client := redisClient{rdb: nil}

		err := client.Close()
		assert.NoError(t, err)
	})
}

func Test_ping(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	t.Run("happy_path", func(t *testing.T) {
		err := ping(context.Background(), client.rdb)
		assert.NoError(t, err)
	})

	t.Run("context_canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := ping(ctx, client.rdb)
		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("nil_rdb", func(t *testing.T) {
		err := ping(context.Background(), nil)
		assert.Error(t, err)
	})
}
