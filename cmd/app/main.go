package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/Cwby333/url-shorter/internal/config"
	"github.com/Cwby333/url-shorter/internal/logger"
	"github.com/Cwby333/url-shorter/internal/repository/postgres"
	"github.com/Cwby333/url-shorter/internal/services/urlsservice"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/registerrouters"
)

func main() {
	cfg := config.Load("dev")

	logger := logger.New(cfg.Env)

	pool, err := postgres.Connect(context.Background(), cfg.Database)

	if err != nil {
		logger.Error("database connect error", slog.String("error", err.Error()))
	}

	urlService, err := urlsservice.New(pool, logger)

	if err != nil {
		os.Exit(1)
	}

	router, err := registerrouters.Register(urlService, logger, cfg.Owner)

	if err != nil {
		logger.Error("init router", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("start server")

	http.ListenAndServe(cfg.Address, router)
}
