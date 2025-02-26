package usersservice

import (
	"context"
	"fmt"
	"time"

	"github.com/Cwby333/url-shorter/internal/entity/tokens"
	"github.com/Cwby333/url-shorter/internal/entity/users"
	"github.com/Cwby333/url-shorter/pkg/generalerrors"
	"golang.org/x/crypto/bcrypt"
)

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

func (service UserService) ChangeCredentials(ctx context.Context, username string, password string, newUsername string, newPassword string) (users.User, error) {
	const op = "internal/services/userservice/ChangeCredentials"

	user, err := service.repo.GetUserByUsername(ctx, username)

	if err != nil {
		return users.User{}, fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	if err != nil {
		return users.User{}, fmt.Errorf("%s: %w", op, err)
	}

	newPass, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)

	if err != nil {
		return users.User{}, fmt.Errorf("%s: %w", op, err)
	}

	user, err = service.repo.ChangeCredentials(ctx, newUsername, string(newPass), username)

	if err != nil {
		return users.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (service UserService) BlockUser(ctx context.Context, uuid string) error {
	const op = "internal/services/userservice/BlockUser"

	err := service.repo.BlockUser(ctx, uuid)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
