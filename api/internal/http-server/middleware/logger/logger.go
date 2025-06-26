package logger

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Pshimaf-Git/url-shortener/api/internal/http-server/handlers"
	"github.com/go-chi/chi/v5/middleware"
)

const reqID = handlers.RequestID

func New(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log.Info("using midleware/Logger")

		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := log.With(
				slog.String(reqID, middleware.GetReqID(r.Context())),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("URI", r.RequestURI),
				slog.String("user_agent", r.UserAgent()),
			)

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now().UTC()
			defer func() {
				entry.Info("request completed",
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.String("duration", time.Since(t1).String()),
				)

			}()

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
