package jwtmiddle

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/Cwby333/url-shorter/internal/transport/httpsrv/urlrouter/lib/mainresponse"
	"github.com/golang-jwt/jwt/v5"
)

func NewRefresh(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value("logger").(*slog.Logger)
		logger = logger.With("component", "logout")

		refreshToken := r.Header.Get("Authorization")
		refreshToken = strings.TrimPrefix(refreshToken, "Bearer ")

		if refreshToken == "" {
			logger.Info("missed auth header, must invalid refresh token")

			resp := mainresponse.NewError("send refresh token")

			data, err := json.Marshal(resp)

			if err != nil {
				logger.Error("json marshal", slog.String("error", err.Error()))

				http.Error(w, "send refresh token", http.StatusBadRequest)
				return
			}

			http.Error(w, string(data), http.StatusBadRequest)
			return
		}

		t, err := jwt.ParseWithClaims(refreshToken, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("APP_JWT_SECRET_KEY")), nil
		}, jwt.WithIssuer(os.Getenv("APP_JWT_ISSUER")), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}), jwt.WithExpirationRequired(), jwt.WithIssuer(os.Getenv("APP_JWT_ISSUER")))

		if err != nil {
			logger.Info("jwt parse", slog.String("error", err.Error()))

			resp := mainresponse.NewError("unauthorized")

			data, err := json.Marshal(resp)

			if err != nil {
				logger.Error("json marshal", slog.String("error", err.Error()))

				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			http.Error(w, string(data), http.StatusUnauthorized)
			return
		}

		if !t.Valid {
			logger.Info("invalid token")

			resp := mainresponse.NewError("unauthorized")

			data, err := json.Marshal(resp)

			if err != nil {
				logger.Error("json marshal", slog.String("error", err.Error()))

				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			http.Error(w, string(data), http.StatusUnauthorized)
			return
		}

		claims, ok := t.Claims.(jwt.MapClaims)

		if !ok {
			logger.Error("wrong type assertion to jwt.MapClaims")

			resp := mainresponse.NewError("internal error")

			data, err := json.Marshal(resp)

			if err != nil {
				logger.Error("json marshal", slog.String("error", err.Error()))

				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			http.Error(w, string(data), http.StatusInternalServerError)
			return
		}

		if claims["type"] != "refresh" {
			logger.Info("wrong token type")

			resp := mainresponse.NewError("unauthorized")

			data, err := json.Marshal(resp)

			if err != nil {
				logger.Error("json marshall", slog.String("error", err.Error()))

				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			http.Error(w, string(data), http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "logger", logger)
		ctx = context.WithValue(ctx, "claims", claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
