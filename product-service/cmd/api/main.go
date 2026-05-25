package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"product-service/internal/config"
	"product-service/internal/handler"
	"product-service/internal/middleware"
	"product-service/internal/repository"
	"product-service/internal/service"
)

func main() {
	cfg := config.LoadConfig()
	db := config.ConnectDB(cfg)

	productRepo := repository.NewProductRepository(db)
	productSvc := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productSvc)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "product-service"})
	})

	v1 := router.Group("/api/v1/products")
	{
		v1.GET("", productHandler.List)
		v1.GET("/:id", productHandler.Get)
		v1.POST("", middleware.AuthMiddleware(cfg.JWTSecret), productHandler.Create)
		v1.PUT("/:id/stock", middleware.AuthMiddleware(cfg.JWTSecret), productHandler.UpdateStock)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Product Service starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
