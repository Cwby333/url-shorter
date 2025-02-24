package usersservice

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Cwby333/url-shorter/internal/config"
	"github.com/Cwby333/url-shorter/internal/entity/tokens"
	"github.com/Cwby333/url-shorter/internal/entity/users"
	"github.com/Cwby333/url-shorter/internal/logger"
	"github.com/Cwby333/url-shorter/pkg/generalerrors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UsersRepository interface {
	CreateUser(ctx context.Context, username string, password string) (uuid string, err error)
	GetUserByUUID(ctx context.Context, uuid string) (users.User, error)
	GetUserByUsername(ctx context.Context, username string) (users.User, error)
}

type RefreshInvalidator interface {
	InvalidRefresh(ctx context.Context, tokenID string, ttl time.Duration) error
	CheckBlacklist(ctx context.Context, tokenID string) error
	CheckCountOfUses(ctx context.Context, tokenID string, ttl time.Duration) error
	UseRefresh(ctx context.Context, tokenID string) error
}

type UserService struct {
	repo        UsersRepository
	invalidator RefreshInvalidator
	jwtCfg      config.JWT
}

func New(repo UsersRepository, invalidator RefreshInvalidator, logger logger.Logger, jwtCfg config.JWT) (UserService, error) {
	const op = "internal/services/userservice/New"

	if repo == (UsersRepository)(nil) {
		logger.Error("nil interface in repo")

		return UserService{}, fmt.Errorf("%s: %w", op, generalerrors.ErrNilPointerInInterface)
	}
	if invalidator == (RefreshInvalidator)(nil) {
		logger.Error("nil interface in repo")

		return UserService{}, fmt.Errorf("%s: %w", op, generalerrors.ErrNilPointerInInterface)

	}

	return UserService{
		repo:        repo,
		invalidator: invalidator,
		jwtCfg:      jwtCfg,
	}, nil
}

func (service UserService) CreateJWT(ctx context.Context, subject string) (accessClaims tokens.JWTAccessClaims, refreshClaims tokens.JWTRefreshClaims, err error) {
	const op = "internal/services/usersservice/createJWT"

	select {
	case <-ctx.Done():
		return tokens.JWTAccessClaims{}, tokens.JWTRefreshClaims{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	accessDur, err := time.ParseDuration(service.jwtCfg.JWTAccess.ExpiredTime)

	if err != nil {
		return tokens.JWTAccessClaims{}, tokens.JWTRefreshClaims{}, fmt.Errorf("%s: %w", op, err)
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, tokens.JWTAccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    service.jwtCfg.Issuer,
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessDur)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.NewString(),
		},
		Type: "access",
	})
	accessSign, err := accessToken.SignedString([]byte(os.Getenv("APP_JWT_SECRET_KEY")))

	if err != nil {
		return tokens.JWTAccessClaims{}, tokens.JWTRefreshClaims{}, fmt.Errorf("%s: %w", op, err)
	}

	accessClaims = accessToken.Claims.(tokens.JWTAccessClaims)
	accessClaims.Sign = accessSign
	accessClaims.Type = "access"

	refreshDur, err := time.ParseDuration(service.jwtCfg.JWTRefresh.ExpiredTime)

	if err != nil {
		return tokens.JWTAccessClaims{}, tokens.JWTRefreshClaims{}, fmt.Errorf("%s: %w", op, err)
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, tokens.JWTRefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    service.jwtCfg.Issuer,
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshDur)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.NewString(),
		},
		Type: "refresh",
	})
	refreshSign, err := refreshToken.SignedString([]byte(os.Getenv("APP_JWT_SECRET_KEY")))

	if err != nil {
		return tokens.JWTAccessClaims{}, tokens.JWTRefreshClaims{}, fmt.Errorf("%s: %w", op, err)
	}

	refreshClaims = refreshToken.Claims.(tokens.JWTRefreshClaims)
	refreshClaims.Sign = refreshSign
	refreshClaims.Type = "refresh"

	return accessClaims, refreshClaims, nil
}

func (service UserService) LogIn(ctx context.Context, username string, password string) (accessClaims tokens.JWTAccessClaims, refreshClaims tokens.JWTRefreshClaims, err error) {
	const op = "internal/services/userservice/LogIn"

	user, err := service.repo.GetUserByUsername(ctx, username)

	if err != nil {
		return tokens.JWTAccessClaims{}, tokens.JWTRefreshClaims{}, fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	if err != nil {
		return tokens.JWTAccessClaims{}, tokens.JWTRefreshClaims{}, fmt.Errorf("%s: %w", op, generalerrors.ErrWrongPassword)
	}

	accessClaims, refreshClaims, err = service.CreateJWT(ctx, user.UUID)

	if err != nil {
		return tokens.JWTAccessClaims{}, tokens.JWTRefreshClaims{}, fmt.Errorf("%s: %w", op, err)
	}

	return accessClaims, refreshClaims, nil
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

func (service UserService) LogOut(ctx context.Context, tokenID string, ttl time.Duration) error {
	const op = "internal/services/userservice/LogOut"

	select {
	case <-ctx.Done():
		return fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	err := service.invalidator.InvalidRefresh(ctx, tokenID, ttl)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (service UserService) CheckBlacklist(ctx context.Context, tokenID string) error {
	const op = "internal/services/userservice/CheckRefresh"

	err := service.invalidator.CheckBlacklist(ctx, tokenID)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (service UserService) CheckCountOfUses(ctx context.Context, tokenID string, ttl time.Duration) error {
	const op = "internal/services/userservice/CheckCountOfUses"

	err := service.invalidator.CheckCountOfUses(ctx, tokenID, ttl)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (service UserService) UseRefresh(ctx context.Context, tokenID string) error {
	const op = "internal/services/userservice/UseRefresh"

	err := service.invalidator.UseRefresh(ctx, tokenID)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
