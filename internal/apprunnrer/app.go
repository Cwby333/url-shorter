package apprunner

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/Cwby333/url-shorter/internal/config"
	"github.com/Cwby333/url-shorter/internal/logger"
	"github.com/Cwby333/url-shorter/internal/repository/postgres"
	"github.com/Cwby333/url-shorter/internal/services/urlsservice"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/httpserver"
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

	cfg := config.Load("dev")

	logger := logger.New(cfg.Env)

	pool, err := postgres.Connect(ctx, cfg.Database)

	if err != nil {
		logger.Error("database connect", slog.String("error", err.Error()))
		return
	}
	defer func() {
		logger.Info("close db connect")

		pool.Close()
	}()

	urlService, err := urlsservice.New(pool, logger)

	if err != nil {
		logger.Error("", slog.String("error", err.Error()))
		return
	}

	server, err := httpserver.New(ctx, cfg.HTTPServer, urlService, logger, cfg.Owner, cfg.Limiter)

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
