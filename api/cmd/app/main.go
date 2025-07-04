package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Pshimaf-Git/url-shortener/api/internal/cache/redis"
	"github.com/Pshimaf-Git/url-shortener/api/internal/config"
	"github.com/Pshimaf-Git/url-shortener/api/internal/database/postgres"
	"github.com/Pshimaf-Git/url-shortener/api/internal/http-server/handlers"
	"github.com/Pshimaf-Git/url-shortener/api/internal/http-server/middleware/logger"
	"github.com/Pshimaf-Git/url-shortener/api/internal/http-server/server"
	"github.com/Pshimaf-Git/url-shortener/api/internal/lib/logger/zaphandler"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

const (
	prod  = "prod"
	dev   = "dev"
	local = "local"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan os.Signal, 1)

	go func() {
		if err := realMain(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "server running error: %s\n", err.Error())
			cancel()
			return
		}
	}()

	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	<-done

	fmt.Fprintln(os.Stderr, "server stoped")

	cancel()
}

func realMain(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log, cleanup, err := setupLogger(cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	log.Info("server starting", slog.String("env", cfg.Env), slog.String("host", cfg.Server.Host), slog.String("port", cfg.Server.Port))

	db, err := postgres.New(ctx, &cfg.Postgres,
		postgres.WithMaxConns(&cfg.Postgres),
		postgres.WithMinConns(&cfg.Postgres),
		postgres.WithMaxConnLifetime(&cfg.Postgres),
		postgres.WithMaxConnIdleTime(&cfg.Postgres),
		postgres.WithCheckHelth(&cfg.Postgres),
	)

	if err != nil {
		return err
	}
	defer db.Close()

	cache, err := redis.New(ctx, &cfg.Redis)

	h := handlers.New(db, cache, &cfg.Server, log)

	mws := []func(http.Handler) http.Handler{
		middleware.RequestID,
		middleware.Recoverer,
		httprate.LimitByRealIP(cfg.Server.RequesLimit, cfg.Server.WindowLength),
		logger.New(log),
	}

	router := h.InitRoutes(mws...)

	serv := server.New(
		server.WithAddr(server.BuildAddr(cfg.Server.Host, cfg.Server.Port)),
		server.WithHandler(router),
		server.WithReadTimeout(cfg.Server.ReadTimeout),
		server.WithWriteTimeout(cfg.Server.WriteTimeout),
		server.WithIdleTimeout(cfg.Server.IdleTimeout),
	)

	defer serv.Shutdown(ctx)

	if err := serv.Run(); err != nil {
		return err
	}

	return nil
}

func setupLogger(cfg *config.Config) (*slog.Logger, func() error, error) {
	lvl, err := cfg.Logger.LevelFromString()
	if err != nil {
		return nil, nil, err
	}

	var (
		h       slog.Handler
		cleanup func() error
		herr    error
	)

	switch strings.ToLower(strings.TrimSpace(cfg.Env)) {
	case prod:
		zapH, err := zaphandler.NewProduction(lvl)
		cleanup = zapH.Logger().Sync
		herr = err
		h = zapH
	case dev:
		zapH, err := zaphandler.NewDevelopment(lvl)
		cleanup = zapH.Logger().Sync
		herr = err
		h = zapH
	case local:
		zapH, err := zaphandler.NewLocal(lvl)
		cleanup = zapH.Logger().Sync
		herr = err
		h = zapH
	}

	if h == nil {
		return nil, nil, errors.New("setupLogger: unknown environment")
	}

	if herr != nil {
		return nil, nil, herr
	}

	return slog.New(h), cleanup, nil
}
