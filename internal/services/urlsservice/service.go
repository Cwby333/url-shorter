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
	UpdateURL(ctx context.Context, newURL, alias string) error
}

type URLService struct {
	urlRepository URLRepository
	logger        logger.Logger
}

func New(repo URLRepository, logger logger.Logger) (URLService, error) {
	const op = "internal/services/urlservice/New"

	if repo == (URLRepository)(nil) {
		logger.Error("nil pointer in interface URLRepository", slog.String("op", op))

		return URLService{}, ErrNilPointerInInterface
	}

	return URLService{
		urlRepository: repo,
		logger:        logger,
	}, nil
}

func (service URLService) SaveAlias(ctx context.Context, url, alias string) (int, error) {
	const op = "internal/services/urlservice/SaveAlias"

	res, err := service.urlRepository.SaveAlias(ctx, url, alias)

	if err != nil {
		return res, fmt.Errorf("%s: %w", op, err)
	}

	return res, nil
}

func (service URLService) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "internal/services/urlservice/GetURL"

	res, err := service.urlRepository.GetURL(ctx, alias)

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return res, nil
}

func (service URLService) DeleteURL(ctx context.Context, alias string) error {
	const op = "internal/services/urlservice/GetURL"

	err := service.urlRepository.DeleteURL(ctx, alias)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (service URLService) UpdateURL(ctx context.Context, newURL, alias string) error {
	const op = "internal/services/urlservice/UpdateURL"

	err := service.urlRepository.UpdateURL(ctx, newURL, alias)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
