package logging

import (
	"context"
	"log/slog"
	"net/http"
)

func New(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger, ok := r.Context().Value("logger").(*slog.Logger)

		if !ok {
			logger.Debug("cannot find logger in context")
			next.ServeHTTP(w, r)

			return
		}

		reqID := r.Header.Get("X-REQUEST-ID")

		logger = logger.With("component", "middleware logger")

		logger = logger.With(slog.Group("logging", slog.String("host", r.Host), slog.String("port", r.URL.Port()), slog.Int("content-length", int(r.ContentLength)), slog.String("pattern", r.Pattern), slog.String("method", r.Method)), slog.String("reqID", reqID))

		newCtx := context.WithValue(r.Context(), "logger", logger)

		r = r.WithContext(newCtx)

		next.ServeHTTP(w, r)
	})
}
