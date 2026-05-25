package model

import "time"

// Order represents a customer order.
type Order struct {
	ID          uint        `json:"id" gorm:"primaryKey"`
	UserID      uint        `json:"user_id" gorm:"index;not null"`
	Items       []OrderItem `json:"items" gorm:"foreignKey:OrderID"`
	TotalAmount float64     `json:"total_amount" gorm:"not null"`
	Status      string      `json:"status" gorm:"not null;default:PENDING"` // PENDING, PAID, FAILED, CANCELLED
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// OrderItem represents a single item within an order.
type OrderItem struct {
	ID        uint    `json:"id" gorm:"primaryKey"`
	OrderID   uint    `json:"order_id" gorm:"index;not null"`
	ProductID uint    `json:"product_id" gorm:"not null"`
	Name      string  `json:"name" gorm:"not null"`
	Price     float64 `json:"price" gorm:"not null"`
	Quantity  int     `json:"quantity" gorm:"not null"`
}

// OrderEvent is published to RabbitMQ after a successful order.
type OrderEvent struct {
	OrderID     uint        `json:"order_id"`
	UserID      uint        `json:"user_id"`
	UserEmail   string      `json:"user_email"`
	TotalAmount float64     `json:"total_amount"`
	Status      string      `json:"status"`
	Items       []OrderItem `json:"items"`
	CreatedAt   time.Time   `json:"created_at"`
}
