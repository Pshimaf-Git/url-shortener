package handlers

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/Pshimaf-Git/url-shortener/internal/lib/api/resp"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/sl"
)

func (h *Handler) NewDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.Dselete"
		c := NewContext(w, r)

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

		affected, err := h.storage.DeleteURL(c.Context(), alias)
		if err != nil {
			log.Error("url deleter", sl.Error(err))

			c.JSON(http.StatusInternalServerError, resp.Error(ErrInternalServer))
			return
		}

		if affected == 0 {
			log.Info("url not found")

			c.JSON(http.StatusBadRequest, resp.Error(ErrURLNotFound))
			return
		}

		log.Info("url deleted")

		c.JSON(http.StatusOK, Responce{
			Response: resp.OK(),
			Alias:    alias,
		})
	}
}
