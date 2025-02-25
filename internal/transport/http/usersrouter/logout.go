package usersrouter

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/Cwby333/url-shorter/internal/transport/http/lib/mainresponse"
	"github.com/golang-jwt/jwt/v5"
)

const (
	SuccessLogout = "success logout"
)

type LogoutResponse struct {
	Response mainresponse.Response
	Message  string `json:"message"`
}

func (router Router) Logout(w http.ResponseWriter, r *http.Request) {
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

	logger = logger.With("component", "logout handler")

	claims := r.Context().Value("claims").(jwt.MapClaims)
	dur := time.Duration(int64(claims["exp"].(float64) * 1000000000)).Seconds()
	dur = (dur - float64(time.Now().Unix())) * 1000000000

	err := router.service.LogOut(r.Context(), claims["jti"].(string), time.Duration(dur))

	if err != nil {
		logger.Error("logout", slog.String("error", err.Error()))

		response := mainresponse.NewError("internal error")
		data, err := json.Marshal(response)

		if err != nil {
			logger.Error("json marshal", slog.String("error", err.Error()))

			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	resp := LogoutResponse{
		Response: mainresponse.NewOK(),
		Message:  SuccessLogout,
	}
	data, err := json.Marshal(resp)

	if err != nil {
		logger.Error("json marshal", slog.String("error", err.Error()))

		_, err = w.Write([]byte("success logout"))

		if err != nil {
			logger.Error("response write", slog.String("error", err.Error()))
		}

		return
	}

	logger.Info("success logout")

	_, err = w.Write(data)

	if err != nil {
		logger.Error("response write", slog.String("error", err.Error()))

		resp := mainresponse.NewError("internal error")
		data, err := json.Marshal(resp)

		if err != nil {
			logger.Error("json marshal", slog.String("error", err.Error()))

			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		http.Error(w, string(data), http.StatusInternalServerError)
	}
}
