package redis

import (
	"context"
	"testing"
	"time"

	"github.com/Pshimaf-Git/url-shortener/internal/cache"
	"github.com/Pshimaf-Git/url-shortener/internal/config"
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

		// Verify value in miniredis
		val, err := mr.Get("test-key")
		assert.NoError(t, err)
		assert.Equal(t, "test-value", val)

		// Verify TTL was set
		ttl := mr.TTL("test-key")
		assert.Greater(t, ttl, time.Duration(0))
	})

	t.Run("empty key", func(t *testing.T) {
		err := client.Set(ctx, "", "value")
		assert.ErrorIs(t, err, cache.ErrEmptyKey)
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
}

func TestDelete(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		mr.Set("to-delete", "value")
		assert.True(t, mr.Exists("to-delete"))

		err := client.Delete(ctx, "to-delete")
		assert.NoError(t, err)
		assert.False(t, mr.Exists("to-delete"))
	})

	t.Run("key not found", func(t *testing.T) {
		err := client.Delete(ctx, "non-existent-key")
		assert.ErrorIs(t, err, cache.ErrKeyNotExist)
	})

	t.Run("empty key", func(t *testing.T) {
		err := client.Delete(ctx, "")
		assert.ErrorIs(t, err, cache.ErrEmptyKey)
	})
}

func TestClose(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	err := client.Close()
	assert.NoError(t, err)

	// Verify connection is closed by trying to perform an operation
	err = client.Set(context.Background(), "test", "value")
	assert.Error(t, err)
}
