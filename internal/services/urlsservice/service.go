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
