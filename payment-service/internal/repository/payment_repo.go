package repository

import (
	"payment-service/internal/model"

	"gorm.io/gorm"
)

// PaymentRepository defines data access operations for payments.
type PaymentRepository interface {
	Create(payment *model.Payment) error
	FindByOrderID(orderID uint) (*model.Payment, error)
}

type paymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository creates a new PaymentRepository instance.
func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(payment *model.Payment) error {
	return r.db.Create(payment).Error
}

func (r *paymentRepository) FindByOrderID(orderID uint) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.Where("order_id = ?", orderID).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}
