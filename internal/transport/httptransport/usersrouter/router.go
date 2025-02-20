package usersrouter

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Cwby333/url-shorter/internal/transport/httptransport/middlewares/logging"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/middlewares/requestid"
	"github.com/Cwby333/url-shorter/pkg/generalerrors"
)

type UsersService interface {
	CreateUser(ctx context.Context, username string, password string) (uuid string, err error)
	LogIn(ctx context.Context, username string, password string) (string, time.Time, error)
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
	router.Router.Handle("POST /create", requestid.New(router.logger)(logging.New(http.HandlerFunc(router.RegisterHandler))))

	router.Router.Handle("POST /login", requestid.New(router.logger)(logging.New(http.HandlerFunc(router.LogInHandler))))
}
