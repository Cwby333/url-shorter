package myredis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Cwby333/url-shorter/internal/config"
	"github.com/Cwby333/url-shorter/pkg/generalerrors"
	"github.com/redis/go-redis/v9"
)

const (
	TokenInBlackList = 999
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

func (r Redis) Close() {
	r.client.Close()
}

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

func (r Redis) InvalidRefresh(ctx context.Context, tokenID string, ttl time.Duration) error {
	const op = "internal/repository/redis/InvalidRefresh"

	res := r.client.HSet(ctx, "refresh", tokenID, TokenInBlackList)

	if res.Err() != nil {
		return fmt.Errorf("%s: %w", op, res.Err())
	}

	select {
	case <-ctx.Done():
		res := r.client.HExpire(context.Background(), "refresh", ttl, tokenID)

		if res.Err() != nil {
			return fmt.Errorf("%s: %w", op, res.Err())
		}
	default:
		res := r.client.HExpire(context.Background(), "refresh", ttl, tokenID)

		if res.Err() != nil {
			return fmt.Errorf("%s: %w", op, res.Err())
		}
	}

	return nil
}

func (r Redis) CheckCountOfUses(ctx context.Context, tokenID string, ttl time.Duration) error {
	const op = "internal/repository/redis/CheckCountOfUses"

	resStringCmd := r.client.HGet(ctx, "refresh", tokenID)

	if resStringCmd.Err() != nil {
		if errors.Is(resStringCmd.Err(), redis.Nil) {
			res := r.client.HSet(ctx, "refresh", tokenID, 1)

			if res.Err() != nil {
				return fmt.Errorf("%s: %w", op, res.Err())
			}

			select {
			case <-ctx.Done():
				res := r.client.HExpire(context.Background(), "refresh", ttl, tokenID)

				if res.Err() != nil {
					return fmt.Errorf("%s: %w", op, res.Err())
				}
			default:
				res := r.client.HExpire(context.Background(), "refresh", ttl, tokenID)

				if res.Err() != nil {
					return fmt.Errorf("%s: %w", op, res.Err())
				}
			}

			return nil
		}

		return fmt.Errorf("%s: %w", op, resStringCmd.Err())
	}

	str, err := resStringCmd.Result()

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	countOfUses, err := strconv.Atoi(str)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if countOfUses > 1 {
		return fmt.Errorf("%s: %w", op, generalerrors.ErrToManyUseOfRefreshToken)
	}

	return nil
}

func (r Redis) CheckBlacklist(ctx context.Context, tokenID string) error {
	const op = "internal/repository/redis/CheckBlacklist"

	res := r.client.HGet(ctx, "refresh", tokenID)

	if res.Err() != nil {
		if errors.Is(res.Err(), redis.Nil) {
			return nil
		}

		return fmt.Errorf("%s: %w", op, res.Err())
	}

	str, err := res.Result()

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	tokenInBlackListCheck, err := strconv.Atoi(str)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if tokenInBlackListCheck == TokenInBlackList {
		return generalerrors.ErrRefreshInBlackList
	}

	return nil
}

func (r Redis) UseRefresh(ctx context.Context, tokenID string) error {
	const op = "internal/repository/redis/UseRefresh"
	resStringCmd := r.client.HGet(ctx, "refresh", tokenID)

	if resStringCmd.Err() != nil {
		return fmt.Errorf("%s: %w", op, resStringCmd.Err())
	}

	str, err := resStringCmd.Result()

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	countOfUses, err := strconv.Atoi(str)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	resIntCmd := r.client.HSet(ctx, "refresh", tokenID, countOfUses+1)

	if resIntCmd.Err() != nil {
		return fmt.Errorf("%s: %w", op, resIntCmd.Err())
	}

	return nil
}
