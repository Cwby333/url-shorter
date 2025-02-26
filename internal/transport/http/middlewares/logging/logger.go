package logging

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Cwby333/url-shorter/internal/transport/http/lib/typeasserterror"
)

func New(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger, ok := r.Context().Value("logger").(*slog.Logger)

		err := typeasserterror.Check(ok, w, slog.Default())

		if err != nil {
			return
		}

		reqID := r.Header.Get("X-REQUEST-ID")

		logger = logger.With("component", "middleware logger")
		logger = logger.With(slog.Group("logging", slog.String("host", r.Host), slog.String("port", r.URL.Port()), slog.Int("content-length", int(r.ContentLength)), slog.String("pattern", r.Pattern), slog.String("method", r.Method)), slog.String("reqID", reqID))

		ctx := r.Context()
		ctx = context.WithValue(ctx, "logger", logger)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
