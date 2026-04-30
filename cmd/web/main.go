package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/diyas/fintrack/config"
	"github.com/diyas/fintrack/internal/app"
	"github.com/diyas/fintrack/pkg/logger"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Load configuration first
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize custom logger
	appLogger, err := logger.New(&cfg.Logger)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer appLogger.Close()

	// Set as global logger
	appLogger.SetGlobal()

	// Log application startup
	appLogger.WithFields(logger.Fields{
		"service":     cfg.Logger.Service,
		"version":     cfg.Logger.Version,
		"environment": cfg.Logger.Environment,
	}).Info("Application starting")

	if err := app.StartWebServer(ctx, cfg, appLogger); err != nil {
		appLogger.WithError(err).Fatal("Failed to start web server")
	}
}
