package usersrouter

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Cwby333/url-shorter/internal/generalerrors"
	"github.com/Cwby333/url-shorter/internal/transport/http/lib/mainresponse"
	validaterequests "github.com/Cwby333/url-shorter/internal/transport/http/lib/validaterequsts"

	"github.com/go-playground/validator/v10"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Response mainresponse.Response
}

func newLoginResponse(err error) ([]byte, error) {
	const op = "internal/transport/httpsrv/usersrouter/login.go/newSaveResponse"

	resp := LoginResponse{
		Response: mainresponse.NewError(err.Error()),
	}

	out, err := json.Marshal(resp)

	if err != nil {
		return []byte{}, fmt.Errorf("%s: %w", op, err)
	}

	return out, nil
}

func (router Router) Login(w http.ResponseWriter, r *http.Request) {
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

	logger = logger.With("component", "login handler")

	req := LoginRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		logger.Error("json decoder", slog.String("error", err.Error()))

		resp, err := newLoginResponse(err)

		if err != nil {
			logger.Error("json marshal", slog.String("error", err.Error()))

			http.Error(w, "internal error", http.StatusInternalServerError)

			return
		}

		http.Error(w, string(resp), http.StatusInternalServerError)

		return
	}

	r.Body.Close()

	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(req)

	if err != nil {
		logger.Info("bad request", slog.String("error", err.Error()))

		errForResp := err.(validator.ValidationErrors)
		resp := validaterequests.Validate(errForResp)
		response := LoginResponse{
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

	accessClaims, refreshClaims, err := router.service.LogIn(r.Context(), req.Username, req.Password)

	if err != nil {
		if errors.Is(err, generalerrors.ErrUserNotFound) {
			logger.Info("wrong username")

			resp, err := newLoginResponse(errors.New("wrong username"))

			if err != nil {
				logger.Error("json marshal", slog.String("error", err.Error()))

				http.Error(w, "wrong username", http.StatusUnauthorized)
				return
			}

			http.Error(w, string(resp), http.StatusUnauthorized)
			return
		}
		if errors.Is(err, generalerrors.ErrWrongPassword) {
			logger.Info("wrong password", slog.String("error", err.Error()))

			resp, err := newLoginResponse(errors.New("wrong password"))

			if err != nil {
				logger.Error("json marshal", slog.String("error", err.Error()))

				http.Error(w, "wrong password", http.StatusUnauthorized)
				return
			}

			http.Error(w, string(resp), http.StatusUnauthorized)
			return
		}

		logger.Error("login handler", slog.String("error", err.Error()))

		resp, err := newLoginResponse(errors.New("internal error"))

		if err != nil {
			logger.Error("json marshal", slog.String("error", err.Error()))

			http.Error(w, "internal error", http.StatusUnauthorized)
			return
		}

		http.Error(w, string(resp), http.StatusUnauthorized)
		return
	}

	response := LoginResponse{
		Response: mainresponse.NewOK(),
	}
	data, err := json.Marshal(response)

	if err != nil {
		logger.Error("json marshal", slog.String("error", err.Error()))
		logger.Info("success login handler")

		_, err = w.Write([]byte("success login"))

		if err != nil {
			logger.Error("response writer", slog.String("error", err.Error()))
		}

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

	logger.Info("success login handler")

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

	_, err = w.Write(data)

	if err != nil {
		logger.Error("response writer", slog.String("error", err.Error()))

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
}
