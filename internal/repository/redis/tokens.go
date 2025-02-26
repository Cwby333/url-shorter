package myredis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Cwby333/url-shorter/pkg/generalerrors"
	"github.com/redis/go-redis/v9"
)

const (
	TokenInBlackList = 999
)

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
