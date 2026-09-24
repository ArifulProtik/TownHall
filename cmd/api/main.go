// Package main boots the TownHall API: config, logging, database, routes.
package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ArifulProtik/TownHall/ent"
	"ArifulProtik/TownHall/internal/auth"
	"ArifulProtik/TownHall/internal/config"
	"ArifulProtik/TownHall/internal/eventbus"
	"ArifulProtik/TownHall/internal/filestore"
	"ArifulProtik/TownHall/internal/notification"
	"ArifulProtik/TownHall/internal/platform"
	"ArifulProtik/TownHall/internal/profile"
	"ArifulProtik/TownHall/internal/social"
	"ArifulProtik/TownHall/pkg/logger"
	"ArifulProtik/TownHall/pkg/validation"

	appmiddleware "ArifulProtik/TownHall/internal/middleware"

	"github.com/labstack/echo/v5"
)

func main() {
	if err := run(); err != nil {
		log.Println("townhall:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.New()
	if err != nil {
		return err
	}
	appLog := logger.New(cfg.AppEnv, cfg.LogLevel)
	slog.SetDefault(appLog)
	ctx := context.Background()

	entClient, err := openWithRetry(ctx, appLog, cfg.DatabaseURL)
	if err != nil {
		appLog.Error("platform open failed", slog.Any("error", err))
		return err
	}
	defer func() { _ = entClient.Close() }()

	if cfg.AppEnv == "development" {
		if err := platform.AutoMigrate(ctx, entClient); err != nil {
			appLog.Error("automigrate failed", slog.Any("error", err))
			return err
		}
	} else if cfg.AutoMigrate {
		// Explicit opt-in (e.g. local compose): non-destructive create only.
		if err := platform.Migrate(ctx, entClient); err != nil {
			appLog.Error("migrate failed", slog.Any("error", err))
			return err
		}
	}

	e := echo.New()
	e.Validator = validation.New()
	// Route Echo's own logs (banner, startup, HTTP errors) through our
	// handler — otherwise they bypass it as JSON.
	e.Logger = appLog
	appmiddleware.Register(e, appLog)

	authSvc := auth.NewService(entClient, cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL, auth.NewMemoryLimiter())
	authHandler := auth.NewHandler(authSvc, cfg.IsProd(), cfg.AppEnv)

	profileSvc := profile.NewService(entClient, filestore.New(cfg.UploadthingToken, "./uploads"))
	profileHandler := profile.NewHandler(profileSvc, cfg.JWTSecret)

	notifBroker, err := notification.NewBroker(cfg.RedisURL)
	if err != nil {
		appLog.Error("notification broker failed", slog.Any("error", err))
		return err
	}
	notifSvc := notification.NewService(entClient, notifBroker)
	notifHandler := notification.NewHandler(notifSvc)

	// The bus carries domain events; notification subscribes. Adding a
	// future event type touches neither this wiring nor publishers.
	bus := eventbus.New()
	notification.Register(bus, notifSvc)

	socialSvc := social.NewService(entClient, bus)
	socialHandler := social.NewHandler(socialSvc, cfg.JWTSecret)

	e.Static("/uploads", "./uploads")

	api := e.Group("/api/v1")
	public := api.Group("")
	protected := api.Group("", auth.Middleware(cfg.JWTSecret))

	authHandler.RegisterRoutes(public, protected)
	profileHandler.RegisterRoutes(public, protected)
	socialHandler.RegisterRoutes(public, protected)
	notifHandler.RegisterRoutes(public, protected)

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sc := echo.StartConfig{
		Address:         ":" + cfg.Port,
		GracefulTimeout: 10 * time.Second,
	}
	appLog.Info("server starting", slog.String("addr", ":"+cfg.Port), slog.String("env", cfg.AppEnv))
	if err := sc.Start(sigCtx, e); err != nil {
		appLog.Info("server stopped", slog.Any("error", err))
	}
	return nil
}

// openWithRetry dials Postgres until it accepts connections or the
// budget (15 × 2s) runs out — containers rarely start in order.
func openWithRetry(ctx context.Context, appLog *slog.Logger, databaseURL string) (*ent.Client, error) {
	var err error
	for attempt := 1; attempt <= 15; attempt++ {
		var client *ent.Client
		client, err = platform.Open(ctx, databaseURL)
		if err == nil {
			return client, nil
		}
		appLog.Warn("database not ready, retrying",
			slog.Int("attempt", attempt), slog.Any("error", err))
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return nil, err
}
