package usersservice

import (
	"context"
	"fmt"
	"time"

	"github.com/Cwby333/url-shorter/internal/config"
	"github.com/Cwby333/url-shorter/internal/entity/users"
	"github.com/Cwby333/url-shorter/internal/logger"
	"github.com/Cwby333/url-shorter/pkg/generalerrors"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UsersRepository interface {
	CreateUser(ctx context.Context, username string, password string) (uuid string, err error)
	GetUserByUUID(ctx context.Context, uuid string) (users.User, error)
	GetUserByUsername(ctx context.Context, username string) (users.User, error)
}

type UserService struct {
	repo   UsersRepository
	jwtCfg config.JWT
}

func New(repo UsersRepository, logger logger.Logger, jwtCfg config.JWT) (UserService, error) {
	const op = "internal/services/userservice/New"

	if repo == (UsersRepository)(nil) {
		logger.Error("nil interface in repo")

		return UserService{}, fmt.Errorf("%s: %w", op, generalerrors.ErrNilPointerInInterface)
	}

	return UserService{
		repo:   repo,
		jwtCfg: jwtCfg,
	}, nil
}

func (service UserService) CreateUser(ctx context.Context, username string, password string) (uuid string, err error) {
	const op = "internal/services/userservice/Create"

	hashPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	uuid, err = service.repo.CreateUser(ctx, username, string(hashPass))

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return uuid, nil
}

func (service UserService) GetUserByUUID(ctx context.Context, uuid string) (users.User, error) {
	const op = "internal/services/userservice/GetByUUID"

	user, err := service.repo.GetUserByUUID(ctx, uuid)

	if err != nil {
		return users.User{}, fmt.Errorf("%s; %w", op, err)
	}

	return user, nil
}

func (service UserService) LogIn(ctx context.Context, username string, password string) (string, time.Time, error) {
	const op = "internal/services/userservice/LogIn"

	user, err := service.repo.GetUserByUsername(ctx, username)

	if err != nil {
		return "", time.Time{}, fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	if err != nil {
		return "'", time.Time{}, fmt.Errorf("%s: %w", op, generalerrors.ErrWrongPassword)
	}

	token, expired, err := service.createJWT(ctx, user.UUID)

	if err != nil {
		return "", time.Time{}, fmt.Errorf("%s: %w", op, err)
	}

	return token, expired, nil
}

func (service UserService) createJWT(ctx context.Context, subject string) (string, time.Time, error) {
	const op = "internal/services/usersservice/createJWT"

	select {
	case <-ctx.Done():
		return "", time.Time{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	dur, err := time.ParseDuration(service.jwtCfg.JWTAccess.ExpiredTime)

	if err != nil {
		return "", time.Time{}, fmt.Errorf("%s: %w", op, err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    service.jwtCfg.Issuer,
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(dur)),
		NotBefore: jwt.NewNumericDate(time.Now()),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	})

	s, err := token.SignedString([]byte(service.jwtCfg.SecretKey))

	if err != nil {
		return "", time.Time{}, fmt.Errorf("%s: %w", op, err)
	}

	claims := token.Claims.(jwt.RegisteredClaims)

	return s, claims.ExpiresAt.Time, nil
}
