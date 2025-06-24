package redis

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/Pshimaf-Git/url-shortener/internal/cache"
	"github.com/Pshimaf-Git/url-shortener/internal/config"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/sl"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/wraper"
	"github.com/redis/go-redis/v9"
)

// Interface implementation checks at compile time
// Ensures redisClient satisfies all required cache interfaces
var (
	_ cache.Setter  = &redisClient{} // Verify Setter interface implementation
	_ cache.Getter  = &redisClient{} // Verify Getter interface implementation
	_ cache.Deleter = &redisClient{} // Verify Deleter interface implementation
	_ cache.Cache   = &redisClient{} // Verify full Cache interface implementation
)

const (
	maxPingRetries = 5
	pingTimeout    = time.Millisecond * 500
)

// redisClient implements Redis-based caching
type redisClient struct {
	cfg *config.RedisCongig
	rdb *redis.Client
}

// New creates and returns a new Redis client instance
func New(ctx context.Context, cfg *config.RedisCongig) (*redisClient, error) {
	const fn = "cache.redis.(*redisClient).New"

	wp := wraper.New(fn)

	rdb := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := pingWithRetries(ctx, rdb); err != nil {
		return nil, wp.Wrap(err)
	}

	return &redisClient{rdb: rdb, cfg: cfg}, nil
}

func pingWithRetries(ctx context.Context, rdb *redis.Client) error {
	const fn = "cache.redis.(*redisClient).pingWithRetries"

	wp := wraper.New(fn)

	if rdb == nil {
		return wp.Wrap(errors.New("nil redis client"))
	}

	var i int
	for ; i < maxPingRetries; i++ {
		if err := rdb.Ping(ctx).Err(); err != nil {
			slog.Error("redis.Client.Ping",
				slog.Int("attempts left", maxPingRetries-i),
				sl.Error(err),
			)

			time.Sleep(pingTimeout)
			continue
		}

		break
	}

	if i == maxPingRetries {
		rdb.Close()
		return wp.Wrap(errors.New("db.Ping, maxPingRetries"))
	}

	return nil
}

// Set stores a key-value pair in Redis with configured TTL
func (r *redisClient) Set(ctx context.Context, key string, value any) error {
	const fn = "cache.redis.(*redisClient).Set"

	wp := wraper.New(fn)

	if strings.EqualFold(key, "") {
		return wp.Wrap(cache.ErrEmptyKey)
	}

	if err := r.rdb.Set(ctx, key, value, r.cfg.TTL).Err(); err != nil {
		return wp.Wrapf(err, "key=%s val=%v", key, value)
	}

	return nil
}

// Get retrieves a value from Redis by key
func (r *redisClient) Get(ctx context.Context, key string) (string, error) {
	const fn = "cache.redis.(*redisClient).Get"

	wp := wraper.New(fn)

	value, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", wp.Wrapf(cache.ErrKeyNotExist, "key=%s", key)
		}

		return "", wp.Wrapf(err, "key=%s", key)
	}

	return value, nil
}

func (r *redisClient) Expire(ctx context.Context, key string) error {
	const fn = "cache.redis.(*redisClient).Expire"

	wp := wraper.New(fn)

	err := r.rdb.Expire(ctx, key, r.cfg.TTL).Err()
	if err != nil {
		if err == redis.Nil {
			return cache.ErrKeyNotExist
		}
		return wp.Wrapf(err, "key=%s", key)
	}

	return nil
}

// Delete delete value from redis by key
func (r *redisClient) Delete(ctx context.Context, key string) error {
	const fn = "cache.redis.(*redisClient).Delete"

	wp := wraper.New(fn)

	if err := r.rdb.Del(ctx, key).Err(); err != nil {
		if err == redis.Nil {
			return cache.ErrKeyNotExist
		}
		return wp.Wrapf(err, "key=%s", key)
	}

	return nil
}

// Close terminates the Redis connection
func (r *redisClient) Close() error {
	const fn = "cache.redis.(*redisClient).Close"

	wp := wraper.New(fn)

	if r.rdb == nil {
		return nil
	}
	r.rdb.Expire(context.Background(), "", 0)

	return wp.Wrap(r.rdb.Close())
}
