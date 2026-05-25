package model

import "time"

// Payment represents a payment transaction.
type Payment struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	OrderID       uint      `json:"order_id" gorm:"index;not null"`
	Amount        float64   `json:"amount" gorm:"not null"`
	Status        string    `json:"status" gorm:"not null"` // SUCCESS, FAILED, PENDING
	TransactionID string    `json:"transaction_id" gorm:"uniqueIndex"`
	CreatedAt     time.Time `json:"created_at"`
}

// ProcessPaymentRequest is the payload for processing a payment.
type ProcessPaymentRequest struct {
	OrderID uint    `json:"order_id" binding:"required"`
	Amount  float64 `json:"amount" binding:"required,gt=0"`
}

// ProcessPaymentResponse is the response after payment processing.
type ProcessPaymentResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}
