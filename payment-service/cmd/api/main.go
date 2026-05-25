package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"payment-service/internal/config"
	"payment-service/internal/handler"
	"payment-service/internal/repository"
	"payment-service/internal/service"
)

func main() {
	cfg := config.LoadConfig()
	db := config.ConnectDB(cfg)

	paymentRepo := repository.NewPaymentRepository(db)
	paymentSvc := service.NewPaymentService(paymentRepo)
	paymentHandler := handler.NewPaymentHandler(paymentSvc)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "payment-service"})
	})

	v1 := router.Group("/api/v1/payments")
	{
		v1.POST("/process", paymentHandler.ProcessPayment)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Payment Service starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
