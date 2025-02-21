package urlrouter

import (
	"encoding/json"
	"errors"
	"log/slog"
	"math/rand"
	"net/http"

	storageErrors "github.com/Cwby333/url-shorter/internal/repository/errors"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/urlrouter/lib/mainresponse"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/urlrouter/lib/respforusers"
	validaterequests "github.com/Cwby333/url-shorter/internal/transport/httptransport/urlrouter/lib/validaterequsts"
	"github.com/go-playground/validator/v10"
)

const (
	aliasRandLength = 6
)

var (
	data = []rune("QWERTYUIOPASDFGHJKLZXCVBNMqwertyuiopasdfghjklzxcvbnm1234567890")
)

type RequestSave struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type ResponseSave struct {
	ID    int    `json:"id"`
	Alias string `json:"alias"`
	mainresponse.Response
}

func newSaveResponse(err error) ([]byte, error) {
	response := ResponseSave{
		ID:       -1,
		Response: mainresponse.NewError(err.Error()),
	}

	out, err := json.Marshal(response)

	if err != nil {
		return []byte{}, err
	}

	return out, nil
}

func (router *Router) Save(w http.ResponseWriter, r *http.Request) {
	logger := r.Context().Value("logger").(*slog.Logger)

	logger = logger.With("component", "save handler")

	var req RequestSave

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		logger.Error("json decoder", slog.String("error", err.Error()))

		out, err := newSaveResponse(errors.New("internal error"))

		if err != nil {
			logger.Error("json marshall", slog.String("error", err.Error()))

			http.Error(w, respforusers.ErrInternalError, http.StatusInternalServerError)

			return
		}

		http.Error(w, string(out), http.StatusInternalServerError)

		return
	}
	r.Body.Close()

	validate := validator.New(validator.WithRequiredStructEnabled())

	err = validate.Struct(req)

	if err != nil {
		logger.Info("bad request", slog.String("error", err.Error()))

		errorsValidation := err.(validator.ValidationErrors)

		errForResp := validaterequests.Validate(errorsValidation)

		response := ResponseSave{
			ID:       -1,
			Response: mainresponse.NewError(errForResp...),
		}

		data, err := json.Marshal(response)

		if err != nil {
			logger.Error("json marshall", slog.String("error", err.Error()))

			out, err := newSaveResponse(errors.New(respforusers.ErrInternalError))

			if err != nil {
				logger.Error("json marshall", slog.String("error", err.Error()))

				http.Error(w, "bad request", http.StatusBadRequest)

				return
			}

			http.Error(w, string(out), http.StatusBadRequest)

			return
		}

		logger.Debug("bad request")

		http.Error(w, string(data), http.StatusBadRequest)

		return
	}

	if req.Alias == "" {
		out := make([]rune, 0, aliasRandLength)

		for range aliasRandLength {
			out = append(out, data[rand.Intn(len(data))])
		}

		req.Alias = string(out)
	}

	id, err := router.urlService.SaveAlias(r.Context(), req.URL, req.Alias)

	if err != nil {
		if errors.Is(err, storageErrors.ErrAliasAlreadyExists) {
			logger.Debug("save alias handler", slog.String("error", err.Error()))

			response := ResponseSave{
				ID:       -1,
				Response: mainresponse.NewError(errors.New(storageErrors.ErrAliasAlreadyExists.Error()).Error()),
			}

			out, err := json.Marshal(response)

			if err != nil {
				logger.Error("json marshall", slog.String("error", err.Error()))

				http.Error(w, respforusers.ErrInternalError, http.StatusInternalServerError)

				return
			}

			http.Error(w, string(out), http.StatusInternalServerError)

			return
		}

		logger.Error("save alias handler", slog.String("error", err.Error()))

		out, err := newSaveResponse(errors.New("internal error"))

		if err != nil {
			logger.Error("json marshall", slog.String("error", err.Error()))

			http.Error(w, respforusers.ErrInternalError, http.StatusInternalServerError)

			return
		}

		http.Error(w, string(out), http.StatusInternalServerError)

		return
	}

	resp := ResponseSave{
		ID:       id,
		Response: mainresponse.NewOK(),
		Alias:    req.Alias,
	}

	responseJson, err := json.Marshal(resp)

	if err != nil {
		out, err := newSaveResponse(errors.New("internal error"))

		if err != nil {
			logger.Error("json marshall", slog.String("error", err.Error()))

			http.Error(w, respforusers.ErrInternalError, http.StatusInternalServerError)

			return
		}

		http.Error(w, string(out), http.StatusInternalServerError)

		return
	}

	logger.Info("success handle request")

	w.WriteHeader(http.StatusOK)
	w.Write(responseJson)
}
