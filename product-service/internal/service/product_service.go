package service

import (
	"errors"

	"product-service/internal/model"
	"product-service/internal/repository"
)

// ProductService defines business logic for products.
type ProductService interface {
	Create(req model.CreateProductRequest) (*model.Product, error)
	List(page, limit int) ([]model.Product, int64, error)
	Get(id uint) (*model.Product, error)
	UpdateStock(id uint, quantity int) error
}

type productService struct {
	repo repository.ProductRepository
}

// NewProductService creates a new ProductService instance.
func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) Create(req model.CreateProductRequest) (*model.Product, error) {
	product := &model.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
		ImageURL:    req.ImageURL,
	}

	if err := s.repo.Create(product); err != nil {
		return nil, errors.New("failed to create product")
	}

	return product, nil
}

func (s *productService) List(page, limit int) ([]model.Product, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.repo.FindAll(page, limit)
}

func (s *productService) Get(id uint) (*model.Product, error) {
	return s.repo.FindByID(id)
}

func (s *productService) UpdateStock(id uint, quantity int) error {
	return s.repo.UpdateStock(id, quantity)
}
