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

type RequestGet struct {
	Alias string `json:"alias" validate:"required"`
}

type ResponseGet struct {
	URL string `json:"url"`
	mainresponse.Response
}

func newResponseGet(err error) ([]byte, error) {
	const op = "internal/transport/http/urlrouter/get.go/newResponseGet"

	response := ResponseGet{
		URL:      "",
		Response: mainresponse.NewError(err.Error()),
	}

	out, err := json.Marshal(response)

	if err != nil {
		return []byte{}, fmt.Errorf("%s: %w", op, err)
	}

	return out, nil
}

func (router *Router) Get(w http.ResponseWriter, r *http.Request) {
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

	logger = logger.With("component", "get handler")

	req := RequestGet{}
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		logger.Error("json decoder", slog.String("error", err.Error()))

		out, err := newResponseGet(errors.New("internal error"))

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

		errorsValidate := err.(validator.ValidationErrors)
		errForResp := validaterequests.Validate(errorsValidate)
		response := ResponseGet{
			URL:      "",
			Response: mainresponse.NewError(errForResp...),
		}
		data, err := json.Marshal(response)

		if err != nil {
			logger.Error("json marshall error", slog.String("error", err.Error()))

			out, err := newSaveResponse(errors.New("bad request"))

			if err != nil {
				logger.Error("json marshall", slog.String("error", err.Error()))

				http.Error(w, "bad request", http.StatusInternalServerError)
				return
			}

			http.Error(w, string(out), http.StatusInternalServerError)
			return
		}

		logger.Debug("bad request")

		http.Error(w, string(data), http.StatusBadRequest)
		return
	}

	response, err := router.urlService.GetResponseFromCache(r.Context(), req.Alias)

	if err != nil {
		if errors.Is(err, generalerrors.ErrCacheMiss) {
			logger.Info("no found in cache", slog.String("alias", req.Alias))
		} else {
			logger.Error("cache", slog.String("error", err.Error()))
		}

		_ = response
	} else {
		logger.Info("success get handler, find from cache")

		_, err = w.Write([]byte(response))

		if err != nil {
			logger.Error("response writer", slog.String("error", err.Error()))

			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		return
	}

	url, err := router.urlService.GetURL(r.Context(), req.Alias)

	if err != nil {
		if errors.Is(err, generalerrors.ErrAliasNotFound) {
			logger.Debug("get url handler", slog.String("error", err.Error()))

			out, err := newSaveResponse(errors.New("alias not found"))

			if err != nil {
				logger.Error("json marshall", slog.String("error", err.Error()))

				http.Error(w, "alias not exist", http.StatusInternalServerError)
				return
			}

			http.Error(w, string(out), http.StatusInternalServerError)
			return
		}

		logger.Error("get url handler error", slog.String("error", err.Error()))

		out, err := newResponseGet(errors.New("internal error"))

		if err != nil {
			logger.Error("json marshall", slog.String("error", err.Error()))

			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		http.Error(w, string(out), http.StatusInternalServerError)
		return
	}

	resp := ResponseGet{
		URL:      url,
		Response: mainresponse.NewOK(),
	}
	responseJSON, err := json.Marshal(resp)

	if err != nil {
		logger.Error("json marshal", slog.String("error", err.Error()))
		http.Error(w, resp.URL, http.StatusInternalServerError)
		return
	}

	err = router.urlService.SaveResponseInCache(r.Context(), req.Alias, string(responseJSON))

	if err != nil {
		logger.Error("cache", slog.String("error", err.Error()))
	}

	logger.Info("success handle request")

	_, err = w.Write(responseJSON)

	if err != nil {
		logger.Error("response write", slog.String("error", err.Error()))

		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
