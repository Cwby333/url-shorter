package ratelimiter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Cwby333/url-shorter/internal/generalerrors"
)

type Limiter struct {
	mu *sync.Mutex

	ttl     time.Duration
	limit   int
	mainCtx context.Context
	storage map[string]Client
}

type Client struct {
	mu *sync.Mutex

	limit          int
	ttl            time.Duration
	ctx            context.Context
	limiter        chan struct{}
	renewalStarted bool
}

func NewLimiter(limit int, ttl time.Duration, mainCtx context.Context) (Limiter, error) {
	const op = "internal/transport/httpsrv/ratelimiter/limiter.go/NewLimiter"

	if limit < 0 {
		return Limiter{}, fmt.Errorf("%s: %w", op, generalerrors.ErrNegativeLimit)
	}

	return Limiter{
		mu:      &sync.Mutex{},
		ttl:     ttl,
		mainCtx: mainCtx,
		limit:   limit,
		storage: make(map[string]Client),
	}, nil
}

func (limiter Limiter) Close() chan error {
	for _, client := range limiter.storage {
		close(client.limiter)
	}

	ch := make(chan error, 1)
	ch <- nil
	return ch
}

func (limiter Limiter) ContextInfo() string {
	return "ratelimiter"
}

func (limiter Limiter) newClient(ip string) {
	client := Client{
		mu:             &sync.Mutex{},
		limit:          limiter.limit,
		ttl:            limiter.ttl,
		ctx:            limiter.mainCtx,
		limiter:        make(chan struct{}, limiter.limit),
		renewalStarted: false,
	}

	limiter.storage[ip] = client
}

func (client Client) startRenewal() {
	client.mu.Lock()
	if client.renewalStarted == true {
		client.mu.Unlock()
		return
	}
	client.renewalStarted = true
	client.mu.Unlock()

	go func() {
		for {
			select {
			case <-client.ctx.Done():
				return
			default:
			}

			select {
			case <-time.After(client.ttl):
				for range client.limit {
					_ = <-client.limiter
				}
			case <-client.ctx.Done():
				return
			}
		}
	}()
}

func (limiter Limiter) Iterate(ip string) error {
	const op = "internal/transport/httpsrv/ratelimiter/Iterate"

	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	client, ok := limiter.storage[ip]

	if !ok {
		limiter.newClient(ip)
		client = limiter.storage[ip]
	}

	client.startRenewal()

	select {
	case client.limiter <- struct{}{}:
		return nil
	default:
		return fmt.Errorf("%s: %w", op, generalerrors.ErrRateLimiterForbidden)
	}
}
