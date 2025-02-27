package myredis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Cwby333/url-shorter/internal/config"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
	urlTTL time.Duration
}

func New(ctx context.Context, cfg config.Redis) (Redis, error) {
	const op = "internal/repository/redis/New"

	select {
	case <-ctx.Done():
		return Redis{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + strconv.Itoa(cfg.Port),
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
		urlTTL: cfg.URLTTL,
	}, nil
}

func (r Redis) Close() chan error {
	err := r.client.Close()
	ch := make(chan error, 1)
	if err != nil {
		ch <- err
		return ch
	}

	ch <- nil
	return ch
}

func (r Redis) ContextInfo() string {
	return "redis"
}
