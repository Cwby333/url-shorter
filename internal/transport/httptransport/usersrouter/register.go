package usersrouter

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	storageErrors "github.com/Cwby333/url-shorter/internal/repository/errors"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/urlrouter/lib/mainresponse"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/urlrouter/lib/respforusers"
)

type CreateResponse struct {
	mainresponse.Response
	UUID     string `json:"uuid"`
	Username string `json:"username"`
}

func (router Router) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	const op = "internal/transport/httptransport/usersrouter/CreateHandler"
	logger := r.Context().Value("logger").(*slog.Logger)

	username, pass, ok := r.BasicAuth()

	if !ok {
		logger.Info("missing auth header")

		resp, err := newCreateResponse(errors.New("unauthorized"))

		if err != nil {
			logger.Error("json marshal", slog.String("error", err.Error()))

			http.Error(w, "unauthorized", http.StatusUnauthorized)

			return
		}

		http.Error(w, string(resp), http.StatusUnauthorized)

		return
	}

	if username == "" || pass == "" {
		logger.Info("wrong credentials")

		resp, err := newCreateResponse(errors.New("empty username or password"))

		if err != nil {
			logger.Error("json marshal", slog.String("error", err.Error()))

			http.Error(w, "empty username or password", http.StatusUnauthorized)

			return
		}

		http.Error(w, string(resp), http.StatusUnauthorized)

		return
	}

	uuid, err := router.service.CreateUser(r.Context(), username, pass)

	if err != nil {
		if errors.Is(err, storageErrors.ErrUsernameAlreadyExists) {
			logger.Info("username already exists", slog.String("username", username))

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

	response := CreateResponse{
		Response: mainresponse.NewOK(),
		UUID:     uuid,
		Username: username,
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
	response := CreateResponse{
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
