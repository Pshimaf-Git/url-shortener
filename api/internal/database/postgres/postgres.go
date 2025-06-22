package postgres

import (
	"database/sql"
	"fmt"
	"net"

	"context"

	"github.com/Pshimaf-Git/url-shortener/internal/config"
	"github.com/Pshimaf-Git/url-shortener/internal/database"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/errors"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
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

func New(ctx context.Context, cfg *config.PostreSQLConfig) (*storage, error) {
	const fn = "database.postgres.New"

	connStr := BuildConnString(cfg)

	db, err := sql.Open(postgresDriver, connStr)
	if err != nil {
		return nil, errors.Wrap(fn, "sql.Open", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Wrap(fn, "db.Ping", err)
	}

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return nil, errors.Wrap(fn, "create table urls", err)
	}

	return &storage{db: db}, nil
}

func (s *storage) SaveURL(ctx context.Context, userURl string, alias string) error {
	const fn = "database.postgres.(*storage).SaveURL"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(fn, "starts a transaction", err)
	}

	stmt, err := tx.PrepareContext(ctx, "SELECT id FROM urls WHERE alias=$1")
	if err != nil {
		_ = tx.Rollback()
		return errors.Wrap(fn, "prepare transaction select statement", err)
	}

	rows, err := stmt.QueryContext(ctx, alias)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			_ = tx.Rollback()
			return errors.Wrap(fn, "", err)
		}
	}
	rows.Close()

	stmt, err = tx.PrepareContext(ctx, "INSERT INTO urls(url, alias) VALUES($1, $2)")
	if err != nil {
		_ = tx.Rollback()
		return errors.Wrap(fn, "prepare transaction insert statement", err)
	}

	_, err = stmt.ExecContext(ctx, userURl, alias)
	if err != nil {
		_ = tx.Rollback()
		return errors.Wrap(fn, "", database.ErrURLExist)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return errors.Wrap(fn, "transaction commit", err)
	}

	return nil
}

func (s *storage) GetURl(ctx context.Context, alias string) (string, error) {
	const fn = "database.postgres.(*storage).GetURL"

	stmt, err := s.db.PrepareContext(ctx, "SELECT url FROM urls WHERE alias=$1")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.Wrap(fn, "", database.ErrURLNotFound)
		}

		return "", errors.Wrap(fn, "", database.ErrURLNotFound)
	}

	row := stmt.QueryRowContext(ctx, alias)

	var url string
	if err := row.Scan(&url); err != nil {
		return "", errors.Wrap(fn, "", database.ErrURLNotFound)
	}

	return url, nil
}

func (s *storage) DeleteURL(ctx context.Context, alias string) (int64, error) {
	const fn = "database.postgres.(*storage).DeleteURL"

	stmt, err := s.db.PrepareContext(ctx, "DELETE FROM urls WHERE alias=$1")
	if err != nil {
		return 0, errors.Wrap(fn, "prepare dalete statement", err)
	}

	res, err := stmt.ExecContext(ctx, alias)
	if err != nil {
		return 0, errors.Wrap(fn, "", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(fn, "get last insert id", err)
	}

	return n, nil
}

func (s *storage) Close() error {
	const fn = "database.postgres.(*storage).Close"

	if err := s.db.Close(); err != nil {
		return errors.Wrap(fn, "db.Close", err)
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
