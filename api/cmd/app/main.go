package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Pshimaf-Git/url-shortener/api/internal/cache/redis"
	"github.com/Pshimaf-Git/url-shortener/api/internal/config"
	"github.com/Pshimaf-Git/url-shortener/api/internal/database/postgres"
	"github.com/Pshimaf-Git/url-shortener/api/internal/http-server/handlers"
	mwlogger "github.com/Pshimaf-Git/url-shortener/api/internal/http-server/middleware/logger"
	"github.com/Pshimaf-Git/url-shortener/api/internal/http-server/middleware/ratelimiter"
	"github.com/Pshimaf-Git/url-shortener/api/internal/http-server/server"
	"github.com/Pshimaf-Git/url-shortener/api/internal/lib/logger/zaphandler"
	"github.com/Pshimaf-Git/url-shortener/api/internal/lib/sl"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// init context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // ensure context is cancelled on exit

	// init config
	cfg, err := config.Load()
	if err != nil {
		panic(err) // handle error appropriately
	}

	// init logger
	logger, sync, err := zaphandler.NewLogger(cfg.Env, level(&cfg.Logger))
	if err != nil {
		panic(err) // handle error appropriately
	}

	defer func() {
		if err := sync(); err != nil {
			slog.Error("failed to sync logger", sl.Error(err))
		}
	}()

	// init database
	db, err := postgres.New(ctx, &cfg.Postgres, postgres.WithConfig(&cfg.Postgres.Options))
	if err != nil {
		logger.Error("failed to initialize database", sl.Error(err))
		return // handle error appropriately
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close database connection", sl.Error(err))
		}
	}()

	// init cache
	cache, err := redis.New(ctx, &cfg.Redis)
	if err != nil {
		logger.Error("failed to initialize cache", sl.Error(err))
		return // handle error appropriately
	}

	defer func() {
		if err := cache.Close(); err != nil {
			logger.Error("failed to close cache connection", sl.Error(err))
		}
	}()

	// init handler
	handler := handlers.New(db, cache, &cfg.Server, logger)

	// middlewares
	middlewares := []func(http.Handler) http.Handler{
		middleware.Recoverer,
		middleware.RealIP,
		middleware.RequestID,
		mwlogger.New(logger),
		ratelimiter.New(cfg.Server.RequesLimit, cfg.Server.WindowLength, handler.NewLimit()),
	}

	// init router
	router := handler.InitRoutes(middlewares...)

	// init server
	server := server.NewWithConfig(&cfg.Server, router)

	// Signals
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	// start server
	logger.Info("starting server", slog.String("address", address(&cfg.Server)))

	if err := server.Run(); err != nil {
		logger.Error("failed to start server", sl.Error(err))
		return // handle error appropriately
	}
}

func level(cfg *config.LoggerConfig) slog.Level {
	lvl, err := cfg.LevelFromString()
	if err != nil {
		slog.Error("failed to parse log level, using level info", sl.Error(err))
		return slog.LevelInfo // default level if parsing fails
	}
	return lvl
}

func address(cfg *config.ServerConfig) string {
	return net.JoinHostPort(cfg.Host, cfg.Port)
}
