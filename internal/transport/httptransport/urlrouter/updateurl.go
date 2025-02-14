package urlrouter

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	storageErrors "github.com/Cwby333/url-shorter/internal/repository/errors"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/urlrouter/lib/mainresponse"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/urlrouter/lib/respforusers"
)

type RequestUpdateURL struct {
	Alias  string `json:"alias" validate:"required"`
	NewURL string `json:"url" validate:":required"`
}

type ResponseUpdateURL struct {
	mainresponse.Response
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

	err = router.urlService.UpdateURL(r.Context(), req.NewURL, req.Alias)

	if err != nil {
		if errors.Is(err, storageErrors.ErrAliasNotFound) {
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
	}

	data, err := json.Marshal(response)

	if err != nil {
		logger.Error("update url handler", slog.String("error", err.Error()))

		w.Write([]byte("success"))
		w.WriteHeader(http.StatusOK)

		return
	}

	w.Write(data)
	w.WriteHeader(http.StatusOK)
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
