package usersservice

import (
	"context"
	"fmt"
	"time"

	"github.com/Cwby333/url-shorter/internal/config"
	"github.com/Cwby333/url-shorter/internal/entity/users"
	"github.com/Cwby333/url-shorter/internal/logger"
	"github.com/Cwby333/url-shorter/pkg/generalerrors"
)

type UsersRepository interface {
	CreateUser(ctx context.Context, username string, password string) (uuid string, err error)
	GetUserByUUID(ctx context.Context, uuid string) (users.User, error)
	GetUserByUsername(ctx context.Context, username string) (users.User, error)
	ChangeCredentials(ctx context.Context, newUsername string, newPassword string, username string) (user users.User, err error)
	BlockUser(ctx context.Context, uuid string) error
}

type RefreshTokenInvalidator interface {
	InvalidRefresh(ctx context.Context, tokenID string, ttl time.Duration) error
	CheckBlacklist(ctx context.Context, tokenID string) error
	CheckCountOfUses(ctx context.Context, tokenID string, ttl time.Duration) error
	UseRefresh(ctx context.Context, tokenID string) error
}

type UserService struct {
	repo        UsersRepository
	invalidator RefreshTokenInvalidator
	jwtCfg      config.JWT
}

func New(repo UsersRepository, invalidator RefreshTokenInvalidator, logger logger.Logger, jwtCfg config.JWT) (UserService, error) {
	const op = "internal/services/userservice/New"

	if repo == (UsersRepository)(nil) {
		logger.Error("nil interface in repo")

		return UserService{}, fmt.Errorf("%s: %w", op, generalerrors.ErrNilPointerInInterface)
	}
	if invalidator == (RefreshTokenInvalidator)(nil) {
		logger.Error("nil interface in repo")

		return UserService{}, fmt.Errorf("%s: %w", op, generalerrors.ErrNilPointerInInterface)
	}

	return UserService{
		repo:        repo,
		invalidator: invalidator,
		jwtCfg:      jwtCfg,
	}, nil
}
