package usersrouter

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Cwby333/url-shorter/internal/entity/tokens"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/middlewares/jwtmiddle"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/middlewares/logging"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/middlewares/requestid"
	"github.com/Cwby333/url-shorter/pkg/generalerrors"
)

type UsersService interface {
	CreateUser(ctx context.Context, username string, password string) (uuid string, err error)
	LogIn(ctx context.Context, username string, password string) (accessClaims tokens.JWTAccessClaims, refreshClaims tokens.JWTRefreshClaims, err error)
	LogOut(ctx context.Context, tokenId string, ttl time.Duration) error
	CreateJWT(ctx context.Context, subject string) (accessClaims tokens.JWTAccessClaims, refreshClaims tokens.JWTRefreshClaims, err error)
	CheckBlacklist(ctx context.Context, tokenID string) error
	CheckCountOfUses(ctx context.Context, tokenID string, ttl time.Duration) error
	UseRefresh(ctx context.Context, tokenID string) error
}

type Router struct {
	Router  *http.ServeMux
	service UsersService
	logger  *slog.Logger
}

func New(service UsersService, logger *slog.Logger) (Router, error) {
	const op = "internal/transport/httptransport/usersrouter/New"

	if service == (UsersService)(nil) {
		return Router{}, fmt.Errorf("%s: %w", op, generalerrors.ErrNilPointerInInterface)
	}

	return Router{
		Router:  http.NewServeMux(),
		service: service,
		logger:  logger,
	}, nil
}

func (router Router) Run() {
	router.Router.Handle("POST /create", requestid.New(router.logger)(logging.New(http.HandlerFunc(router.Register))))

	router.Router.Handle("POST /login", requestid.New(router.logger)(logging.New(http.HandlerFunc(router.Login))))

	router.Router.Handle("POST /logout", requestid.New(router.logger)(logging.New(jwtmiddle.NewRefresh(http.HandlerFunc(router.Logout)))))

	router.Router.Handle("POST /refresh", requestid.New(router.logger)(logging.New(jwtmiddle.NewRefresh(http.HandlerFunc(router.RefreshTokens)))))
}
