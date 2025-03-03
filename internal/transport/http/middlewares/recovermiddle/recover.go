package recovermiddle

import (
	"log/slog"
	"net/http"
)

func New(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := recover()
		if rec != nil {
			slog.Error("panic", slog.Any("recover", rec))
		}

		next.ServeHTTP(w, r)
	})
}
