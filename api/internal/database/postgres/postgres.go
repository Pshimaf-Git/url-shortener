package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"

	"github.com/Pshimaf-Git/url-shortener/internal/config"
	"github.com/Pshimaf-Git/url-shortener/internal/database"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/random"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/wraper"
	"github.com/lib/pq"
)

var _ database.Database = &storage{}

type storage struct {
	db *sql.DB
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

const pgUniqueConstraintViolation = "23505"

func New(ctx context.Context, cfg *config.PostreSQLConfig) (*storage, error) {
	const fn = "database.postgres.New"

	wp := wraper.New(fn)

	connStr := BuildConnString(cfg)

	db, err := sql.Open(postgresDriver, connStr)
	if err != nil {
		return nil, wp.WrapMsg("sql.Open", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, wp.WrapMsg("db.Ping", err)
	}

	if _, err := db.ExecContext(ctx, schema); err != nil {
		db.Close()
		return nil, wp.WrapMsg("create table urls", err)
	}

	return &storage{db: db}, nil
}

func (s *storage) SaveURL(ctx context.Context, originalURL string, alias string) error {
	const fn = "database.postgres.(*storage).SaveURL"

	wp := wraper.New(fn)

	query := `INSERT INTO urls(url, alias) VALUES($1, $2)`

	_, err := s.db.ExecContext(ctx, query, originalURL, alias)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code.Name() == pgUniqueConstraintViolation {
			return wp.WrapMsg("alias already exists", database.ErrURLExist)
		}

		return wp.WrapMsg("failed to save URL", err)
	}
	return nil
}

func (s *storage) SaveGeneratedURl(ctx context.Context, originalURL string, length, maxAttempts int) (string, error) {
	const fn = "database.postgres.(*storage).SaveGeneratedURl"

	wp := wraper.New(fn)

	for i := 0; i < maxAttempts; i++ {
		alias := random.StringRandV2(length)

		query := `INSERT INTO urls(url, alias) VALUES($1, $2) RETURNING alias`

		var insertedAlias string
		row := s.db.QueryRowContext(ctx, query, originalURL, alias)

		err := row.Scan(&insertedAlias)
		if err == nil {
			return insertedAlias, nil
		}

		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code.Name() == pgUniqueConstraintViolation {
			continue
		}

		return "", wp.Wrap(err)
	}

	return "", wp.Wrap(database.ErrMaxRetriesForGenerate)
}

func (s *storage) GetURl(ctx context.Context, alias string) (string, error) {
	const fn = "database.postgres.(*storage).GetURL"

	wp := wraper.New(fn)

	row := s.db.QueryRowContext(ctx, "SELECT url FROM urls WHERE alias=$1", alias)

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

	stmt, err := s.db.PrepareContext(ctx, "DELETE FROM urls WHERE alias=$1")
	if err != nil {
		return 0, wp.WrapMsg("prepare delete statement", err)
	}

	res, err := stmt.ExecContext(ctx, alias)
	if err != nil {
		return 0, wp.Wrap(err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return 0, wp.WrapMsg("get last insert id", err)
	}

	return n, nil
}

func (s *storage) Close() error {
	const fn = "database.postgres.(*storage).Close"

	wp := wraper.New(fn)

	if err := s.db.Close(); err != nil {
		return wp.WrapMsg("db.Close", err)
	}

	return nil
}

// Exmaple output: postgres://user:password@host:port/database?ssl_mode=enable
func BuildConnString(cfg *config.PostreSQLConfig) string {
	s := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.User, cfg.Password, net.JoinHostPort(cfg.Host, cfg.Port), cfg.Name, cfg.SSLMode,
	)

	return s
}
