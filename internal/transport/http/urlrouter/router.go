package urlrouter

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/Cwby333/url-shorter/internal/generalerrors"
	"github.com/Cwby333/url-shorter/internal/logger"
	"github.com/Cwby333/url-shorter/internal/transport/http/middlewares/jwtmiddle"
	"github.com/Cwby333/url-shorter/internal/transport/http/middlewares/limitermidde"
	"github.com/Cwby333/url-shorter/internal/transport/http/middlewares/logging"
	"github.com/Cwby333/url-shorter/internal/transport/http/middlewares/requestid"
	"github.com/Cwby333/url-shorter/internal/transport/http/ratelimiter"
)

type URLService interface {
	SaveAlias(ctx context.Context, url, alias string) (int, error)
	GetURL(ctx context.Context, alias string) (string, error)
	DeleteURL(ctx context.Context, alias string) error
	UpdateURL(ctx context.Context, alias, newURL string) error
}

type Router struct {
	mu         *sync.RWMutex
	urlService URLService
	limiter    ratelimiter.Limiter
	logger     logger.Logger
	Router     *http.ServeMux
}

func New(service URLService, logger logger.Logger, limiter ratelimiter.Limiter) (*Router, error) {
	const op = "internal/transport/httptransport/urlrouter/New"

	if service == (URLService)(nil) {
		logger.Error("nil pointer in URLService interface", slog.String("op", op))

		return nil, generalerrors.ErrNilPointerInInterface
	}

	return &Router{
		mu:         &sync.RWMutex{},
		urlService: service,
		limiter:    limiter,
		logger:     logger,
		Router:     http.NewServeMux(),
	}, nil
}

func (router *Router) Run() {
	router.Router.Handle("POST /create", requestid.New(router.logger.Logger)(logging.New(jwtmiddle.NewAccess(limitermidde.New(router.limiter)(http.HandlerFunc(router.Save))))))
	router.Router.Handle("GET /get", requestid.New(router.logger.Logger)(logging.New(limitermidde.New(router.limiter)(http.HandlerFunc(router.Get)))))
	router.Router.Handle("DELETE /delete", requestid.New(router.logger.Logger)(logging.New(jwtmiddle.NewAccess(limitermidde.New(router.limiter)(http.HandlerFunc(router.Delete))))))
	router.Router.Handle("PUT /update", requestid.New(router.logger.Logger)(logging.New(jwtmiddle.NewAccess(limitermidde.New(router.limiter)(http.HandlerFunc(router.UpdateURL))))))
}
