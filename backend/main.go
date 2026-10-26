package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wireguard-ui/api"
	"wireguard-ui/config"
	"wireguard-ui/db"
	"wireguard-ui/stats"
	"wireguard-ui/wg"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	wg.SetConfigDir(cfg.ConfigDir)

	if err := db.Init(cfg.DBPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	sampler := stats.Start(ctx)

	srv := &http.Server{
		Addr:              cfg.Listen,
		Handler:           api.SetupRouter(cfg),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       time.Minute,
		IdleTimeout:       2 * time.Minute,
	}
	log.Printf("WireGuard UI listening on %s (db=%s)", cfg.Listen, cfg.DBPath)

	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.ListenAndServe() }()

	select {
	case err := <-serveErr:
		log.Fatalf("Failed to start server: %v", err)
	case <-ctx.Done():
	}
	stop() // a second signal now kills the process instead of waiting
	log.Print("Shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown: %v", err)
	}
	<-sampler
	if err := db.Close(); err != nil {
		log.Printf("Closing database: %v", err)
	}
}
