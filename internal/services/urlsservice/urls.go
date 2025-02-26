package urlsservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Cwby333/url-shorter/internal/generalerrors"
)

func (service URLService) SaveAlias(ctx context.Context, url, alias string) (int, error) {
	const op = "internal/services/urlservice/SaveAlias"

	res, err := service.repo.SaveAlias(ctx, url, alias)

	if err != nil {
		return res, fmt.Errorf("%s: %w", op, err)
	}

	return res, nil
}

func (service URLService) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "internal/services/urlservice/GetURL"

	res, err := service.cache.GetResponseFromCache(ctx, alias)

	switch err {
	case nil:
		service.logger.Info("take from cache", slog.String("res", res))
		return res, nil
	default:
		if errors.Is(err, generalerrors.ErrCacheMiss) {
			service.logger.Info("cache mis", slog.String("error", err.Error()))
			break
		}

		service.logger.Error("cache", slog.String("error", err.Error()))
	}

	res, err = service.repo.GetURL(ctx, alias)

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	err = service.cache.SaveResponseInCache(ctx, alias, res)

	if err != nil {
		service.logger.Error("cache", slog.String("error", err.Error()))
	}

	return res, nil
}

func (service URLService) DeleteURL(ctx context.Context, alias string) (err error) {
	const op = "internal/services/urlservice/GetURL"

	defer func() {
		if err == nil {
			e := service.cache.RemoveResponseFromCache(ctx, alias)

			if e != nil {
				service.logger.Error("cache", slog.String("error", err.Error()))
			}
		}
	}()

	err = service.repo.DeleteURL(ctx, alias)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (service URLService) UpdateURL(ctx context.Context, newURL, alias string) (err error) {
	const op = "internal/services/urlservice/UpdateURL"

	var url string

	defer func() {
		if err == nil {
			e := service.cache.SaveResponseInCache(ctx, alias, url)

			if e != nil {
				service.logger.Error("cache", slog.String("error", err.Error()))
			}
		}
	}()

	url, err = service.repo.UpdateURL(ctx, newURL, alias)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
