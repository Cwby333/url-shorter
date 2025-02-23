package usersrouter

import (
	"log/slog"
	"net/http"
)

func (router Router) LogOut(w http.ResponseWriter, r *http.Request) {
	logger := r.Context().Value("logger").(*slog.Logger)	
	logger = logger.With("component", "logout handler")


}	