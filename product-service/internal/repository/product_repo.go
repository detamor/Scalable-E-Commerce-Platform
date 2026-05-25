package repository

import (
	"product-service/internal/model"

	"gorm.io/gorm"
)

// ProductRepository defines data access operations for products.
type ProductRepository interface {
	Create(product *model.Product) error
	FindAll(page, limit int) ([]model.Product, int64, error)
	FindByID(id uint) (*model.Product, error)
	UpdateStock(id uint, quantity int) error
}

type productRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new ProductRepository instance.
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(product *model.Product) error {
	return r.db.Create(product).Error
}

func (r *productRepository) FindAll(page, limit int) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64

	r.db.Model(&model.Product{}).Count(&total)

	offset := (page - 1) * limit
	if err := r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepository) FindByID(id uint) (*model.Product, error) {
	var product model.Product
	if err := r.db.First(&product, id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) UpdateStock(id uint, quantity int) error {
	// Use a transaction to safely update stock
	return r.db.Transaction(func(tx *gorm.DB) error {
		var product model.Product
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&product, id).Error; err != nil {
			return err
		}

		newStock := product.Stock + quantity
		if newStock < 0 {
			return gorm.ErrInvalidData
		}

		return tx.Model(&product).Update("stock", newStock).Error
	})
}
