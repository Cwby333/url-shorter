package apprunner

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Cwby333/url-shorter/internal/config"
	"github.com/Cwby333/url-shorter/internal/logger"
	"github.com/Cwby333/url-shorter/internal/repository/postgres"
	"github.com/Cwby333/url-shorter/internal/repository/redis"
	"github.com/Cwby333/url-shorter/internal/services/urlsservice"
	"github.com/Cwby333/url-shorter/internal/services/usersservice"
	"github.com/Cwby333/url-shorter/internal/transport/http/ratelimiter"
	"github.com/Cwby333/url-shorter/internal/transport/http/server"

	"golang.org/x/sync/errgroup"
)

const (
	dev   = "dev"
	local = "local"
)

type App struct {
}

func New() App {
	return App{}
}

func (app App) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)

	group, groupCtx := errgroup.WithContext(ctx)

	go func() {
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
		<-signalChan
		cancel()
	}()

	cfg, err := config.Load(dev)

	if err != nil {
		log.Printf("config error: %v", err.Error())
		return
	}

	logger := logger.New(cfg.Env)
	slog.SetDefault(logger.Logger)

	pool, err := postgres.Connect(ctx, cfg.Database)

	if err != nil {
		logger.Error("database connect", slog.String("error", err.Error()))
		return
	}
	defer func() {
		logger.Info("close db connect")
		pool.Close()
	}()

	client, err := myredis.New(ctx, cfg.Redis)

	if err != nil {
		logger.Error("", slog.String("error", err.Error()))
		return
	}
	defer func() {
		logger.Info("close cache connect")
		client.Close()
	}()

	urlService, err := urlsservice.New(pool, client, logger)

	if err != nil {
		logger.Error("", slog.String("error", err.Error()))
		return
	}

	userService, err := usersservice.New(pool, client, logger, cfg.JWT)

	if err != nil {
		logger.Error("", slog.String("error", err.Error()))
		return
	}

	rateLimiter, err := ratelimiter.NewLimiter(cfg.RateLimiter.Limit, cfg.RateLimiter.TTL, ctx)

	if err != nil {
		logger.Error("setup ratelimiter", slog.String("error", err.Error()))
		return
	}
	defer func() {
		logger.Info("ratelimiter shutdown")
		rateLimiter.Shutdown()
	}()

	server, err := httpserver.New(ctx, cfg.HTTPServer, urlService, logger, userService, rateLimiter)

	if err != nil {
		logger.Error("server init", slog.String("error", err.Error()))
		return
	}

	group.Go(func() error {
		logger.Info("server start")
		return server.Server.ListenAndServe()
	})

	group.Go(func() error {
		<-groupCtx.Done()
		logger.Info("start shutdown")

		ctxShutdown, canc := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer func() {
			select {
			case <-ctxShutdown.Done():
			default:
				canc()
			}
		}()

		return server.Server.Shutdown(ctxShutdown)
	})

	if err := group.Wait(); err != nil {
		logger.Info("exit", slog.String("error", err.Error()))
	}
}
