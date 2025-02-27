package gracefuler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

const (
	DefaultTimeoutForOneClose = time.Duration(time.Second * 3)
)

type Closer interface {
	Close() chan error
	ContextInfo() string
}

type Gracefuler struct {
	mu     *sync.Mutex
	funcs  []Closer
	logger *slog.Logger
}

func New(logger *slog.Logger) *Gracefuler {
	return &Gracefuler{
		mu:     &sync.Mutex{},
		funcs:  make([]Closer, 0),
		logger: logger,
	}
}

func (c *Gracefuler) Add(closer Closer) {
	c.mu.Lock()
	c.funcs = append(c.funcs, closer)
	c.mu.Unlock()
}

func (c *Gracefuler) startShutdown() {
	c.mu.Lock()
}

func (c *Gracefuler) Close(ctx context.Context) []error {
	c.startShutdown()

	errs := make([]error, 0)
	complete := make(chan struct{})

	go func() {
		for i := len(c.funcs) - 1; i > -1; i-- {
			select {
			case err := <-c.funcs[i].Close():
				msg := c.funcs[i].ContextInfo()
				if err != nil {
					errs = append(errs, fmt.Errorf("%s: %w", msg, err))
				}
			case <-time.After(DefaultTimeoutForOneClose):
				msg := c.funcs[i].ContextInfo()
				errs = append(errs, fmt.Errorf("%s: %w", msg, errors.New("closer denied by timeout")))
			}
		}

		complete <- struct{}{}
	}()

	select {
	case <-complete:
		c.logger.Info("success shutdown")
	case <-ctx.Done():
		c.logger.Info("shutdown by timeout")
	}

	return errs
}
