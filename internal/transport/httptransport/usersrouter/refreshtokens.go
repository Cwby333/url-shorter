package usersrouter

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Cwby333/url-shorter/internal/transport/httptransport/urlrouter/lib/mainresponse"
	"github.com/Cwby333/url-shorter/pkg/generalerrors"
	"github.com/golang-jwt/jwt/v5"
)

type RefreshTokensResponse struct {
	Response mainresponse.Response
	Message  string `json:"message"`
}

func (router Router) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	logger := r.Context().Value("logger").(*slog.Logger)
	logger = logger.With("component", "refresh tokens handler")

	claims := r.Context().Value("claims").(jwt.MapClaims)
	tokenID := claims["jti"].(string)
	sub := claims["sub"].(string)
	dur := time.Duration(int64(claims["exp"].(float64) * 1000000000)).Seconds()
	dur = (dur - float64(time.Now().Unix())) * 1000000000

	err := router.service.CheckCountOfUses(r.Context(), tokenID, time.Duration(dur))

	if err != nil {
		if errors.Is(err, generalerrors.ErrToManyUseOfRefreshToken) {
			logger.Error("to many uses of refresh token", slog.String("tokenID", tokenID))

			resp := RefreshTokensResponse{
				Response: mainresponse.NewError("unauthorized"),
				Message:  "please change credentials",
			}
			data, err := json.Marshal(resp)

			if err != nil {
				logger.Error("json marshal", slog.String("error", err.Error()))

				http.Error(w, "unauthorized, please change credentials", http.StatusUnauthorized)
				return
			}

			http.Error(w, string(data), http.StatusUnauthorized)
			return
		}

		logger.Error("refresh", slog.String("error", err.Error()))

		resp := RefreshTokensResponse{
			Response: mainresponse.NewError("unauthorized"),
			Message:  "please change credentials",
		}
		data, err := json.Marshal(resp)

		if err != nil {
			logger.Error("json marshal", slog.String("error", err.Error()))

			http.Error(w, "unauthorized, please change credentials", http.StatusUnauthorized)
			return
		}

		http.Error(w, string(data), http.StatusUnauthorized)
		return
	}

	err = router.service.CheckBlacklist(r.Context(), tokenID)

	if err != nil {
		if errors.Is(err, generalerrors.ErrRefreshInBlackList) {
			logger.Info("refresh token in blacklist", slog.String("tokenID", tokenID))

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

		logger.Error("", slog.String("error", err.Error()))

		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	accessClaims, refreshClaims, err := router.service.CreateJWT(r.Context(), sub)

	if err != nil {
		logger.Error("create jwt", slog.String("error", err.Error()))

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

	response := LoginResponse{
		Response: mainresponse.NewOK(),
	}

	data, err := json.Marshal(response)

	if err != nil {
		logger.Error("json marshal", slog.String("error", err.Error()))
		logger.Info("success login handler")

		w.Write([]byte("Success login"))

		http.SetCookie(w, &http.Cookie{
			Name:     "jwt-access",
			Value:    accessClaims.Sign,
			HttpOnly: true,
			Secure:   true,
			Expires:  accessClaims.ExpiresAt.Time,
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh-token",
			Value:    refreshClaims.Sign,
			HttpOnly: true,
			Secure:   true,
			Expires:  refreshClaims.ExpiresAt.Time,
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt-access",
		Value:    accessClaims.Sign,
		HttpOnly: true,
		Expires:  accessClaims.ExpiresAt.Time,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh-token",
		Value:    refreshClaims.Sign,
		HttpOnly: true,
		Expires:  refreshClaims.ExpiresAt.Time,
		Path:     "/api/users/refresh",
	})

	err = router.service.UseRefresh(r.Context(), tokenID)

	if err != nil {
		logger.Error("refresh", slog.String("error", err.Error()))

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

	logger.Info("success refresh handler")

	w.Write(data)
}
