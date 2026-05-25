package main

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/gin-gonic/gin"

	"order-service/internal/client"
	"order-service/internal/config"
	"order-service/internal/handler"
	"order-service/internal/middleware"
	"order-service/internal/repository"
	"order-service/internal/service"
)

func main() {
	cfg := config.LoadConfig()
	db := config.ConnectDB(cfg)

	// Connect to RabbitMQ
	rabbitConn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to RabbitMQ: %v. Order events will not be published.", err)
	} else {
		defer rabbitConn.Close()
		log.Println("RabbitMQ connected successfully")
	}

	// Initialize clients for inter-service communication
	serviceClient := client.NewServiceClient(
		cfg.CartServiceURL,
		cfg.ProductServiceURL,
		cfg.PaymentServiceURL,
	)

	orderRepo := repository.NewOrderRepository(db)
	orderSvc := service.NewOrderService(orderRepo, serviceClient, rabbitConn)
	orderHandler := handler.NewOrderHandler(orderSvc)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "order-service"})
	})

	v1 := router.Group("/api/v1/orders")
	v1.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		v1.POST("/checkout", orderHandler.Checkout)
		v1.GET("", orderHandler.ListOrders)
		v1.GET("/:id", orderHandler.GetOrder)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Order Service starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
