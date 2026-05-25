package main

import (
	"log"

	"notification-service/internal/config"
	"notification-service/internal/consumer"
)

func main() {
	cfg := config.LoadConfig()

	log.Println("Starting Notification Service (RabbitMQ Consumer)...")

	if err := consumer.StartConsumer(cfg.RabbitMQURL); err != nil {
		log.Fatalf("Notification service error: %v", err)
	}
}
