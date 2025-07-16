package main

import (
	"go-bucket-manager-bff/internal/api"
	"go-bucket-manager-bff/internal/auth"
	"go-bucket-manager-bff/internal/config"
	"go-bucket-manager-bff/internal/storage"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize storage strategies
	strategies := storage.NewStrategies(cfg)

	// Initialize Auth
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)
	oauthManager := auth.NewOAuthManager(cfg)

	// Setup Gin
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	// Auth endpoints
	api.RegisterAuthRoutes(r, oauthManager, jwtManager, cfg)

	// API endpoints (JWT protected)
	api.RegisterAPIRoutes(r, strategies, jwtManager, cfg)

	log.Printf("Starting server on :%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
