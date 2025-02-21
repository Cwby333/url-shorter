package jwtmiddle

import (
	"log/slog"
	"net/http"
)

func New(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const op = "internal/transport/httptransort/middlewares/jwtmiddle"

		logger := r.Context().Value("logger").(*slog.Logger)
		logger = logger.With("component", "json middleware")

	})
}
