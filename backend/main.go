package main

import (
	"log"
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

	stats.Start()

	r := api.SetupRouter(cfg)
	log.Printf("WireGuard UI listening on %s (db=%s)", cfg.Listen, cfg.DBPath)
	if err := r.Run(cfg.Listen); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
