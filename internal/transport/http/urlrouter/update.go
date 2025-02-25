package urlrouter

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Cwby333/url-shorter/internal/transport/http/lib/mainresponse"
	"github.com/Cwby333/url-shorter/internal/transport/http/lib/respforusers"
	validaterequests "github.com/Cwby333/url-shorter/internal/transport/http/lib/validaterequsts"
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

func newUpdateURLResponse(err error) ([]byte, error) {
	const op = "internal/transport/http/urlsrouter/newUpdateURLResponse"

	response := ResponseUpdateURL{
		Response: mainresponse.NewError(err.Error()),
	}

	out, err := json.Marshal(response)

	if err != nil {
		return []byte{}, fmt.Errorf("%s: %w", op, err)
	}

	return out, nil
}

func (router *Router) UpdateURL(w http.ResponseWriter, r *http.Request) {
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

	logger = logger.With("component", "update url handler")

	req := RequestUpdateURL{}
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

	r.Body.Close()

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

		_, err := w.Write([]byte("success"))

		if err != nil {
			logger.Error("response write", slog.String("error", err.Error()))

			resp := ResponseUpdateURL{
				Response: mainresponse.NewError("internal error"),
				URL:      "",
			}
			data, err := json.Marshal(resp)

			if err != nil {
				logger.Error("json marshal", slog.String("error", err.Error()))

				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			http.Error(w, string(data), http.StatusInternalServerError)
			return
		}

		return
	}

	err = router.urlService.SaveResponseInCache(r.Context(), req.Alias, string(data))

	if err != nil {
		logger.Error("cache", slog.String("error", err.Error()))
	}

	_, err = w.Write(data)

	if err != nil {
		logger.Error("response write", slog.String("error", err.Error()))

		resp := ResponseUpdateURL{
			Response: mainresponse.NewError("internal error"),
			URL:      "",
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
