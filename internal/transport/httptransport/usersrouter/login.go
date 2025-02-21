package usersrouter

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Cwby333/url-shorter/internal/transport/httptransport/urlrouter/lib/mainresponse"
	validaterequests "github.com/Cwby333/url-shorter/internal/transport/httptransport/urlrouter/lib/validaterequsts"
	"github.com/Cwby333/url-shorter/pkg/generalerrors"
	"github.com/go-playground/validator/v10"
)

type LogInRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LogInResponse struct {
	Response mainresponse.Response
}

func (router Router) LogInHandler(w http.ResponseWriter, r *http.Request) {
	const op = "internal/transport/httptransport/usersrouter/LogInHandler"

	logger := r.Context().Value("logger").(*slog.Logger)
	logger = logger.With("component", "login handler")

	var req LogInRequest

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		logger.Error("json decoder", slog.String("error", err.Error()))

		resp, err := newLogInResponse(err)

		if err != nil {
			logger.Error("json marshal", slog.String("error", err.Error()))

			http.Error(w, "internal error", http.StatusInternalServerError)

			return
		}

		http.Error(w, string(resp), http.StatusInternalServerError)

		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	err = validate.Struct(req)

	if err != nil {
		logger.Info("bad request", slog.String("error", err.Error()))

		errForResp := err.(validator.ValidationErrors)

		resp := validaterequests.Validate(errForResp)

		response := LogInResponse{
			Response: mainresponse.NewError(resp...),
		}

		data, err := json.Marshal(response)

		if err != nil {
			logger.Error("json marshal", slog.String("error", err.Error()))

			http.Error(w, "bad request", http.StatusBadRequest)

			return
		}

		http.Error(w, string(data), http.StatusBadRequest)
		return
	}

	token, expired, err := router.service.LogIn(r.Context(), req.Username, req.Password)

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
		Expires:  expired,
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
