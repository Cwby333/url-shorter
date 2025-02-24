package urlrouter

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Cwby333/url-shorter/internal/transport/httpsrv/urlrouter/lib/mainresponse"
	"github.com/Cwby333/url-shorter/internal/transport/httpsrv/urlrouter/lib/respforusers"
	validaterequests "github.com/Cwby333/url-shorter/internal/transport/httpsrv/urlrouter/lib/validaterequsts"
	"github.com/Cwby333/url-shorter/pkg/generalerrors"
	"github.com/go-playground/validator/v10"
)

type RequestUpdateURL struct {
	Alias  string `json:"alias" validate:"required"`
	NewURL string `json:"url" validate:"required"`
}

type ResponseUpdateURL struct {
	mainresponse.Response
	URL string `json:"url"`
}

func (router *Router) UpdateURL(w http.ResponseWriter, r *http.Request) {
	logger := r.Context().Value("logger").(*slog.Logger)

	logger = logger.With("component", "update url handler")

	var req RequestUpdateURL

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		out, err := newUpdateURLResponse(err)

		if err != nil {
			logger.Error("json marshal", slog.String("error", err.Error()))

			http.Error(w, respforusers.ErrInternalError, http.StatusInternalServerError)
			return
		}

		http.Error(w, string(out), http.StatusInternalServerError)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	err = validate.Struct(req)

	if err != nil {
		logger.Info("bad request", slog.String("error", err.Error()))

		errorsValidation := err.(validator.ValidationErrors)

		errForResp := validaterequests.Validate(errorsValidation)

		response := ResponseUpdateURL{
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

	err = router.urlService.UpdateURL(r.Context(), req.NewURL, req.Alias)

	if err != nil {
		if errors.Is(err, generalerrors.ErrAliasNotFound) {
			logger.Debug("update url handler", slog.String("error", err.Error()))

			out, err := newUpdateURLResponse(err)

			if err != nil {
				logger.Error("json marshal", slog.String("error", err.Error()))

				http.Error(w, "alias not found", http.StatusNotFound)
			}

			http.Error(w, string(out), http.StatusNotFound)
			return
		}

		logger.Error("update url handler", slog.String("error", err.Error()))

		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response := ResponseUpdateURL{
		Response: mainresponse.NewOK(),
		URL:      req.NewURL,
	}

	data, err := json.Marshal(response)

	if err != nil {
		logger.Error("update url handler", slog.String("error", err.Error()))

		w.Write([]byte("success"))
		return
	}

	err = router.urlService.SaveResponseInCache(r.Context(), req.Alias, string(data))

	if err != nil {
		logger.Error("cache", slog.String("error", err.Error()))
	}

	w.Write(data)
}

func newUpdateURLResponse(err error) ([]byte, error) {
	response := ResponseUpdateURL{
		Response: mainresponse.NewError(err.Error()),
	}

	out, err := json.Marshal(response)

	if err != nil {
		return []byte{}, err
	}

	return out, nil
}
