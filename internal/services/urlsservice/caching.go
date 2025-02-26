package urlsservice

import (
	"context"
	"fmt"
)

func (service URLService) SaveResponseInCache(ctx context.Context, alias string, response string) error {
	const op = "internal/services/urlsservice/SaveResponseInCache"

	err := service.cache.SaveResponseInCache(ctx, alias, response)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (service URLService) GetResponseFromCache(ctx context.Context, alias string) (string, error) {
	const op = "internal/services/urlsservice/GetResponseFromCache"

	response, err := service.cache.GetResponseFromCache(ctx, alias)

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}

func (service URLService) RemoveResponseFromCache(ctx context.Context, alias string) error {
	const op = "internal/services/urlsservice/RemoveResponseFromCache"

	err := service.cache.RemoveResponseFromCache(ctx, alias)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
