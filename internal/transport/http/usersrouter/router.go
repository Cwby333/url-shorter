package usersrouter

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Cwby333/url-shorter/internal/entity/tokens"
	"github.com/Cwby333/url-shorter/internal/entity/users"
	"github.com/Cwby333/url-shorter/internal/generalerrors"
	"github.com/Cwby333/url-shorter/internal/transport/http/middlewares/jwtmiddle"
	"github.com/Cwby333/url-shorter/internal/transport/http/middlewares/limitermidde"
	"github.com/Cwby333/url-shorter/internal/transport/http/middlewares/logging"
	"github.com/Cwby333/url-shorter/internal/transport/http/middlewares/requestid"
	"github.com/Cwby333/url-shorter/internal/transport/http/ratelimiter"
)

type UsersService interface {
	CreateUser(ctx context.Context, username string, password string) (uuid string, err error)

	GetUserByUUID(ctx context.Context, uuid string) (users.User, error)

	LogIn(ctx context.Context, username string, password string) (accessClaims tokens.JWTAccessClaims, refreshClaims tokens.JWTRefreshClaims, err error)

	LogOut(ctx context.Context, tokenID string, ttl time.Duration) error

	CreateJWT(ctx context.Context, subject string) (accessClaims tokens.JWTAccessClaims, refreshClaims tokens.JWTRefreshClaims, err error)

	CheckBlacklist(ctx context.Context, tokenID string) error

	CheckCountOfUsesRefreshToken(ctx context.Context, tokenID string, ttl time.Duration) error

	UseRefresh(ctx context.Context, tokenID string) error

	ChangeCredentials(ctx context.Context, username string, password string, newUsername string, newPassword string) (users.User, error)

	BlockUser(ctx context.Context, uuid string) error
}

type Router struct {
	Router  *http.ServeMux
	service UsersService
	limiter ratelimiter.Limiter
	logger  *slog.Logger
}

func New(service UsersService, logger *slog.Logger, limiter ratelimiter.Limiter) (Router, error) {
	const op = "internal/transport/httptransport/usersrouter/New"

	if service == (UsersService)(nil) {
		return Router{}, fmt.Errorf("%s: %w", op, generalerrors.ErrNilPointerInInterface)
	}

	return Router{
		Router:  http.NewServeMux(),
		service: service,
		limiter: limiter,
		logger:  logger,
	}, nil
}

func (router Router) Run() {
	router.Router.Handle("POST /create", requestid.New(router.logger)(logging.New(limitermidde.New(router.limiter)(http.HandlerFunc(router.Register)))))

	router.Router.Handle("POST /login", requestid.New(router.logger)(logging.New(limitermidde.New(router.limiter)(http.HandlerFunc(router.Login)))))

	router.Router.Handle("POST /logout", requestid.New(router.logger)(logging.New(jwtmiddle.NewRefresh(limitermidde.New(router.limiter)(http.HandlerFunc(router.Logout))))))

	router.Router.Handle("POST /refresh", requestid.New(router.logger)(logging.New(jwtmiddle.NewRefresh(limitermidde.New(router.limiter)(http.HandlerFunc(router.RefreshTokens))))))

	router.Router.Handle("PUT /update", requestid.New(router.logger)(logging.New(limitermidde.New(router.limiter)(http.HandlerFunc(router.Update)))))
}
