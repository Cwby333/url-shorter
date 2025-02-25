package logging

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Cwby333/url-shorter/internal/transport/http/lib/mainresponse"
)

func New(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger, ok := r.Context().Value("logger").(*slog.Logger)

		if !ok {
			slog.Error("wrong type assertion to logger")

			resp := mainresponse.NewError("internal error")
			data, err := json.Marshal(resp)

			if err != nil {
				slog.Error("json marshal", slog.String("error", err.Error()))

				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			http.Error(w, string(data), http.StatusInternalServerError)
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
