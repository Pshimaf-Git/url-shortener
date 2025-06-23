package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Pshimaf-Git/url-shortener/internal/database"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/api/resp"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/sl"
)

func (h *Handler) NewRedirect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.Redirect"
		c := NewContext(w, r)

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
