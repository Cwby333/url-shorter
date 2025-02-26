package usersservice

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Cwby333/url-shorter/internal/entity/tokens"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func (service UserService) CreateJWT(ctx context.Context, subject string) (accessClaims tokens.JWTAccessClaims, refreshClaims tokens.JWTRefreshClaims, err error) {
	const op = "internal/services/usersservice/createJWT"

	user, err := service.GetUserByUUID(ctx, subject)
	if err != nil {
		return tokens.JWTAccessClaims{}, tokens.JWTRefreshClaims{}, fmt.Errorf("%s: %w", op, err)
	}

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
		Type:    "refresh",
		Version: user.Version,
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

func (service UserService) CheckCountOfUsesRefreshToken(ctx context.Context, tokenID string, ttl time.Duration) error {
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

func (service UserService) CheckBlacklist(ctx context.Context, tokenID string) error {
	const op = "internal/services/userservice/CheckRefresh"

	err := service.invalidator.CheckBlacklist(ctx, tokenID)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
