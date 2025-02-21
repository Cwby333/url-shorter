package myredis

import (
	"context"
	"fmt"
	"time"

	"github.com/Cwby333/url-shorter/internal/config"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func New(cfg config.Redis) Redis {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + fmt.Sprintf("%d", cfg.Port),
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return Redis{
		client: client,
	}
}

func (redis Redis) AddRefreshToBL(ctx context.Context, tokenID string, ttl time.Time)
