package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Pshimaf-Git/url-shortener/internal/config"
	"github.com/Pshimaf-Git/url-shortener/internal/database"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/random"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/wraper"
)

var _ database.Database = &storage{}

type storage struct {
	pool *pgxpool.Pool
}

const postgresDriver = "postgres"

const schema = `
	CREATE TABLE IF NOT EXISTS urls (
  id BIGSERIAL PRIMARY KEY,
  url TEXT NOT NULL,
  alias TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_alias ON urls(alias);
`

const pgconnUniqueConstraintViolation = "23505"

func New(ctx context.Context, cfg *config.PostreSQLConfig, opts ...OptFunc) (*storage, error) {
	const fn = "database.postgres.New"

	wp := wraper.New(fn)

	poolConfig, err := pgxpool.ParseConfig(BuildConnString(cfg))
	if err != nil {
		return nil, wp.WrapMsg("parse pgxpool config from database URl", err)
	}

	executeOptFuncs(poolConfig, opts...)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, wp.WrapMsg("pgxpool.NewWithConfig", err)
	}

	if err := ping(ctx, pool); err != nil {
		return nil, wp.Wrap(err)
	}

	if _, err := pool.Exec(ctx, schema); err != nil {
		pool.Close()
		return nil, wp.WrapMsg("create table urls", err)
	}

	return &storage{pool: pool}, nil
}

func (s *storage) SaveURL(ctx context.Context, originalURL string, alias string) error {
	const fn = "database.postgres.(*storage).SaveURL"

	wp := wraper.New(fn)

	query := `INSERT INTO urls(url, alias) VALUES($1, $2)`

	_, err := s.pool.Exec(ctx, query, originalURL, alias)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgconnUniqueConstraintViolation {
			return wp.WrapMsg("alias already exists", database.ErrURLExist)
		}

		return wp.WrapMsg("failed to save URL", err)
	}
	return nil
}

func executeOptFuncs(c *pgxpool.Config, opts ...OptFunc) {
	if c == nil || len(opts) == 0 {
		return
	}

	for _, fn := range opts {
		fn(c)
	}
}

func ping(ctx context.Context, p *pgxpool.Pool) error {
	const fn = "database.postgres.ping"

	wp := wraper.New(fn)

	if p == nil {
		return wp.Wrap(errors.New("nil pgx pool"))
	}

	if err := p.Ping(ctx); err != nil {
		p.Close()
		return wp.Wrap(errors.New("db.Ping"))
	}

	return nil

}

func (s *storage) SaveGeneratedURl(ctx context.Context, originalURL string, length, maxAttempts int) (string, error) {
	const fn = "database.postgres.(*storage).SaveGeneratedURl"

	wp := wraper.New(fn)

	if err := validateSaveGeneratedURl(originalURL, length, maxAttempts); err != nil {
		return "", wp.Wrap(err)
	}

	for i := 0; i < maxAttempts; i++ {
		alias := random.StringRandV2(length)

		query := `INSERT INTO urls(url, alias) VALUES($1, $2) RETURNING alias`

		var insertedAlias string
		row := s.pool.QueryRow(ctx, query, originalURL, alias)

		err := row.Scan(&insertedAlias)
		if err == nil {
			return insertedAlias, nil
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgconnUniqueConstraintViolation {
			continue
		}

		return "", wp.Wrap(err)
	}

	return "", wp.Wrap(database.ErrMaxRetriesForGenerate)
}

func validateSaveGeneratedURl(originalURL string, length, maxAttempts int) error {
	errs := make([]error, 0, 3)
	if originalURL == "" {
		errs = append(errs, errors.New("url must not be empty"))
	}

	if length <= 0 {
		errs = append(errs, errors.New("invlid alias length"))
	}

	if maxAttempts <= 0 {
		errs = append(errs, errors.New("invalid max attemts"))
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

func (s *storage) GetURl(ctx context.Context, alias string) (string, error) {
	const fn = "database.postgres.(*storage).GetURL"

	wp := wraper.New(fn)

	row := s.pool.QueryRow(ctx, "SELECT url FROM urls WHERE alias=$1", alias)

	var url string
	err := row.Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", wp.WrapMsg("url not found", database.ErrURLNotFound)
		}

		return "", wp.WrapMsg("failed to get URL", err)
	}

	return url, nil
}

func (s *storage) DeleteURL(ctx context.Context, alias string) (int64, error) {
	const fn = "database.postgres.(*storage).DeleteURL"

	wp := wraper.New(fn)

	query := `DELETE FROM urls WHERE alias=$1`

	res, err := s.pool.Exec(ctx, query, alias)
	if err != nil {
		return 0, wp.Wrap(err)
	}

	return res.RowsAffected(), nil
}

func (s *storage) Close() error {
	if s.pool == nil {
		return nil
	}

	s.pool.Close()
	return nil
}

// Exmaple output: postgres://user:password@host:port/database?ssl_mode=enable
func BuildConnString(cfg *config.PostreSQLConfig) string {
	s := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.User, cfg.Password, net.JoinHostPort(cfg.Host, cfg.Port), cfg.Name, cfg.SSLMode,
	)

	return s
}
