package limitermidde

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Cwby333/url-shorter/internal/transport/http/lib/mainresponse"
	"github.com/Cwby333/url-shorter/internal/transport/http/ratelimiter"
)

func New(limiter ratelimiter.Limiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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

			logger = logger.With("component", "ratelimiter middleware")

			IP := r.RemoteAddr
			err := limiter.Iterate(IP)

			if err != nil {
				logger.Info("forbidden by ratelimiter", slog.String("IP", IP))

				resp := mainresponse.NewError("please, repeat after a few minutes")
				data, err := json.Marshal(resp)

				if err != nil {
					logger.Error("json marshal", slog.String("error", err.Error()))

					http.Error(w, "please, repeat after a few minutes", http.StatusTooManyRequests)
					return
				}

				http.Error(w, string(data), http.StatusTooManyRequests)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, "logger", logger)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
