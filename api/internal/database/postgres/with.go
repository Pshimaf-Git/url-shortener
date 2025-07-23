package postgres

import (
	"time"

	"github.com/Pshimaf-Git/url-shortener/api/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxConn          = 10
	defaultMinConn          = 5
	defaultCheckHelthPeriod = time.Minute
	defaultMaxConnIdleTime  = time.Minute
	defaultMaxConnLifeTime  = time.Minute
)

type OptFunc func(*pgxpool.Config)

func WithMaxConns(max int32) OptFunc {
	if max <= 0 {
		max = defaultMaxConn
	}

	return func(c *pgxpool.Config) {
		c.MaxConns = max
	}
}

func WithMinConns(min int32) OptFunc {
	if min <= 0 {
		min = defaultMinConn
	}

	return func(c *pgxpool.Config) {
		c.MinConns = min
	}
}

func WithCheckHelth(period time.Duration) OptFunc {
	if period <= time.Duration(0) {
		period = defaultCheckHelthPeriod
	}

	return func(c *pgxpool.Config) {
		c.HealthCheckPeriod = period
	}
}

func WithMaxConnIdleTime(timeout time.Duration) OptFunc {
	if timeout <= time.Duration(0) {
		timeout = defaultMaxConnIdleTime
	}

	return func(c *pgxpool.Config) {
		c.MaxConnIdleTime = timeout
	}
}

func WithMaxConnLifetime(timeout time.Duration) OptFunc {
	if timeout <= time.Duration(0) {
		timeout = defaultMaxConnLifeTime
	}

	return func(c *pgxpool.Config) {
		c.MaxConnLifetime = timeout
	}
}

func WithConfig(cfg *config.OptionalPostgreSQLConfig) OptFunc {
	return func(c *pgxpool.Config) {
		WithMinConns(int32(cfg.MinConns))
		WithMaxConns(int32(cfg.MaxConns))
		WithMaxConnLifetime(cfg.MaxConnLifetime)
		WithMaxConnIdleTime(cfg.MaxConnIdleTime)
		WithCheckHelth(cfg.CheckHelthPeriod)
	}
}
