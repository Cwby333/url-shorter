package urlsservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Cwby333/url-shorter/internal/logger"
)

var (
	ErrNilPointerInInterface = errors.New("nil pointer in interface")
)

type URLRepository interface {
	SaveAlias(ctx context.Context, url, alias string) (int, error)
	GetURL(ctx context.Context, alias string) (string, error)
	DeleteURL(ctx context.Context, alias string) error
}

type UrlService struct {
	urlRepository URLRepository
	logger        logger.Logger
}

func New(repo URLRepository, logger logger.Logger) (UrlService, error) {
	const path = "internal/services/urlservice/New"

	if repo == (URLRepository)(nil) {
		logger.Error("nil pointer in interface URLRepository", slog.String("path", path))

		return UrlService{}, ErrNilPointerInInterface
	}

	return UrlService{
		urlRepository: repo,
		logger:        logger,
	}, nil
}

func (service UrlService) SaveAlias(ctx context.Context, url, alias string) (int, error) {
	const path = "internal/services/urlservice/SaveAlias"

	res, err := service.urlRepository.SaveAlias(ctx, url, alias)

	if err != nil {
		return res, fmt.Errorf("%s:%w", path, err)
	}

	return res, nil
}

func (service UrlService) GetURL(ctx context.Context, alias string) (string, error) {
	const path = "internal/services/urlservice/GetURL"

	res, err := service.urlRepository.GetURL(ctx, alias)

	if err != nil {
		return "", fmt.Errorf("%s:%w", path, err)
	}

	return res, nil
}

func (service UrlService) DeleteURL(ctx context.Context, alias string) error {
	const path = "internal/services/urlservice/GetURL"

	err := service.urlRepository.DeleteURL(ctx, alias)

	if err != nil {
		return fmt.Errorf("%s:%w", path, err)
	}

	return nil
}
