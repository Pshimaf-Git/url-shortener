package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/Pshimaf-Git/url-shortener/internal/cache/redis"
	"github.com/Pshimaf-Git/url-shortener/internal/config"
	"github.com/Pshimaf-Git/url-shortener/internal/database/postgres"
	"github.com/Pshimaf-Git/url-shortener/internal/http-server/handlers"
	"github.com/Pshimaf-Git/url-shortener/internal/http-server/middleware/logger"
	"github.com/Pshimaf-Git/url-shortener/internal/http-server/server"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/errors"
	"github.com/go-chi/chi/v5/middleware"
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

	signal.Notify(done, os.Interrupt, os.Kill)

	<-done

	fmt.Fprintln(os.Stderr, "server stoped")

	cancel()
}

func realMain(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log, err := setupLogger(cfg)
	if err != nil {
		return err
	}

	log.Info("server starting", slog.String("env", cfg.Env), slog.String("host", cfg.Server.Host), slog.String("port", cfg.Server.Port))

	db, err := postgres.New(ctx, &cfg.Postgres)
	if err != nil {
		return err
	}
	defer db.Close()

	cache, err := redis.New(&cfg.Redis)

	h := handlers.New(db, cache, cfg, log)

	mws := []func(http.Handler) http.Handler{
		logger.New(log),
		middleware.RequestID,
		middleware.Recoverer}

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

func setupLogger(cfg *config.Config) (*slog.Logger, error) {
	lvl, err := cfg.LevelFromString()
	if err != nil {
		return nil, errors.Wrap("setupLogger", "", err)
	}

	switch strings.ToLower(strings.TrimSpace(cfg.Env)) {
	case prod:
		return slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}),
		), nil

	case dev:
		return slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}),
		), nil

	case local:
		return slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}),
		), nil
	}

	return nil, errors.New("setupLogger: unknown environment")
}
