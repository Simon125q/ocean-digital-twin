package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ocean-digital-twin/internal/database"
	"ocean-digital-twin/internal/server"
	"ocean-digital-twin/internal/utils/scheduler"
)

const (
	minLat = 40.83
	minLon = 1.10
	maxLat = 41.26
	maxLon = 2.53
)

func main() {
	// set up logger
	// logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	server, dbService := server.NewServer()

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// create context for the app
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// init automatic data updater
	updater := scheduler.NewUpdater(
		dbService,
		logger,
		1*time.Hour,
		minLat,
		minLon,
		maxLat,
		maxLon,
	)

	// start the updater in goroutine
	go updater.Start(ctx)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, dbService, done, logger)

	logger.Info(fmt.Sprintf("Starting server on port:%v...\n", server.Addr))
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	logger.Info("Graceful shutdown complete.")
}

func gracefulShutdown(apiServer *http.Server, dbService database.Service, done chan bool, logger *slog.Logger) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	logger.Info("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		logger.Info("Server forced to shutdown with error", "err", err)
	}

	if dbService != nil {
		if err := dbService.Close(); err != nil {
			logger.Error("Error closing database connection", "err", err)
		}
	}

	logger.Info("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}
