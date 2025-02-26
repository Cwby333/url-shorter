package typeasserterror

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Cwby333/url-shorter/internal/transport/http/lib/mainresponse"
)

func Check(ok bool, w http.ResponseWriter, logger *slog.Logger) error {
	if !ok {
		logger.Error("wrong type assertion")

		resp := mainresponse.NewError("internal error")
		data, err := json.Marshal(resp)

		if err != nil {
			slog.Error("json marshal", slog.String("error", err.Error()))

			http.Error(w, "internal error", http.StatusInternalServerError)
			return errors.New("type assert, json marshal")
		}

		http.Error(w, string(data), http.StatusInternalServerError)
		return errors.New("type assert")
	}

	return nil
}
