package httpserver

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/Cwby333/url-shorter/internal/config"
	"github.com/Cwby333/url-shorter/internal/logger"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/registerrouters"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/urlrouter"
)

type Server struct {
	Server *http.Server
	ErrCh  chan error
}

func New(ctx context.Context, cfg config.HTTPServer, urlService urlrouter.URLService, logger logger.Logger, cfgOwner config.Owner, cfgLimiter config.Limiter) (Server, error) {
	const op = "transport/http/httpserver/New"

	mux, err := registerrouters.New(urlService, logger, cfgOwner, cfgLimiter)

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
