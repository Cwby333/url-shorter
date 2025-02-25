package jwtmiddle

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/Cwby333/url-shorter/internal/transport/http/lib/mainresponse"
	"github.com/Cwby333/url-shorter/internal/transport/http/lib/typeasserterror"
	"github.com/golang-jwt/jwt/v5"
)

func NewAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger, ok := r.Context().Value("logger").(*slog.Logger)

		err := typeasserterror.Check(ok, w, slog.Default())

		if err != nil {
			return
		}

		logger = logger.With("component", "json middleware")

		tokenString := r.Header.Get("Authorization")
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		if tokenString == "" {
			logger.Info("unauthorized")

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

		secretKey := os.Getenv("APP_JWT_SECRET_KEY")

		t, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		},
			jwt.WithIssuer(os.Getenv("APP_JWT_ISSUER")),
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
			jwt.WithExpirationRequired(),
			jwt.WithIssuer(os.Getenv("APP_JWT_ISSUER")))

		if err != nil {
			logger.Info("jwt parse", slog.String("error", err.Error()))

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

		if !t.Valid {
			logger.Info("invalid jwt")

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

		claims, ok := t.Claims.(jwt.MapClaims)

		err = typeasserterror.Check(ok, w, logger)

		if err != nil {
			return
		}

		typeToken, ok := claims["type"].(string)

		err = typeasserterror.Check(ok, w, logger)

		if err != nil {
			return
		}

		if typeToken != "access" {
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
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
