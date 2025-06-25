package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Pshimaf-Git/url-shortener/internal/cache"
	"github.com/Pshimaf-Git/url-shortener/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE:
//
// BEFORE EXEC THESE COMMAND IN TETMINAL
//
// docker run --name TEST-REDIS -p 6379:6379 -d redis:7.0-alpine
//
// these command create a docker container with pistgres database for test
//
// AFTER EXEC THESE COMMAND
//
// docker stop TEST-REDIS && docker rm TEST-REDIS

func setupTestRedis(t *testing.T) (*redisClient, context.Context, func()) {
	cfg := &config.RedisCongig{
		Host:     "localhost",
		Port:     "6379",
		Password: "",
		DB:       1,
		TTL:      10 * time.Minute,
	}

	ctx := context.Background()
	client, err := New(ctx, cfg)
	require.NoError(t, err)

	// Clear test database
	err = client.rdb.FlushDB(ctx).Err()
	require.NoError(t, err)

	return client, ctx, func() {
		client.Close()
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.RedisCongig
		wantError error
	}{
		{
			name: "Success - valid config",
			cfg: &config.RedisCongig{
				Host: "localhost",
				Port: "6379",
				DB:   1,
			},
		},
		{
			name: "Error - invalid host",
			cfg: &config.RedisCongig{
				Host: "invalid-host",
				Port: "6379",
			},
			wantError: errors.New("redis.client.Ping"),
		},
		{
			name: "Error - invalid port",
			cfg: &config.RedisCongig{
				Host: "localhost",
				Port: "99999",
			},
			wantError: errors.New("redis.client.Ping"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := New(ctx, tt.cfg)

			if tt.wantError != nil {
				assert.ErrorContains(t, err, tt.wantError.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestSet(t *testing.T) {
	client, ctx, teardown := setupTestRedis(t)
	defer teardown()

	tests := []struct {
		name      string
		key       string
		value     string
		wantError error
	}{
		{
			name:  "Success - set key-value",
			key:   "test-key",
			value: "test-value",
		},
		{
			name:      "Error - empty key",
			key:       "",
			value:     "test-value",
			wantError: cache.ErrEmptyKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.Set(ctx, tt.key, tt.value)

			if tt.wantError != nil {
				assert.ErrorIs(t, err, tt.wantError)
				return
			}

			assert.NoError(t, err)
			// Verify the value was set
			got, err := client.rdb.Get(ctx, tt.key).Result()
			assert.NoError(t, err)
			assert.Equal(t, tt.value, got)
		})
	}
}

func TestGet(t *testing.T) {
	client, ctx, teardown := setupTestRedis(t)
	defer teardown()

	// Setup test data
	testKey := "existing-key"
	testValue := "test-value"
	err := client.Set(ctx, testKey, testValue)
	require.NoError(t, err)

	tests := []struct {
		name      string
		key       string
		want      string
		wantError error
	}{
		{
			name: "Success - get existing key",
			key:  testKey,
			want: testValue,
		},
		{
			name:      "Error - non-existent key",
			key:       "non-existent-key",
			wantError: cache.ErrKeyNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.Get(ctx, tt.key)

			if tt.wantError != nil {
				assert.ErrorIs(t, err, tt.wantError)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExpire(t *testing.T) {
	client, ctx, teardown := setupTestRedis(t)
	defer teardown()

	// Setup test data
	testKey := "expire-test-key"
	err := client.Set(ctx, testKey, "test-value")
	require.NoError(t, err)

	tests := []struct {
		name      string
		key       string
		wantError error
	}{
		{
			name: "Success - set expiration on existing key",
			key:  testKey,
		},
		{
			name:      "Error - non-existent key",
			key:       "non-existent-key",
			wantError: cache.ErrKeyNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.Expire(ctx, tt.key)

			if tt.wantError != nil {
				assert.ErrorIs(t, err, tt.wantError)
				return
			}

			assert.NoError(t, err)
			// Verify TTL was set
			ttl, err := client.rdb.TTL(ctx, tt.key).Result()
			assert.NoError(t, err)
			assert.Greater(t, ttl, time.Duration(0))
		})
	}
}

func TestDelete(t *testing.T) {
	client, ctx, teardown := setupTestRedis(t)
	defer teardown()

	// Setup test data
	testKey := "delete-test-key"
	err := client.Set(ctx, testKey, "test-value")
	require.NoError(t, err)

	tests := []struct {
		name      string
		key       string
		wantError error
	}{
		{
			name: "Success - delete existing key",
			key:  testKey,
		},
		{
			name:      "Error - non-existent key",
			key:       "non-existent-key",
			wantError: cache.ErrKeyNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.Delete(ctx, tt.key)

			if tt.wantError != nil {
				assert.ErrorIs(t, err, tt.wantError)
				return
			}

			assert.NoError(t, err)
			// Verify key was deleted
			_, err = client.rdb.Get(ctx, tt.key).Result()
			assert.ErrorIs(t, err, redis.Nil)
		})
	}
}

func TestClose(t *testing.T) {
	client, ctx, _ := setupTestRedis(t)

	tests := []struct {
		name      string
		setup     func() *redisClient
		wantError bool
	}{
		{
			name: "Success - close connection",
			setup: func() *redisClient {
				return client
			},
		},
		{
			name: "Success - nil client",
			setup: func() *redisClient {
				return &redisClient{rdb: nil}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			err := c.Close()

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify connection is closed
			if c.rdb != nil {
				err = c.rdb.Ping(ctx).Err()
				assert.Error(t, err)
			}
		})
	}
}
