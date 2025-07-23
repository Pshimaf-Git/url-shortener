package handlers

import (
	"log/slog"
	"net/http"

	"github.com/Pshimaf-Git/url-shortener/api/internal/http-server/reqcontext"
	"github.com/Pshimaf-Git/url-shortener/api/internal/lib/api/resp"
	"github.com/go-chi/httprate"
)

func (h *Handler) NewLimit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.Limit"

		c := reqcontext.New(w, r)

		log := h.log.With(
			slog.String("fn", fn),
			slog.String(RequestID, c.RequestID()),
		)

		realIP, _ := httprate.KeyByRealIP(r)

		log.Warn("requsts limit",
			slog.String("ip", realIP),
		)

		c.JSON(http.StatusTooManyRequests, resp.Error(ErrTooManyRequests))
	}
}
