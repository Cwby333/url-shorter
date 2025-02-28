package httpserver

import (
	"context"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"

	"github.com/Cwby333/url-shorter/internal/config"
	"github.com/Cwby333/url-shorter/internal/logger"
	"github.com/Cwby333/url-shorter/internal/transport/http/ratelimiter"
	"github.com/Cwby333/url-shorter/internal/transport/http/registerrouters"
	"github.com/Cwby333/url-shorter/internal/transport/http/urlrouter"
	"github.com/Cwby333/url-shorter/internal/transport/http/usersrouter"
)

type Server struct {
	Server *http.Server
}

func New(ctx context.Context, cfg config.HTTPServer, urlService urlrouter.URLService, logger logger.Logger, userService usersrouter.UsersService, limiter ratelimiter.Limiter) (Server, error) {
	const op = "transport/http/httpserver/New"

	mux, err := registerrouters.New(urlService, logger, userService, limiter)

	if err != nil {
		return Server{}, fmt.Errorf("%s:%w", op, err)
	}

	server := Server{
		Server: &http.Server{
			Addr:         cfg.Address,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
			Handler:      mux,
			BaseContext: func(_ net.Listener) context.Context {
				return ctx
			},
		},
	}

	return server, nil
}
