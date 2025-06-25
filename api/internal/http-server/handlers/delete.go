package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Pshimaf-Git/url-shortener/internal/cache"
	"github.com/Pshimaf-Git/url-shortener/internal/database"
	"github.com/Pshimaf-Git/url-shortener/internal/http-server/reqcontext"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/api/resp"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/sl"
)

func (h *Handler) NewDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.Dselete"
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

		log.Error("deleted from cache", sl.Error(err))

		c.JSON(http.StatusOK, Responce{
			Response: resp.OK(),
			Alias:    alias,
		})
	}
}
