// Package main boots the TownHall API: config, logging, database, routes.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ArifulProtik/TownHall/internal/auth"
	"ArifulProtik/TownHall/internal/config"
	"ArifulProtik/TownHall/internal/platform"
	"ArifulProtik/TownHall/pkg/logger"
	"ArifulProtik/TownHall/pkg/validation"

	appmiddleware "ArifulProtik/TownHall/internal/middleware"

	"github.com/labstack/echo/v5"
)

func main() {
	// run logs failures itself; exit non-zero so supervisors restart.
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	cfg := config.New()
	log := logger.New(cfg.AppEnv, cfg.LogLevel)
	ctx := context.Background()

	entClient, err := platform.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("platform open failed", slog.Any("error", err))
		return err
	}
	defer func() { _ = entClient.Close() }()

	if cfg.AppEnv == "development" {
		if err := platform.AutoMigrate(ctx, entClient); err != nil {
			log.Error("automigrate failed", slog.Any("error", err))
			return err
		}
	}

	e := echo.New()
	e.Validator = validation.New()
	// Route Echo's own logs (banner, startup, HTTP errors) through our
	// handler — otherwise they bypass it as JSON.
	e.Logger = log
	appmiddleware.Register(e, log)

	authSvc := auth.NewService(cfg, entClient, log)
	authHandler := auth.NewHandler(authSvc, log)

	api := e.Group("/api/v1")
	authHandler.RegisterRoutes(api.Group("/auth"))

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sc := echo.StartConfig{
		Address:         ":" + cfg.Port,
		GracefulTimeout: 10 * time.Second,
	}
	log.Info("server starting", slog.String("addr", ":"+cfg.Port), slog.String("env", cfg.AppEnv))
	if err := sc.Start(sigCtx, e); err != nil {
		log.Info("server stopped", slog.Any("error", err))
	}
	return nil
}
