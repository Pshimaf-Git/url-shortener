package redis

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/Pshimaf-Git/url-shortener/internal/cache"
	"github.com/Pshimaf-Git/url-shortener/internal/config"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/errors"
	"github.com/redis/go-redis/v9"
)

// Interface implementation checks at compile time
// Ensures redisClient satisfies all required cache interfaces
var (
	_ cache.Setter = &redisClient{} // Verify Setter interface implementation
	_ cache.Getter = &redisClient{} // Verify Getter interface implementation
	_ cache.Cache  = &redisClient{} // Verify full Cache interface implementation
)

// redisClient implements Redis-based caching
type redisClient struct {
	cfg *config.RedisCongig
	rdb *redis.Client
}

// New creates and returns a new Redis client instance
func New(cfg *config.RedisCongig) (*redisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &redisClient{rdb: rdb, cfg: cfg}, nil
}

// Set stores a key-value pair in Redis with configured TTL
func (r *redisClient) Set(ctx context.Context, key string, value any) error {
	const fn = "cache.redis.(*cache).Set"

	if strings.EqualFold(key, "") {
		return errors.Wrap(fn, "", cache.ErrEmptyKey)
	}

	if err := r.rdb.Set(ctx, key, value, r.cfg.TTL).Err(); err != nil {
		return errors.Wrap(fn, fmt.Sprintf("key=%s val=%v", key, value), err)
	}

	return nil
}

// Get retrieves a value from Redis by key
func (r *redisClient) Get(ctx context.Context, key string) (string, error) {
	const fn = "cache.redis.(*cache).Get"

	value, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", errors.Wrap(fn, fmt.Sprintf("key=%s", key), cache.ErrKeyNotExist)
		}

		return "", errors.Wrap(fn, fmt.Sprintf("key=%s", key), err)
	}

	return value, nil
}

func (r *redisClient) Expire(ctx context.Context, key string) error {
	const fn = "cache.redis.(*cache).Expire"

	err := r.rdb.Expire(ctx, key, r.cfg.TTL).Err()
	if err != nil {
		if err == redis.Nil {
			return cache.ErrKeyNotExist
		}
		return errors.Wrap(fn, fmt.Sprintf("key=%s", key), err)
	}

	return nil
}

// Close terminates the Redis connection
func (r *redisClient) Close() error {
	const fn = "cache.redis.(*cache).Close"
	if r.rdb == nil {
		return nil
	}
	r.rdb.Expire(context.Background(), "", 0)

	return errors.Wrap(fn, "", r.rdb.Close())
}
