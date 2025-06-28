package handlers

import (
	"net/http"

	"github.com/Pshimaf-Git/url-shortener/api/internal/http-server/reqcontext"
	"github.com/Pshimaf-Git/url-shortener/api/internal/lib/api/resp"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) InitRoutes(middlewares ...func(http.Handler) http.Handler) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middlewares...)

	router.Get("/helthy", h.helthy)

	router.Get("/api/v1/url", h.NewRedirect())
	router.Post("/api/v1/url", h.NewSave())
	router.Delete("/api/v1/url", h.NewDelete())

	return router
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
