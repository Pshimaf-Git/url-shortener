package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/Pshimaf-Git/url-shortener/internal/lib/sl"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres driver
	_ "github.com/golang-migrate/migrate/v4/source/file"       // file driver
)

func main() {
	var migrTable, migrDBURL, migrDir string

	flag.StringVar(&migrDBURL, "db", "", "connection string for database")
	flag.StringVar(&migrDir, "dir", "", "path to directory with migration files")
	flag.StringVar(&migrTable, "table", "migrations", "table with migrations")
	flag.Parse()

	log := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	if err := validate(migrDBURL, migrDir); err != nil {
		log.Error("invalid input", sl.Error(err))
		return
	}

	m, err := migrate.New(
		"file://"+migrDir,
		fmt.Sprintf("postgres://%s?x-migrations-table=%s", migrDBURL, migrTable),
	)

	log = log.With(
		slog.String("dir", migrDir),
		slog.String("table", migrTable),
	)

	if err != nil {
		log.Error("migrate.New",
			sl.Error(err),
		)
		return
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info("no migrations to apply")
			return
		}

		log.Error("during Up", sl.Error(err))
		return
	}

	// if err := closeMigrate(m); err != nil {
	// 	log.Error("migrate closing", sl.Error(err))
	// 	return
	// }

	log.Info("migrations apply successfully")
}

func validate(migrDBURL, migrDir string) error {
	errs := make([]error, 0, 2)

	if migrDBURL == "" {
		errs = append(errs, errors.New("db url must not be empty"))
	}

	if migrDir == "" {
		errs = append(errs, errors.New("migration directory path must not be empty"))
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

func closeMigrate(m *migrate.Migrate) error {
	errs := make([]error, 0, 2)

	serr, dberr := m.Close()

	if serr != nil {
		errs = append(errs, fmt.Errorf("migrate source error: %w", serr))
	}

	if dberr != nil {
		errs = append(errs, fmt.Errorf("migrate databse error: %w", dberr))
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}
