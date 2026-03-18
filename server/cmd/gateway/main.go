package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ffreville/mmo-team-test/server/internal/auth"
	"github.com/ffreville/mmo-team-test/server/internal/config"
	"github.com/ffreville/mmo-team-test/server/internal/database"
	"github.com/ffreville/mmo-team-test/server/internal/game/world"
	"github.com/ffreville/mmo-team-test/server/internal/network"
	"github.com/gorilla/websocket"
	"strconv"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting WebSocket Gateway v%s (commit: %s)", cfg.BuildVersion, cfg.BuildCommit)
	log.Printf("Listening on %s:%d", cfg.Server.Host, cfg.Server.Port)

	var db *database.PostgresDB
	var redisClient *database.RedisClient
	var authService *auth.AuthService

	if cfg.Database.Host != "" {
		db, err = database.NewPostgresDB(
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Name,
			cfg.Database.MaxOpenConns,
			cfg.Database.MaxIdleConns,
			cfg.Database.ConnMaxLifetime,
		)
		if err != nil {
			log.Printf("Warning: Failed to connect to database: %v", err)
			log.Printf("Running without database (using in-memory storage)")
		}
	}

	if cfg.Redis.Host != "" {
		redisClient, err = database.NewRedisClient(
			cfg.Redis.Host,
			cfg.Redis.Port,
			cfg.Redis.DB,
			cfg.Redis.Password,
		)
		if err != nil {
			log.Printf("Warning: Failed to connect to Redis: %v", err)
			log.Printf("Running without Redis (sessions will not persist)")
		}
	}

	if db != nil && redisClient != nil {
		authService = auth.NewAuthService(
			db,
			redisClient,
			cfg.Auth.JWTSecret,
			cfg.Auth.BcryptCost,
		)
	}

	gameWorld := world.NewWorld(db)

	gateway := network.NewGateway(authService, gameWorld, redisClient)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade failed: %v", err)
			return
		}
		gateway.HandleConnection(conn)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:         cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("Gateway listening on %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interrupt

	log.Println("Shutting down gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	server.Shutdown(ctx)
	log.Println("Gateway stopped")
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var (
	BuildVersion = "0.1.0-dev"
	BuildCommit  = "unknown"
	BuildTime    = "unknown"
)
