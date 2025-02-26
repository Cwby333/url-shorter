package urlsservice

import (
	"context"
	"fmt"
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

	res, err := service.repo.GetURL(ctx, alias)

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return res, nil
}

func (service URLService) DeleteURL(ctx context.Context, alias string) error {
	const op = "internal/services/urlservice/GetURL"

	err := service.repo.DeleteURL(ctx, alias)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (service URLService) UpdateURL(ctx context.Context, newURL, alias string) error {
	const op = "internal/services/urlservice/UpdateURL"

	err := service.repo.UpdateURL(ctx, newURL, alias)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
