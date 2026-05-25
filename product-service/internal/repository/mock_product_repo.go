package repository

import (
	"errors"

	"product-service/internal/model"
)

// MockProductRepository is a mock implementation of ProductRepository for testing.
type MockProductRepository struct {
	Products map[uint]*model.Product
	NextID   uint
}

// NewMockProductRepository creates a new MockProductRepository for testing.
func NewMockProductRepository() *MockProductRepository {
	return &MockProductRepository{
		Products: make(map[uint]*model.Product),
		NextID:   1,
	}
}

func (m *MockProductRepository) Create(product *model.Product) error {
	product.ID = m.NextID
	m.NextID++
	m.Products[product.ID] = product
	return nil
}

func (m *MockProductRepository) FindAll(page, limit int) ([]model.Product, int64, error) {
	products := make([]model.Product, 0, len(m.Products))
	for _, p := range m.Products {
		products = append(products, *p)
	}
	total := int64(len(products))

	// Simple pagination
	start := (page - 1) * limit
	if start >= len(products) {
		return []model.Product{}, total, nil
	}
	end := start + limit
	if end > len(products) {
		end = len(products)
	}

	return products[start:end], total, nil
}

func (m *MockProductRepository) FindByID(id uint) (*model.Product, error) {
	product, exists := m.Products[id]
	if !exists {
		return nil, errors.New("product not found")
	}
	return product, nil
}

func (m *MockProductRepository) UpdateStock(id uint, quantity int) error {
	product, exists := m.Products[id]
	if !exists {
		return errors.New("product not found")
	}
	newStock := product.Stock + quantity
	if newStock < 0 {
		return errors.New("insufficient stock")
	}
	product.Stock = newStock
	return nil
}
