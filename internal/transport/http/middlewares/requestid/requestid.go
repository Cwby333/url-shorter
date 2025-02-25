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
			reqID := r.Header.Get("X-REQUEST-ID")

			if reqID == "" {
				reqID = uuid.NewString()

				r.Header.Set("X-REQUEST-ID", reqID)
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, "logger", logger)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
