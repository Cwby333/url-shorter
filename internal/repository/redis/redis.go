package myredis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Cwby333/url-shorter/internal/config"
	"github.com/Cwby333/url-shorter/pkg/generalerrors"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func New(ctx context.Context, cfg config.Redis) (Redis, error) {
	const op = "internal/repository/redis/New"
	
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + fmt.Sprintf("%d", cfg.Port),
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	status := client.Ping(ctx)

	if status.Err() != nil {
		return Redis{}, fmt.Errorf("%s: %w", op, status.Err())
	}

	return Redis{
		client: client,
	}, nil
}

func (r Redis) InvalidRefresh(ctx context.Context, tokenID string, ttl time.Duration)  {		
	r.client.HSet(ctx, "refresh", tokenID, struct{}{})
	r.client.HExpire(ctx, "refresh", ttl, tokenID)
}

func (r Redis) CheckRefresh(ctx context.Context, tokenID string) (error) {	
	const op = "internal/repository/redis/CheckRefresh"

	res := r.client.HGet(ctx, "refresh", tokenID)

	if res.Err() != nil {
		if errors.Is(res.Err(), redis.Nil) {
			return nil
		}

		return fmt.Errorf("%s: %w", op, res.Err())
	}

	return generalerrors.ErrRefreshInBlackList
}