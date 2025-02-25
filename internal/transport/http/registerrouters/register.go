package registerrouters

import (
	"fmt"
	"net/http"

	"github.com/Cwby333/url-shorter/internal/logger"
	"github.com/Cwby333/url-shorter/internal/transport/http/ratelimiter"
	"github.com/Cwby333/url-shorter/internal/transport/http/urlrouter"
	"github.com/Cwby333/url-shorter/internal/transport/http/usersrouter"
)

func New(urlService urlrouter.URLService, logger logger.Logger, usersService usersrouter.UsersService, limiter ratelimiter.Limiter) (*http.ServeMux, error) {
	const op = "internal/transports/httptransport/registerrouters/register.go/Register"

	mux := http.NewServeMux()

	routerURLS, err := urlrouter.New(urlService, logger, limiter)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	routerURLS.Run()

	routerUsers, err := usersrouter.New(usersService, logger.Logger, limiter)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	routerUsers.Run()

	mux.Handle("/api/urls/", http.StripPrefix("/api/urls", routerURLS.Router))
	mux.Handle("/api/users/", http.StripPrefix("/api/users", routerUsers.Router))

	return mux, nil
}
