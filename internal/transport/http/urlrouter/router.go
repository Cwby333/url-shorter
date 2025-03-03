package urlrouter

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/Cwby333/url-shorter/internal/generalerrors"
	"github.com/Cwby333/url-shorter/internal/logger"
	"github.com/Cwby333/url-shorter/internal/transport/http/middlewares/jwtmiddle"
	"github.com/Cwby333/url-shorter/internal/transport/http/middlewares/limitermidde"
	"github.com/Cwby333/url-shorter/internal/transport/http/middlewares/logging"
	"github.com/Cwby333/url-shorter/internal/transport/http/middlewares/recovermiddle"
	"github.com/Cwby333/url-shorter/internal/transport/http/middlewares/requestid"
	"github.com/Cwby333/url-shorter/internal/transport/http/ratelimiter"
	"github.com/Cwby333/url-shorter/internal/transport/http/urlrouter/popaliases"
	"github.com/go-playground/validator/v10"
)

const (
	defaultTimeSendPopAlias = time.Duration(time.Second * 20)
)

type URLService interface {
	SaveAlias(ctx context.Context, url, alias string) (int, error)
	GetURL(ctx context.Context, alias string) (string, error)
	DeleteURL(ctx context.Context, alias string) error
	UpdateURL(ctx context.Context, alias, newURL string) error
	SendPopAlias(ctx context.Context, alias string, countOfReq int) error
}

type Router struct {
	mainCtx context.Context

	mu         *sync.RWMutex
	urlService URLService
	limiter    ratelimiter.Limiter
	logger     logger.Logger
	Router     *http.ServeMux
	validator  *validator.Validate

	sliceForRandAlias []rune

	popAlias popaliases.PopAlias
}

func New(service URLService, logger logger.Logger, limiter ratelimiter.Limiter, mainCtx context.Context) (*Router, error) {
	const op = "internal/transport/httptransport/urlrouter/New"

	if service == (URLService)(nil) {
		logger.Error("nil pointer in URLService interface", slog.String("op", op))

		return nil, generalerrors.ErrNilPointerInInterface
	}

	data := []rune("QWERTYUIOPASDFGHJKLZXCVBNMqwertyuiopasdfghjklzxcvbnm1234567890")

	return &Router{
		mainCtx:           mainCtx,
		mu:                &sync.RWMutex{},
		urlService:        service,
		limiter:           limiter,
		logger:            logger,
		sliceForRandAlias: data,
		validator:         validator.New(validator.WithRequiredStructEnabled()),
		Router:            http.NewServeMux(),
		popAlias:          popaliases.New(defaultTimeSendPopAlias),
	}, nil
}

func (router *Router) Run() {
	router.Router.Handle("POST /create", recovermiddle.New(requestid.New(router.logger.Logger)(logging.New(jwtmiddle.NewAccess(limitermidde.New(router.limiter)(http.HandlerFunc(router.Save)))))))

	router.Router.Handle("GET /get", recovermiddle.New(requestid.New(router.logger.Logger)(logging.New(limitermidde.New(router.limiter)(http.HandlerFunc(router.Get))))))

	router.Router.Handle("DELETE /delete", recovermiddle.New(requestid.New(router.logger.Logger)(logging.New(jwtmiddle.NewAccess(limitermidde.New(router.limiter)(http.HandlerFunc(router.Delete)))))))

	router.Router.Handle("PUT /update", recovermiddle.New(requestid.New(router.logger.Logger)(logging.New(jwtmiddle.NewAccess(limitermidde.New(router.limiter)(http.HandlerFunc(router.UpdateURL)))))))

	router.StartProcessPopAlias()
}
