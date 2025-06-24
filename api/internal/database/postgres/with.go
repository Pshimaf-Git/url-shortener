package postgres

import (
	"time"

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

func WithMaxConnIdleTime(t time.Duration) OptFunc {
	if t <= time.Duration(0) {
		t = defaultMaxConnIdleTime
	}

	return func(c *pgxpool.Config) {
		c.MaxConnIdleTime = t
	}
}

func WithMaxConnLifetime(t time.Duration) OptFunc {
	if t <= time.Duration(0) {
		t = defaultMaxConnLifeTime
	}

	return func(c *pgxpool.Config) {
		c.MaxConnLifetime = t
	}
}
