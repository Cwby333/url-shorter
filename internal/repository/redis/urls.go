package myredis

import (
	"context"
	"errors"
	"fmt"

	"github.com/Cwby333/url-shorter/internal/generalerrors"
	"github.com/redis/go-redis/v9"
)

func (r Redis) SaveResponseInCache(ctx context.Context, alias string, response string) error {
	const op = "internal/repository/redis/SaveResponse"

	res := r.client.HSet(ctx, "urls", alias, response)

	if res.Err() != nil {
		return fmt.Errorf("%s; %w", op, res.Err())
	}

	select {
	case <-ctx.Done():
		res2 := r.client.HExpire(context.Background(), "urls", r.urlTTL, alias)

		if res2.Err() != nil {
			return fmt.Errorf("%s: %w", op, res2.Err())
		}
	default:
		res2 := r.client.HExpire(ctx, "urls", r.urlTTL, alias)

		if res2.Err() != nil {
			return fmt.Errorf("%s: %w", op, res2.Err())
		}
	}

	return nil
}

func (r Redis) GetResponseFromCache(ctx context.Context, alias string) (string, error) {
	const op = "internal/repository/redis/GetResponse"

	res := r.client.HGet(ctx, "urls", alias)

	if res.Err() != nil {
		if errors.Is(res.Err(), redis.Nil) {
			return "", fmt.Errorf("%s: %w", op, generalerrors.ErrCacheMiss)
		}

		return "", fmt.Errorf("%s: %w", op, res.Err())
	}

	response, err := res.Result()

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}

func (r Redis) RemoveResponseFromCache(ctx context.Context, alias string) error {
	const op = "internal/repository/redis/RemoveFromCache"

	res := r.client.HDel(ctx, "urls", alias)

	if res.Err() != nil {
		return fmt.Errorf("%s: %w", op, res.Err())
	}

	return nil
}
