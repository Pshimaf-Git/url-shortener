package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/url"

	"github.com/Pshimaf-Git/url-shortener/internal/cache"
	"github.com/Pshimaf-Git/url-shortener/internal/config"
	"github.com/Pshimaf-Git/url-shortener/internal/database"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/sl"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/wraper"
)

type Handler struct {
	cache   cache.Cache
	storage database.Database
	log     *slog.Logger
	cfg     *config.ServerConfig
}

func New(storage database.Database, cache cache.Cache, cfg *config.ServerConfig, log *slog.Logger) *Handler {
	return &Handler{
		storage: storage,
		cache:   cache,
		cfg:     cfg,
		log:     log,
	}
}

func (h *Handler) GetURLWithCache(ctx context.Context, alias string) (string, error) {
	const fn = "handlers.handler.(*Handler).GetURLWithCache"

	wp := wraper.New(fn)

	url, err := h.cache.Get(ctx, alias)
	if err == nil {
		h.cache.Expire(ctx, alias)
		return url, nil
	}

	if !errors.Is(err, cache.ErrKeyNotExist) {
		return "", wp.Wrap(err)
	}

	url, err = h.storage.GetURl(ctx, alias)
	if err != nil {
		if errors.Is(err, database.ErrURLNotFound) {
			return "", wp.Wrap(err)
		}
		return "", wp.Wrap(err)
	}

	h.log.Info("starting a setter goroutine", slog.String("alias", alias), slog.String("url", url))

	go func() {
		if err := h.cache.Set(ctx, alias, url); err != nil {
			if errors.Is(err, cache.ErrKeyNotExist) {
				h.log.Info("alias not found in cache", slog.String("alias", alias))
				return
			}

			h.log.Error("cache URL",
				slog.String("key", alias),
				slog.String("value", url),
				sl.Error(err),
			)
		}
	}()

	return url, nil
}

func ValidateURLFormat(urlToCheck string) error {
	_, err := url.ParseRequestURI(urlToCheck)
	return err
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
	ErrCanNotGenAlias   = errors.New("could not generate random unique alias, please try again")
)

const (
	RequestID = "request_id"
)
