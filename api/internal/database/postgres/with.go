package postgres

import (
	"time"

	"github.com/Pshimaf-Git/url-shortener/internal/config"
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

func WithMaxConns(cfg *config.PostreSQLConfig) OptFunc {
	if cfg.Options.MaxConns <= 0 {
		cfg.Options.MaxConns = defaultMaxConn
	}

	return func(c *pgxpool.Config) {
		c.MaxConns = int32(cfg.Options.MaxConns)
	}
}

func WithMinConns(cfg *config.PostreSQLConfig) OptFunc {
	if cfg.Options.MinConns <= 0 {
		cfg.Options.MinConns = defaultMinConn
	}

	return func(c *pgxpool.Config) {
		c.MinConns = int32(cfg.Options.MinConns)
	}
}

func WithCheckHelth(cfg *config.PostreSQLConfig) OptFunc {
	if cfg.Options.CheckHelthPeriod <= time.Duration(0) {
		cfg.Options.CheckHelthPeriod = defaultCheckHelthPeriod
	}

	return func(c *pgxpool.Config) {
		c.HealthCheckPeriod = cfg.Options.CheckHelthPeriod
	}
}

func WithMaxConnIdleTime(cfg *config.PostreSQLConfig) OptFunc {
	if cfg.Options.MaxConnIdleTime <= time.Duration(0) {
		cfg.Options.MaxConnIdleTime = defaultMaxConnIdleTime
	}

	return func(c *pgxpool.Config) {
		c.MaxConnIdleTime = cfg.Options.MaxConnIdleTime
	}
}

func WithMaxConnLifetime(cfg *config.PostreSQLConfig) OptFunc {
	if cfg.Options.MaxConnLifetime <= time.Duration(0) {
		cfg.Options.MaxConnLifetime = defaultMaxConnLifeTime
	}

	return func(c *pgxpool.Config) {
		c.MaxConnLifetime = cfg.Options.MaxConnLifetime
	}
}
