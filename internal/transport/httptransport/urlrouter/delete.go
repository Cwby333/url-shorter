package urlrouter

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Cwby333/url-shorter/internal/transport/httptransport/urlrouter/lib/mainresponse"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/urlrouter/lib/respforusers"
	validaterequests "github.com/Cwby333/url-shorter/internal/transport/httptransport/urlrouter/lib/validaterequsts"
	"github.com/go-playground/validator/v10"
)

type RequestDelete struct {
	Alias string `json:"alias" validate:"required"`
}

type ResponseDelete struct {
	mainresponse.Response
}

func newDeleteResponse(err error) ([]byte, error) {
	response := ResponseDelete{
		Response: mainresponse.NewError(err.Error()),
	}

	out, err := json.Marshal(response)

	if err != nil {
		return []byte{}, err
	}

	return out, nil
}

func (router *Router) Delete(w http.ResponseWriter, r *http.Request) {
	logger := r.Context().Value("logger").(*slog.Logger)

	logger = logger.With("component", "delete handler")

	var req RequestDelete

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		logger.Error("json decoder", slog.String("error", err.Error()))

		out, err := newDeleteResponse(errors.New("internal error"))

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

		response := ResponseDelete{
			Response: mainresponse.NewError(errForResp...),
		}

		data, err := json.Marshal(response)

		if err != nil {
			logger.Error("json marshall error", slog.String("error", err.Error()))

			out, err := newDeleteResponse(errors.New("internal error"))

			if err != nil {
				logger.Error("json marshall", slog.String("error", err.Error()))

				http.Error(w, respforusers.ErrInternalError, http.StatusInternalServerError)

				return
			}

			http.Error(w, string(out), http.StatusInternalServerError)

			return
		}

		logger.Debug("bad request")

		http.Error(w, string(data), http.StatusBadRequest)

		return
	}

	err = router.urlService.DeleteURL(r.Context(), req.Alias)

	if err != nil {
		logger.Error("delete alias handler", slog.String("error", err.Error()))

		out, err := newDeleteResponse(errors.New("internal error"))

		if err != nil {
			logger.Error("json marshall", slog.String("error", err.Error()))

			http.Error(w, respforusers.ErrInternalError, http.StatusInternalServerError)

			return
		}

		http.Error(w, string(out), http.StatusInternalServerError)

		return
	}

	w.Write([]byte("Success deleted"))
}
