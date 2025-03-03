package urlrouter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/Cwby333/url-shorter/internal/generalerrors"
	"github.com/Cwby333/url-shorter/internal/transport/http/lib/mainresponse"
	"github.com/Cwby333/url-shorter/internal/transport/http/lib/respforusers"
	"github.com/Cwby333/url-shorter/internal/transport/http/lib/typeasserterror"
	validaterequests "github.com/Cwby333/url-shorter/internal/transport/http/lib/validaterequsts"

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

	err := typeasserterror.Check(ok, w, slog.Default())

	if err != nil {
		return
	}

	logger = logger.With("component", "get handler")

	req := RequestGet{}
	err = json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		if !errors.Is(err, io.EOF) {
			out, err := newResponseGet(errors.New("internal error"))

			if err != nil {
				logger.Error("json marshall", slog.String("error", err.Error()))

				http.Error(w, respforusers.ErrInternalError, http.StatusInternalServerError)
				return
			}

			http.Error(w, string(out), http.StatusInternalServerError)
			return
		}
	}

	r.Body.Close()

	err = router.validator.Struct(req)

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

	logger.Info("success handle request")

	router.popAlias.Inc(req.Alias)

	_, err = w.Write(responseJSON)

	if err != nil {
		logger.Error("response write", slog.String("error", err.Error()))

		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
