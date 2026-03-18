package main

import (
	"log"

	"github.com/ffreville/mmo-team-test/server/internal/config"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting Game Server v%s (commit: %s)", cfg.BuildVersion, cfg.BuildCommit)
	log.Printf("Listening on %s:%d", cfg.Server.Host, cfg.Server.Port)

	// TODO: Initialize database connection
	// TODO: Initialize Redis connection
	// TODO: Initialize game world
	// TODO: Start WebSocket gateway
	// TODO: Start HTTP health endpoint

	// Placeholder - server runs until interrupt
	select {}
}

// Build info injected at compile time
var (
	BuildVersion = "0.1.0-dev"
	BuildCommit  = "unknown"
	BuildTime    = "unknown"
)
