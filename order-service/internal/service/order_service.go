package service

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"order-service/internal/client"
	"order-service/internal/model"
	"order-service/internal/repository"
)

// OrderService defines business logic for orders.
type OrderService interface {
	Checkout(userID uint, email string, authToken string) (*model.Order, error)
	ListOrders(userID uint) ([]model.Order, error)
	GetOrder(orderID uint) (*model.Order, error)
}

type orderService struct {
	repo          repository.OrderRepository
	serviceClient *client.ServiceClient
	rabbitConn    *amqp.Connection
}

// NewOrderService creates a new OrderService instance.
func NewOrderService(repo repository.OrderRepository, sc *client.ServiceClient, rabbitConn *amqp.Connection) OrderService {
	return &orderService{
		repo:          repo,
		serviceClient: sc,
		rabbitConn:    rabbitConn,
	}
}

func (s *orderService) Checkout(userID uint, email string, authToken string) (*model.Order, error) {
	// 1. Fetch cart items from Cart Service
	cartResp, err := s.serviceClient.GetCart(authToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cart: %w", err)
	}

	if len(cartResp.Items) == 0 {
		return nil, fmt.Errorf("cart is empty")
	}

	// 2. Reduce stock for each product in Product Service
	for _, item := range cartResp.Items {
		if err := s.serviceClient.ReduceStock(authToken, item.ProductID, item.Quantity); err != nil {
			return nil, fmt.Errorf("failed to reduce stock for product %d: %w", item.ProductID, err)
		}
	}

	// 3. Create order record with PENDING status
	orderItems := make([]model.OrderItem, len(cartResp.Items))
	for i, item := range cartResp.Items {
		orderItems[i] = model.OrderItem{
			ProductID: item.ProductID,
			Name:      item.Name,
			Price:     item.Price,
			Quantity:  item.Quantity,
		}
	}

	order := &model.Order{
		UserID:      userID,
		Items:       orderItems,
		TotalAmount: cartResp.Total,
		Status:      "PENDING",
	}

	if err := s.repo.Create(order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// 4. Process payment via Payment Service
	paymentResp, err := s.serviceClient.ProcessPayment(order.ID, order.TotalAmount)
	if err != nil {
		_ = s.repo.UpdateStatus(order.ID, "FAILED")
		return nil, fmt.Errorf("payment processing failed: %w", err)
	}

	// 5. Update order status based on payment result
	if paymentResp.Status == "SUCCESS" {
		_ = s.repo.UpdateStatus(order.ID, "PAID")
		order.Status = "PAID"

		// 6. Clear user's cart after successful payment
		if err := s.serviceClient.ClearCart(authToken); err != nil {
			log.Printf("Warning: failed to clear cart for user %d: %v", userID, err)
		}
	} else {
		_ = s.repo.UpdateStatus(order.ID, "FAILED")
		order.Status = "FAILED"
	}

	// 7. Publish order event to RabbitMQ for notification service
	s.publishOrderEvent(order, email)

	return order, nil
}

func (s *orderService) ListOrders(userID uint) ([]model.Order, error) {
	return s.repo.FindByUserID(userID)
}

func (s *orderService) GetOrder(orderID uint) (*model.Order, error) {
	return s.repo.FindByID(orderID)
}

func (s *orderService) publishOrderEvent(order *model.Order, email string) {
	if s.rabbitConn == nil {
		log.Println("RabbitMQ connection not available, skipping event publish")
		return
	}

	ch, err := s.rabbitConn.Channel()
	if err != nil {
		log.Printf("Failed to open RabbitMQ channel: %v", err)
		return
	}
	defer ch.Close()

	// Declare the queue (idempotent)
	q, err := ch.QueueDeclare(
		"order_notifications", // queue name
		true,                  // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		log.Printf("Failed to declare queue: %v", err)
		return
	}

	event := model.OrderEvent{
		OrderID:     order.ID,
		UserID:      order.UserID,
		UserEmail:   email,
		TotalAmount: order.TotalAmount,
		Status:      order.Status,
		Items:       order.Items,
		CreatedAt:   time.Now(),
	}

	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal order event: %v", err)
		return
	}

	err = ch.Publish(
		"",     // exchange (default)
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	if err != nil {
		log.Printf("Failed to publish order event: %v", err)
		return
	}

	log.Printf("Published order event for order #%d to RabbitMQ", order.ID)
}
