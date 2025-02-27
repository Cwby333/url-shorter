package usersrouter

import (
	"encoding/json"
	"errors"
	"log/slog"
	"math/rand/v2"
	"net/http"

	"github.com/Cwby333/url-shorter/internal/generalerrors"
	"github.com/Cwby333/url-shorter/internal/transport/http/lib/mainresponse"
	"github.com/Cwby333/url-shorter/internal/transport/http/lib/respforusers"
	"github.com/Cwby333/url-shorter/internal/transport/http/lib/typeasserterror"
)

const (
	defaultUsernameLength = 8
	defaultPasswordLength = 8
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	mainresponse.Response
	UUID     string `json:"uuid,omitempty"`
	Username string `json:"username,omitempty"`
}

func (router Router) Register(w http.ResponseWriter, r *http.Request) {
	logger, ok := r.Context().Value("logger").(*slog.Logger)

	err := typeasserterror.Check(ok, w, slog.Default())

	if err != nil {
		return
	}

	logger = logger.With("component", "register handler")

	request := RegisterRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)

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

	r.Body.Close()

	if request.Username == "" {
		out := make([]rune, 0, defaultUsernameLength)

		for range defaultUsernameLength {
			out = append(out, router.dataRandomUsernamePassword[rand.IntN(len(router.dataRandomUsernamePassword))])
		}

		request.Username = string(out)
	}
	if len(request.Password) < 8 {
		logger.Info("to small password")

		response, err := newCreateResponse(errors.New("len password smaller than 8"))

		if err != nil {
			logger.Error("json marshal", slog.String("error", err.Error()))

			http.Error(w, "password smaller than 8", http.StatusBadRequest)
			return
		}

		http.Error(w, string(response), http.StatusBadRequest)
		return
	}

	uuid, err := router.service.CreateUser(r.Context(), request.Username, request.Password)

	if err != nil {
		if errors.Is(err, generalerrors.ErrUsernameAlreadyExists) {
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
	data, err := json.Marshal(response)

	if err != nil {
		logger.Error("json marshal", slog.String("error", err.Error()))

		http.Error(w, "success created", http.StatusOK)
		return
	}

	logger.Info("success create handler")

	_, err = w.Write(data)

	if err != nil {
		logger.Error("response write", slog.String("error", err.Error()))

		resp := RegisterResponse{
			Response: mainresponse.NewError("internal error"),
			UUID:     "",
			Username: "",
		}
		data, err := json.Marshal(resp)

		if err != nil {
			logger.Error("json marshal", slog.String("error", err.Error()))

			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		http.Error(w, string(data), http.StatusInternalServerError)
	}
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
