package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"ocean-digital-twin/internal/server"
	"ocean-digital-twin/internal/utils/erddap"
)

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func main() {
	minLat := 40.83
	minLon := 1.10
	maxLat := 41.26
	maxLon := 2.53
	endTime := time.Now().UTC()
	startTime := endTime.Add(-7 * 24 * time.Hour)
	logger := slog.Default()
	downloader := erddap.NewDownloader(logger, minLat, minLon, maxLat, maxLon)
	data, err := downloader.DownloadChlorophyllData(context.Background(), startTime, endTime)
	if err != nil {
		slog.Error("error downloading chlor data", "err", err)
	}
	fmt.Println(len(data))

	log.Println("Getting new server...")
	server := server.NewServer()
	log.Println("Gotten new server")

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, done)

	log.Printf("Starting server on port%v...\n", server.Addr)
	err = server.ListenAndServe()
	log.Println("Started server on port", server.Addr)
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
