package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"cart-service/internal/config"
	"cart-service/internal/handler"
	"cart-service/internal/middleware"
	"cart-service/internal/repository"
	"cart-service/internal/service"
)

func main() {
	cfg := config.LoadConfig()
	redisClient := config.ConnectRedis(cfg)

	cartRepo := repository.NewCartRepository(redisClient)
	cartSvc := service.NewCartService(cartRepo)
	cartHandler := handler.NewCartHandler(cartSvc)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "cart-service"})
	})

	v1 := router.Group("/api/v1/cart")
	v1.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		v1.GET("", cartHandler.GetCart)
		v1.POST("", cartHandler.AddItem)
		v1.PUT("/:productId", cartHandler.UpdateQuantity)
		v1.DELETE("/:productId", cartHandler.RemoveItem)
		v1.DELETE("", cartHandler.ClearCart)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Cart Service starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
