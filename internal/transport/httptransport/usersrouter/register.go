package usersrouter

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	storageErrors "github.com/Cwby333/url-shorter/internal/repository/errors"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/urlrouter/lib/mainresponse"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/urlrouter/lib/respforusers"
	validaterequests "github.com/Cwby333/url-shorter/internal/transport/httptransport/urlrouter/lib/validaterequsts"
	"github.com/go-playground/validator/v10"
)

type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RegisterResponse struct {
	mainresponse.Response
	UUID     string `json:"uuid"`
	Username string `json:"username"`
}

func (router Router) Register(w http.ResponseWriter, r *http.Request) {
	logger := r.Context().Value("logger").(*slog.Logger)
	logger = logger.With("component", "register handler")

	var request RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		logger.Error("json decoder", slog.String("error", err.Error()))

		resp, err := newCreateResponse(errors.New("internal error"))

		if err != nil {
			logger.Error("json marshal", slog.String("error", err.Error()))

			http.Error(w, "internal error", http.StatusUnauthorized)
			return
		}

		http.Error(w, string(resp), http.StatusUnauthorized)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	err = validate.Struct(request)

	if err != nil {
		logger.Info("bad request", slog.String("error", err.Error()))

		errorsSlice := err.(validator.ValidationErrors)
		errForResp := validaterequests.Validate(errorsSlice)
		resp := RegisterResponse{
			UUID:     "",
			Username: "",
			Response: mainresponse.NewError(errForResp...),
		}
		data, err := json.Marshal(resp)

		if err != nil {
			logger.Error("json marshall error", slog.String("error", err.Error()))

			out, err := newCreateResponse(errors.New("bad request"))

			if err != nil {
				logger.Error("json marshall", slog.String("error", err.Error()))

				http.Error(w, "bad request", http.StatusBadRequest)

				return
			}

			http.Error(w, string(out), http.StatusBadRequest)

			return
		}

		http.Error(w, string(data), http.StatusBadRequest)
		return
	}

	uuid, err := router.service.CreateUser(r.Context(), request.Username, request.Password)

	if err != nil {

		if errors.Is(err, storageErrors.ErrUsernameAlreadyExists) {
			logger.Info("username already exists", slog.String("username", request.Username))

			resp, err := newCreateResponse(errors.New("username already exists"))

			if err != nil {
				logger.Error("json marshal", slog.String("error", err.Error()))

				http.Error(w, "username already exists", http.StatusUnauthorized)
				return
			}

			http.Error(w, string(resp), http.StatusUnauthorized)
			return
		}

		logger.Error("create user handler", slog.String("error", err.Error()))

		resp, err := newCreateResponse(errors.New("internal error"))

		if err != nil {
			logger.Error("json marshal", slog.String("error", err.Error()))

			http.Error(w, respforusers.ErrInternalError, http.StatusUnauthorized)
			return
		}

		http.Error(w, string(resp), http.StatusUnauthorized)
		return
	}

	response := RegisterResponse{
		Response: mainresponse.NewOK(),
		UUID:     uuid,
		Username: request.Username,
	}

	out, err := json.Marshal(response)

	if err != nil {
		logger.Error("json marshal", slog.String("error", err.Error()))

		http.Error(w, "success created", http.StatusOK)
		return
	}

	logger.Info("success create handler")

	w.Write(out)
}

func newCreateResponse(err error) ([]byte, error) {
	response := RegisterResponse{
		Response: mainresponse.NewError(err.Error()),
		UUID:     "",
		Username: "",
	}

	out, err := json.Marshal(response)

	if err != nil {
		return []byte{}, err
	}

	return out, nil
}
