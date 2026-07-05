package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// FingguMain wires config, Redis, the alert manager, the background
// collector, and the HTTP server together, then blocks until the process
// receives a shutdown signal (Ctrl+C, SIGTERM from Docker/k8s, etc).
func FingguMain() {
	cfg := FingguLoadConfig()

	redisClient := FingguNewRedisClient(cfg)
	if err := redisClient.Ping(); err != nil {
		log.Fatalf("finggu: cannot connect to redis at %s: %v", cfg.RedisAddr, err)
	}
	log.Printf("finggu: connected to redis at %s", cfg.RedisAddr)

	alertManager := FingguNewAlertManager(cfg)
	collector := FingguNewCollector(cfg, redisClient, alertManager)
	go collector.Start()

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	api := FingguNewAPI(redisClient, cfg)
	FingguSetupRoutes(router, api)

	server := &http.Server{
		Addr:    "0.0.0.0:" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("finggu: server listening on :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("finggu: server failed: %v", err)
		}
	}()

	fingguFn_WaitForShutdown(server, collector)
}

// fingguFn_WaitForShutdown blocks until SIGINT/SIGTERM, then stops the
// collector and drains in-flight HTTP requests before exiting.
func fingguFn_WaitForShutdown(server *http.Server, collector *FingguCollector) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("finggu: shutting down gracefully...")
	collector.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("finggu: forced shutdown: %v", err)
	}
	log.Println("finggu: shutdown complete")
}

func main() {
	FingguMain()
}
