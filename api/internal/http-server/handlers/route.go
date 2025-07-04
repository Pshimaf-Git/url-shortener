package handlers

import (
	"net/http"

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
