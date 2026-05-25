package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"user-service/internal/config"
	"user-service/internal/handler"
	"user-service/internal/middleware"
	"user-service/internal/repository"
	"user-service/internal/service"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	db := config.ConnectDB(cfg)

	// Initialize layers
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo, cfg.JWTSecret)
	userHandler := handler.NewUserHandler(userSvc)

	// Setup Gin router
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "user-service"})
	})

	// User routes
	v1 := router.Group("/api/v1/users")
	{
		v1.POST("/register", userHandler.Register)
		v1.POST("/login", userHandler.Login)
		v1.GET("/me", middleware.AuthMiddleware(cfg.JWTSecret), userHandler.GetMe)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("User Service starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
