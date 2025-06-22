package handlers

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/Pshimaf-Git/url-shortener/internal/cache"
	"github.com/Pshimaf-Git/url-shortener/internal/config"
	"github.com/Pshimaf-Git/url-shortener/internal/database"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/errors"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/sl"
)

type Handler struct {
	cache   cache.Cache
	storage database.Database
	cfg     *config.Config
	log     *slog.Logger
}

func New(storage database.Database, cache cache.Cache, cfg *config.Config, log *slog.Logger) *Handler {
	return &Handler{
		storage: storage,
		cache:   cache,
		cfg:     cfg,
		log:     log,
	}
}

func (h *Handler) GetURLWithCache(ctx context.Context, alias string) (string, error) {
	const fn = "handlers.handler.(*Handler).GetURLWithCache"

	url, err := h.cache.Get(ctx, alias)
	if err == nil {
		h.cache.Expire(ctx, alias)
		return url, nil
	}

	if !errors.Is(err, cache.ErrKeyNotExist) {
		return "", errors.Wrap(fn, "", err)
	}

	url, err = h.storage.GetURl(ctx, alias)
	if err != nil {
		if errors.Is(err, database.ErrURLNotFound) {
			return "", errors.Wrap(fn, "", err)
		}
		return "", errors.Wrap(fn, "", err)
	}

	if err := h.cache.Set(ctx, alias, url); err != nil {
		h.log.Error("cache URL",
			slog.String("key", alias),
			slog.String("value", url),
			sl.Error(err),
		)
	}

	return url, nil
}

func ValidateURLFormat(urlToCheck string) error {
	if _, err := url.ParseRequestURI(urlToCheck); err != nil {
		return err
	}

	return nil
}

type Request struct {
	Alias string `json:"alias,omitempty"`
	URL   string `json:"url"`
}

var (
	ErrURLNotFound      = errors.New("url not found")
	ErrEmptyAlias       = errors.New("url must not be empty")
	ErrEmprtyURl        = errors.New("url must not be empty")
	ErrInternalServer   = errors.New("internal server error")
	ErrAliasExist       = errors.New("alias already exist")
	ErrInvalidURLFormat = errors.New("invalid url format")
)

const (
	RequestID = "request_id"
)
