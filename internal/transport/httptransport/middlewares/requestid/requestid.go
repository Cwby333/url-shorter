package requestid

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

func New(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		logger = logger.With("component", "middleware reqId")

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqId := r.Header.Get("X-REQUEST-ID")

			if reqId == "" {
				reqId = uuid.NewString()

				r.Header.Set("X-REQUEST-ID", reqId)
			}

			ctx := r.Context()

			logger = logger.With("request_id", reqId)
			ctx = context.WithValue(ctx, "logger", logger)

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
