package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Pshimaf-Git/url-shortener/api/internal/cache"
	"github.com/Pshimaf-Git/url-shortener/api/internal/database"
	"github.com/Pshimaf-Git/url-shortener/api/internal/http-server/reqcontext"
	"github.com/Pshimaf-Git/url-shortener/api/internal/lib/api/resp"
	"github.com/Pshimaf-Git/url-shortener/api/internal/lib/sl"
)

type Responce struct {
	resp.Response
	Alias string `json:"alias"`
}

const maxRetries int = 10

func (h *Handler) NewSave() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.Save"

		c := reqcontext.New(w, r)
		log := h.log.With(
			slog.String("fn", fn),
			slog.String(RequestID, c.RequestID()),
		)

		var req Request

		log.Info("decoding request body")

		if err := c.DecodeJSON(&req); err != nil {
			log.Error("decode req body", sl.Error(err))

			c.JSON(http.StatusInternalServerError, resp.Error(ErrInternalServer))
			return
		}

		log.Info("decoded requst body", slog.Any("request body", req))

		userProvaidedAlias := req.Alias
		url := req.URL
		stdLength := h.cfg.StdAliasLen

		if strings.EqualFold(url, "") {
			log.Info("request without url")

			c.JSON(http.StatusBadRequest, resp.Error(ErrEmprtyURl))
			return
		}

		if !isValidURL(url) {
			h.log.Error("invalid URL format", slog.String("url", req.URL))

			c.JSON(http.StatusBadRequest, resp.Error(ErrInvalidURLFormat))

			return
		}

		var (
			finalAlias string
			err        error
		)

		if strings.EqualFold(userProvaidedAlias, "") {
			finalAlias, err = h.storage.SaveGeneratedURl(c.Context(), url, stdLength, maxRetries)
			if err != nil {
				if errors.Is(err, database.ErrMaxRetriesForGenerate) {
					log.Error("generate randim alias",
						slog.Int("max retries", maxRetries),
						slog.Int("standart alias length", stdLength),
						sl.Error(err),
					)

					c.JSON(http.StatusInternalServerError, resp.Error(ErrCanNotGenAlias))

					return
				}

				log.Error("saving URl with generated alias", slog.String("url", url), sl.Error(err))

				c.JSON(http.StatusInternalServerError, resp.Error(ErrInternalServer))

				return
			}

			log.Info("added alias", slog.String("url", url), slog.String("alias", finalAlias))

			c.JSON(http.StatusCreated, Responce{
				Response: resp.OK(),
				Alias:    finalAlias,
			})

		} else {
			log = log.With(slog.String("url", url), slog.String("alias", userProvaidedAlias))

			err := h.storage.SaveURL(c.Context(), url, userProvaidedAlias)
			if err != nil {
				if errors.Is(err, database.ErrURLExist) {
					log.Info("alias already exist")

					c.JSON(http.StatusBadRequest, resp.Error(ErrAliasExist))
					return
				}

				log.Error("url saver", sl.Error(err))

				c.JSON(http.StatusInternalServerError, resp.Error(ErrInternalServer))
				return
			}

			log.Info("url added")

			c.JSON(http.StatusCreated, Responce{
				Response: resp.OK(),
				Alias:    userProvaidedAlias,
			})
		}
	}
}

func (h *Handler) NewDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.Delete"
		c := reqcontext.New(w, r)

		log := h.log.With(
			slog.String("fn", fn),
			slog.String(RequestID, c.RequestID()),
		)

		alias := c.GetParam("alias")

		if strings.EqualFold(alias, "") {
			log.Info("empty alias")

			c.JSON(http.StatusBadRequest, resp.Error(ErrEmptyAlias))

			return
		}

		log = log.With(slog.String("alias", alias))

		log.Info("decoded requst body")

		_, err := h.storage.DeleteURL(c.Context(), alias)
		if err != nil {
			if errors.Is(err, database.ErrURLNotFound) {
				log.Info("url not found")

				c.JSON(http.StatusBadRequest, resp.Error(ErrURLNotFound))
				return
			}

			log.Error("url deleter", sl.Error(err))

			c.JSON(http.StatusInternalServerError, resp.Error(ErrInternalServer))
			return
		}

		log.Info("url deleted from database")

		if err := h.cache.Delete(c.Context(), alias); err != nil {
			if errors.Is(err, cache.ErrKeyNotExist) {
				log.Info("alias alredy deleted or not found")
			} else {
				log.Error("delete from cache", sl.Error(err))

				c.JSON(http.StatusInternalServerError, Responce{
					Response: resp.Error(ErrInternalServer),
					Alias:    alias,
				})

				return
			}
		}

		log.Info("delete from cache successesfuly")

		c.JSON(http.StatusOK, Responce{
			Response: resp.OK(),
			Alias:    alias,
		})
	}
}

func (h *Handler) NewRedirect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.Redirect"
		c := reqcontext.New(w, r)

		log := h.log.With(
			slog.String("fn", fn),
			slog.String(RequestID, c.RequestID()),
		)

		alias := c.GetParam("alias")
		if strings.TrimSpace(alias) == "" {
			log.Info("empty alias")
			c.JSON(http.StatusBadRequest, resp.Error(ErrEmptyAlias))
			return
		}

		url, err := h.GetURLWithCache(c.Context(), alias)
		if err != nil {
			switch {
			case errors.Is(err, database.ErrURLNotFound):
				log.Info("url not found", slog.String("alias", alias))
				c.JSON(http.StatusNotFound, resp.Error(ErrURLNotFound))
				return

			default:
				log.Error("failed to get URL", slog.String("alias", alias), sl.Error(err))
				c.JSON(http.StatusInternalServerError, resp.Error(ErrInternalServer))
				return
			}
		}

		log.Info("redirecting",
			slog.String("alias", alias),
			slog.String("url", url),
		)

		http.Redirect(
			c.ResponceWriter(),
			c.Request(),
			url,
			http.StatusFound,
		)
	}
}
