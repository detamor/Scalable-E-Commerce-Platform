package service

import (
	"fmt"
	"math/rand"
	"time"

	"payment-service/internal/model"
	"payment-service/internal/repository"
)

// PaymentService defines business logic for payment processing.
type PaymentService interface {
	ProcessPayment(req model.ProcessPaymentRequest) (*model.ProcessPaymentResponse, error)
}

type paymentService struct {
	repo repository.PaymentRepository
}

// NewPaymentService creates a new PaymentService instance.
func NewPaymentService(repo repository.PaymentRepository) PaymentService {
	return &paymentService{repo: repo}
}

func (s *paymentService) ProcessPayment(req model.ProcessPaymentRequest) (*model.ProcessPaymentResponse, error) {
	// Generate a mock transaction ID
	transactionID := generateTransactionID()

	// Simulate payment processing — 90% success rate
	status := "SUCCESS"
	if rand.Intn(10) == 0 {
		status = "FAILED"
	}

	payment := &model.Payment{
		OrderID:       req.OrderID,
		Amount:        req.Amount,
		Status:        status,
		TransactionID: transactionID,
	}

	if err := s.repo.Create(payment); err != nil {
		return nil, fmt.Errorf("failed to record payment: %w", err)
	}

	return &model.ProcessPaymentResponse{
		TransactionID: transactionID,
		Status:        status,
	}, nil
}

func generateTransactionID() string {
	return fmt.Sprintf("TXN-%d-%d", time.Now().UnixNano(), rand.Intn(10000))
}
