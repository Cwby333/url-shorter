package usersrouter

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Cwby333/url-shorter/internal/transport/httpsrv/urlrouter/lib/mainresponse"
	validaterequests "github.com/Cwby333/url-shorter/internal/transport/httpsrv/urlrouter/lib/validaterequsts"
	"github.com/Cwby333/url-shorter/pkg/generalerrors"
	"github.com/go-playground/validator/v10"
)

type UpdateRequest struct {
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required"`
	NewUsername string `json:"new_username" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

type UpdateResponse struct {
	Response mainresponse.Response
	UUID     string `json:"uuid"`
}

func (router Router) Update(w http.ResponseWriter, r *http.Request) {
	logger := r.Context().Value("logger").(*slog.Logger)
	logger = logger.With("component", "update user handler")

	var req UpdateRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		logger.Error("json decoder", slog.String("error", err.Error()))

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

	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(req)

	if err != nil {
		errorsValidate := err.(validator.ValidationErrors)

		resp := validaterequests.Validate(errorsValidate)
		response := mainresponse.NewError(resp...)
		data, err := json.Marshal(response)

		if err != nil {
			logger.Error("json marshal", slog.String("error", err.Error()))

			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		http.Error(w, string(data), http.StatusBadRequest)
		return
	}

	user, err := router.service.ChangeCredentials(r.Context(), req.Username, req.Password, req.NewUsername, req.NewPassword)

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

		resp := mainresponse.NewError("internal error")
		data, err := json.Marshal(resp)

		if err != nil {
			logger.Error("json marshal", slog.String("error", err.Error()))

			http.Error(w, "internal error", http.StatusUnauthorized)
			return
		}

		http.Error(w, string(data), http.StatusUnauthorized)
		return
	}

	resp := UpdateResponse{
		Response: mainresponse.NewOK(),
		UUID:     user.UUID,
	}
	data, err := json.Marshal(resp)

	if err != nil {
		logger.Error("json marshal", slog.String("error", err.Error()))

		w.Write([]byte("success update, uuid:" + " " + user.UUID))
		return
	}

	logger.Info("success update handler")

	w.Write(data)
}
