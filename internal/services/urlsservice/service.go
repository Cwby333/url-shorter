package urlsservice

import (
	"context"
	"fmt"

	"github.com/Cwby333/url-shorter/internal/logger"
	"github.com/Cwby333/url-shorter/pkg/generalerrors"
)

type URLRepository interface {
	SaveAlias(ctx context.Context, url, alias string) (int, error)
	GetURL(ctx context.Context, alias string) (string, error)
	DeleteURL(ctx context.Context, alias string) error
	UpdateURL(ctx context.Context, newURL, alias string) error
}

type URLCache interface {
	SaveResponseInCache(ctx context.Context, alias string, response string) error
	GetResponseFromCache(ctx context.Context, alias string) (string, error)
	RemoveResponseFromCache(ctx context.Context, alias string) error
}

type URLService struct {
	repo  URLRepository
	cache URLCache
}

func New(repo URLRepository, cache URLCache, logger logger.Logger) (URLService, error) {
	const op = "internal/services/urlservice/New"

	if repo == (URLRepository)(nil) {
		logger.Error("nil pointer in interface URLRepository")

		return URLService{}, fmt.Errorf("%s: %w", op, generalerrors.ErrNilPointerInInterface)
	}
	if cache == (URLCache)(nil) {
		logger.Error("nil pointer in interface URLCache")

		return URLService{}, fmt.Errorf("%s: %w", op, generalerrors.ErrNilPointerInInterface)
	}

	return URLService{
		repo:  repo,
		cache: cache,
	}, nil
}

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
