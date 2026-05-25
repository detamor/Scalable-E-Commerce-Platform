package repository

import (
	"order-service/internal/model"

	"gorm.io/gorm"
)

// OrderRepository defines data access operations for orders.
type OrderRepository interface {
	Create(order *model.Order) error
	FindByUserID(userID uint) ([]model.Order, error)
	FindByID(id uint) (*model.Order, error)
	UpdateStatus(id uint, status string) error
}

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new OrderRepository instance.
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) FindByUserID(userID uint) ([]model.Order, error) {
	var orders []model.Order
	if err := r.db.Preload("Items").Where("user_id = ?", userID).Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *orderRepository) FindByID(id uint) (*model.Order, error) {
	var order model.Order
	if err := r.db.Preload("Items").First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Order{}).Where("id = ?", id).Update("status", status).Error
}
