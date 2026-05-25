package consumer

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// OrderEvent matches the event published by the Order Service.
type OrderEvent struct {
	OrderID     uint    `json:"order_id"`
	UserID      uint    `json:"user_id"`
	UserEmail   string  `json:"user_email"`
	TotalAmount float64 `json:"total_amount"`
	Status      string  `json:"status"`
	Items       []struct {
		ProductID uint    `json:"product_id"`
		Name      string  `json:"name"`
		Price     float64 `json:"price"`
		Quantity  int     `json:"quantity"`
	} `json:"items"`
	CreatedAt time.Time `json:"created_at"`
}

// StartConsumer connects to RabbitMQ and starts consuming order notification messages.
func StartConsumer(rabbitURL string) error {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	// Declare the same queue as the producer (idempotent)
	q, err := ch.QueueDeclare(
		"order_notifications",
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Set prefetch count to 1 for fair dispatch
	if err := ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",    // consumer tag
		false, // auto-ack (manual ack for reliability)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	log.Printf("✅ Notification Service is listening on queue: %s", q.Name)
	log.Println("Waiting for order events...")

	// Block and process messages
	for msg := range msgs {
		processMessage(msg)
	}

	return nil
}

func processMessage(msg amqp.Delivery) {
	var event OrderEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("❌ Failed to unmarshal message: %v", err)
		msg.Nack(false, false) // Don't requeue malformed messages
		return
	}

	// Simulate sending notification
	log.Println("═══════════════════════════════════════════════════")
	log.Printf("📧 NOTIFICATION - Order #%d", event.OrderID)
	log.Println("═══════════════════════════════════════════════════")
	log.Printf("  To:      %s", event.UserEmail)
	log.Printf("  Status:  %s", event.Status)
	log.Printf("  Total:   $%.2f", event.TotalAmount)
	log.Printf("  Items:")
	for _, item := range event.Items {
		log.Printf("    - %s (x%d) @ $%.2f", item.Name, item.Quantity, item.Price)
	}

	if event.Status == "PAID" {
		log.Printf("  ✅ Payment successful! Order confirmed.")
		log.Printf("  📱 [SIMULATED] SMS sent to user %d", event.UserID)
		log.Printf("  📧 [SIMULATED] Email sent to %s", event.UserEmail)
	} else {
		log.Printf("  ❌ Payment failed. Order not confirmed.")
		log.Printf("  📧 [SIMULATED] Failure notification sent to %s", event.UserEmail)
	}
	log.Println("═══════════════════════════════════════════════════")

	// Acknowledge the message
	msg.Ack(false)
}
