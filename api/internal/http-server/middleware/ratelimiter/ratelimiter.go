package ratelimiter

import (
	"net/http"
	"time"

	"github.com/Pshimaf-Git/url-shortener/api/internal/config"
	"github.com/go-chi/httprate"
)

func New(requstLimit int, windowLength time.Duration, limitHandler http.HandlerFunc, options ...httprate.Option) func(http.Handler) http.Handler {
	options = append(options, httprate.WithKeyByRealIP(), httprate.WithLimitHandler(limitHandler))
	return httprate.Limit(
		requstLimit,
		windowLength,
		options...,
	)
}

func NewWithConfig(cfg *config.ServerConfig, limitHandler http.HandlerFunc, options ...httprate.Option) func(http.Handler) http.Handler {
	return New(cfg.RequesLimit, cfg.WindowLength, limitHandler, options...)
}
