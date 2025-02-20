package usersrouter

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Cwby333/url-shorter/internal/transport/httptransport/urlrouter/lib/mainresponse"
	"github.com/Cwby333/url-shorter/pkg/generalerrors"
)

type LogInResponse struct {
	Response mainresponse.Response
}

func (router Router) LogInHandler(w http.ResponseWriter, r *http.Request) {
	const op = "internal/transport/httptransport/usersrouter/LogInHandler"

	logger := r.Context().Value("logger").(*slog.Logger)

	username, password, ok := r.BasicAuth()

	if !ok {
		logger.Info("missing auth header")

		resp, err := newLogInResponse(errors.New("unauthorized"))

		if err != nil {
			logger.Error("json marshal", slog.String("error", err.Error()))

			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		http.Error(w, string(resp), http.StatusUnauthorized)
		return
	}

	if username == "" || password == "" {
		logger.Info("wrong credentials")

		resp, err := newLogInResponse(errors.New("empty username or password"))

		if err != nil {
			logger.Error("json marshal", slog.String("error", err.Error()))

			http.Error(w, "empty username or password", http.StatusUnauthorized)
			return
		}

		http.Error(w, string(resp), http.StatusUnauthorized)
		return
	}

	token, expired, err := router.service.LogIn(r.Context(), username, password)

	if err != nil {
		if errors.Is(err, generalerrors.ErrUserNotFound) {
			logger.Info("wrong username")

			resp, err := newLogInResponse(errors.New("wrong username"))

			if err != nil {
				logger.Error("json marshal", slog.String("error", err.Error()))

				http.Error(w, "wrong username", http.StatusUnauthorized)
				return
			}

			http.Error(w, string(resp), http.StatusUnauthorized)
			return
		}

		if errors.Is(err, generalerrors.ErrWrongPassword) {
			logger.Info("wrong password")

			resp, err := newLogInResponse(errors.New("wrong password"))

			if err != nil {
				logger.Error("json marshal", slog.String("error", err.Error()))

				http.Error(w, "wrong password", http.StatusUnauthorized)

				return
			}

			http.Error(w, string(resp), http.StatusUnauthorized)
			return
		}

		logger.Error("login handler", slog.String("error", err.Error()))

		resp, err := newLogInResponse(errors.New("internal error"))

		if err != nil {
			logger.Error("json marshal", slog.String("error", err.Error()))

			http.Error(w, "internal error", http.StatusUnauthorized)

			return
		}

		http.Error(w, string(resp), http.StatusUnauthorized)
		return
	}

	response := LogInResponse{
		Response: mainresponse.NewOK(),
	}

	data, err := json.Marshal(response)

	if err != nil {
		logger.Error("json marshal", slog.String("error", err.Error()))
		logger.Info("success login handler")

		w.Write([]byte("Success login"))

		w.Header().Set("Authorization", token)
		return
	}

	logger.Info("success login handler")

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt-access",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		Expires: expired,
	})

	w.Write(data)
}

func newLogInResponse(err error) ([]byte, error) {
	resp := LogInResponse{
		Response: mainresponse.NewError(err.Error()),
	}

	out, err := json.Marshal(resp)

	if err != nil {
		return []byte{}, err
	}

	return out, nil
}
