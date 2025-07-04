package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Pshimaf-Git/url-shortener/api/internal/cache"
	"github.com/Pshimaf-Git/url-shortener/api/internal/config"
	"github.com/Pshimaf-Git/url-shortener/api/internal/database"
	"github.com/Pshimaf-Git/url-shortener/api/internal/http-server/reqcontext"
	"github.com/Pshimaf-Git/url-shortener/api/internal/lib/api/resp"
	"github.com/Pshimaf-Git/url-shortener/api/internal/lib/sl"
	"github.com/Pshimaf-Git/url-shortener/api/internal/lib/wraper"
)

const setCacheTimeout = time.Second

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
		if err := h.cache.Expire(ctx, alias); err != nil {
			h.log.Error("expire alias",
				slog.String("key", alias),
				slog.String("value", url),
				sl.Error(err),
			)
		}
		return url, nil
	}

	if !errors.Is(err, cache.ErrKeyNotExist) {
		h.log.Error("get URL from cache",
			slog.String("key", alias),
			sl.Error(err),
		)
	}

	url, err = h.storage.GetURl(ctx, alias)
	if err != nil {
		if errors.Is(err, database.ErrURLNotFound) {
			return "", wp.Wrap(err)
		}
		return "", wp.Wrap(err)
	}

	h.log.Info("starting a setter goroutine", slog.String("alias", alias), slog.String("url", url))

	setCtx, cancel := context.WithTimeout(context.Background(), setCacheTimeout)

	go func() {
		defer cancel()

		if err := h.cache.Set(setCtx, alias, url); err != nil {
			h.log.Error("cache URL",
				slog.String("key", alias),
				slog.String("value", url),
				sl.Error(err),
			)
		}
	}()

	return url, nil
}

func (h *Handler) helthy(w http.ResponseWriter, r *http.Request) {
	c := reqcontext.New(w, r)

	if h.badConfigurate() {
		c.JSON(http.StatusInternalServerError, resp.Error(ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, resp.OK())
}

func (h *Handler) badConfigurate() bool {
	return h.cache == nil || h.storage == nil || h.log == nil || h.cfg == nil
}

func isValidURL(urlToCheck string) bool {
	const (
		HTTP  = "http"
		HTTPs = "https"
	)

	if urlToCheck == "" {
		return false
	}

	if strings.ContainsAny(urlToCheck, " \t\n\r\b\a") {
		return false
	}

	u, err := url.ParseRequestURI(urlToCheck)
	if err != nil {
		return false
	}

	if u.Scheme == "" || u.Host == "" {
		return false
	}

	if u.Scheme != HTTP && u.Scheme != HTTPs {
		return false
	}

	if strings.Contains(u.Host, "..") || strings.Contains(u.Host, " ") {
		return false
	}

	return true
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
