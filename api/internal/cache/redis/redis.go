package redis

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/Pshimaf-Git/url-shortener/internal/cache"
	"github.com/Pshimaf-Git/url-shortener/internal/config"
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

	if err := ping(ctx, rdb); err != nil {
		return nil, wp.Wrap(err)
	}

	return &redisClient{rdb: rdb, cfg: cfg}, nil
}

func ping(ctx context.Context, rdb *redis.Client) error {
	const fn = "cache.redis.(*redisClient).ping"

	wp := wraper.New(fn)

	if rdb == nil {
		return wp.Wrap(errors.New("nil redis client"))
	}

	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return wp.Wrap(errors.New("redis.client.Ping"))
	}

	return nil
}

// Set stores a key-value pair in Redis with configured TTL
func (r *redisClient) Set(ctx context.Context, key string, value any) error {
	const fn = "cache.redis.(*redisClient).Set"

	wp := wraper.New(fn)

	if isEmpty(key) {
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

	if isEmpty(key) {
		return "", wp.Wrap(cache.ErrEmptyKey)
	}

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

	if isEmpty(key) {
		return wp.Wrap(cache.ErrEmptyKey)
	}

	cmd := r.rdb.Expire(ctx, key, r.cfg.TTL)
	if err := cmd.Err(); err != nil {
		return wp.Wrapf(err, "key=%s", key)
	}

	expiredSet, err := cmd.Result()
	if err != nil {
		return wp.Wrapf(err, "key=%s", key)
	}

	if !expiredSet {
		return wp.Wrap(cache.ErrKeyNotExist)
	}

	return nil
}

// Delete delete value from redis by key
func (r *redisClient) Delete(ctx context.Context, key string) error {
	const fn = "cache.redis.(*redisClient).Delete"

	wp := wraper.New(fn)

	if isEmpty(key) {
		return wp.Wrap(cache.ErrEmptyKey)
	}

	cmd := r.rdb.Del(ctx, key)
	if err := cmd.Err(); err != nil {
		return wp.Wrapf(err, "key=%s", key)
	}

	n, err := cmd.Result()

	if err != nil {
		return wp.Wrapf(err, "key=%s", key)
	}

	if n == 0 {
		return wp.Wrapf(cache.ErrKeyNotExist, "key=%s", key)
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

	return wp.Wrap(r.rdb.Close())
}

func isEmpty(key string) bool {
	return strings.TrimSpace(key) == ""
}
