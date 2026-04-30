package app

import (
	"context"
	"fmt"

	"github.com/diyas/fintrack/config"
	"github.com/diyas/fintrack/internal/delivery"
	"github.com/diyas/fintrack/internal/repository"
	"github.com/diyas/fintrack/internal/server"
	"github.com/diyas/fintrack/internal/service"
	postgres "github.com/diyas/fintrack/pkg/adapter"
	"github.com/diyas/fintrack/pkg/logger"
)

func StartWebServer(ctx context.Context, cfg *config.Config, appLogger *logger.Logger) error {
	appLogger.WithComponent("app").Info("Initializing web server")

	appLogger.WithComponent("database").Info("Connecting to database")
	pg, err := postgres.New(ctx, &cfg.PG)
	if err != nil {
		appLogger.WithComponent("database").WithError(err).Error("Failed to initialize database connection")
		return fmt.Errorf("could not init postgres connection: %w", err)
	}
	defer func() {
		appLogger.WithComponent("database").Info("Closing database connection")
		pg.Close()
	}()
	appLogger.WithComponent("database").Info("Database connection established")

	appLogger.WithComponent("redis").Info("Connecting to redis")
	rdb, err := postgres.NewRedis(ctx, &cfg.Redis)
	if err != nil {
		appLogger.WithComponent("redis").WithError(err).Error("Failed to initialize redis connection")
		return fmt.Errorf("could not init redis connection: %w", err)
	}
	defer func() {
		appLogger.WithComponent("redis").Info("Closing redis connection")
		_ = rdb.Close()
	}()
	appLogger.WithComponent("redis").Info("Redis connection established")

	appLogger.WithComponent("repository").Info("Initializing repositories")
	repos := repository.NewRepositories(pg)

	appLogger.WithComponent("service").Info("Initializing services")
	services := service.NewServices(service.Deps{
		Repos:  repos,
		Config: cfg,
		Redis:  rdb,
	})

	appLogger.WithComponent("handler").Info("Initializing handlers")
	handlers := delivery.NewHandler(services, appLogger)

	appLogger.WithComponent("server").Info("Initializing HTTP server")
	srv := server.NewServer(cfg, handlers.Init(cfg), appLogger)

	appLogger.WithComponent("server").WithFields(logger.Fields{
		"host": cfg.Http.Host,
		"port": cfg.Http.Port,
	}).Info("Starting HTTP server")

	defer func() {
		appLogger.WithComponent("server").Info("Stopping HTTP server")
		srv.Stop()
	}()

	srv.Run()
	appLogger.WithComponent("server").Info("HTTP server started successfully")

	<-ctx.Done()
	appLogger.WithComponent("app").Info("Received shutdown signal")

	return nil
}
