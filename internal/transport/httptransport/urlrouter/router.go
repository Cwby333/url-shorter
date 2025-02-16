package urlrouter

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Cwby333/url-shorter/internal/config"
	"github.com/Cwby333/url-shorter/internal/logger"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/middlewares/basicauth"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/middlewares/logging"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/middlewares/requestid"
)

type URLService interface {
	SaveAlias(ctx context.Context, url, alias string) (int, error)
	GetURL(ctx context.Context, alias string) (string, error)
	DeleteURL(ctx context.Context, alias string) error
	UpdateURL(ctx context.Context, alias, newURL string) error
}

type Router struct {
	urlService URLService
	logger     logger.Logger
	Router     *http.ServeMux
	config.Owner
}

func New(service URLService, logger logger.Logger, owner config.Owner) (*Router, error) {
	const op = "internal/transport/httptransport/urlrouter/New"

	if service == (URLService)(nil) {
		logger.Error("nil pointer in URLService interface", slog.String("op", op))

		return nil, ErrNilPointerInInterface
	}

	return &Router{
		urlService: service,
		logger:     logger,
		Router:     http.NewServeMux(),
		Owner:      owner,
	}, nil
}

func (router *Router) Run() {
	router.Router.Handle("POST /create", requestid.New(router.logger.Logger)(logging.New(basicauth.New(router.Username, router.Password)(http.HandlerFunc(router.Save)))))

	router.Router.Handle("GET /get", requestid.New(router.logger.Logger)(logging.New(http.HandlerFunc(router.Get))))

	router.Router.Handle("DELETE /delete", requestid.New(router.logger.Logger)(basicauth.New(router.Username, router.Password)(logging.New(http.HandlerFunc(router.Delete)))))

	router.Router.Handle("PUT /update", requestid.New(router.logger.Logger)(basicauth.New(router.Username, router.Password)(logging.New(http.HandlerFunc(router.UpdateURL)))))
}
