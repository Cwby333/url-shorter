package basicauth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Cwby333/url-shorter/internal/transport/httptransport/urlrouter/lib/mainresponse"
)

func New(username, password string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger, ok := r.Context().Value("logger").(*slog.Logger)

			if !ok {
				logger.Debug("cannot find logger in context")
				next.ServeHTTP(w, r)

				return
			}

			logger = logger.With("component", "middleware basic auth")

			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				logger.Info("unset auth header")

				response := mainresponse.NewError(errors.New("unset Authorization Header").Error())

				data, err := json.Marshal(response)

				if err != nil {
					logger.Error("json marshall", slog.String("error", err.Error()))

					http.Error(w, "unset Authorization Header", http.StatusUnauthorized)

					return
				}

				http.Error(w, string(data), http.StatusUnauthorized)

				return
			}

			rUser, rPass, _ := r.BasicAuth()

			if rUser != username || rPass != password {
				logger.Info("forbidden for user", slog.String("username", rUser))

				response := mainresponse.NewError(errors.New("wrong username or password").Error())

				data, err := json.Marshal(response)

				if err != nil {
					logger.Error("json marshall", slog.String("error", err.Error()))

					http.Error(w, "unset Authorization Header", http.StatusUnauthorized)

					return
				}

				http.Error(w, string(data), http.StatusUnauthorized)

				return
			}

			ctx := r.Context()

			ctx = context.WithValue(ctx, "logger", logger)

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
