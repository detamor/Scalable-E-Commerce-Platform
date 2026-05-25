package service

import (
	"testing"

	"product-service/internal/model"
	"product-service/internal/repository"
)

func TestCreateProduct_Success(t *testing.T) {
	mockRepo := repository.NewMockProductRepository()
	svc := NewProductService(mockRepo)

	req := model.CreateProductRequest{
		Name:        "Gaming Laptop",
		Description: "High-end gaming laptop",
		Price:       1299.99,
		Stock:       50,
		Category:    "Electronics",
	}

	product, err := svc.Create(req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if product.Name != req.Name {
		t.Errorf("expected name %s, got %s", req.Name, product.Name)
	}

	if product.Price != req.Price {
		t.Errorf("expected price %f, got %f", req.Price, product.Price)
	}

	if product.Stock != req.Stock {
		t.Errorf("expected stock %d, got %d", req.Stock, product.Stock)
	}

	if product.ID == 0 {
		t.Error("expected product ID to be set")
	}
}

func TestListProducts_Success(t *testing.T) {
	mockRepo := repository.NewMockProductRepository()
	svc := NewProductService(mockRepo)

	// Create some products
	for i := 0; i < 5; i++ {
		req := model.CreateProductRequest{
			Name:  "Product",
			Price: 10.0,
			Stock: 10,
		}
		_, _ = svc.Create(req)
	}

	products, total, err := svc.List(1, 10)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}

	if len(products) != 5 {
		t.Errorf("expected 5 products, got %d", len(products))
	}
}

func TestListProducts_DefaultPagination(t *testing.T) {
	mockRepo := repository.NewMockProductRepository()
	svc := NewProductService(mockRepo)

	// Test that invalid page/limit values are handled
	_, _, err := svc.List(0, 0)
	if err != nil {
		t.Fatalf("expected no error with default pagination, got: %v", err)
	}
}

func TestGetProduct_Success(t *testing.T) {
	mockRepo := repository.NewMockProductRepository()
	svc := NewProductService(mockRepo)

	req := model.CreateProductRequest{
		Name:  "Test Product",
		Price: 25.99,
		Stock: 10,
	}

	created, _ := svc.Create(req)

	product, err := svc.Get(created.ID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if product.Name != req.Name {
		t.Errorf("expected name %s, got %s", req.Name, product.Name)
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	mockRepo := repository.NewMockProductRepository()
	svc := NewProductService(mockRepo)

	_, err := svc.Get(999)
	if err == nil {
		t.Fatal("expected error for non-existent product, got nil")
	}
}

func TestUpdateStock_Success(t *testing.T) {
	mockRepo := repository.NewMockProductRepository()
	svc := NewProductService(mockRepo)

	req := model.CreateProductRequest{
		Name:  "Test Product",
		Price: 25.99,
		Stock: 50,
	}
	created, _ := svc.Create(req)

	// Reduce stock by 10
	err := svc.UpdateStock(created.ID, -10)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify stock was updated
	product, _ := svc.Get(created.ID)
	if product.Stock != 40 {
		t.Errorf("expected stock 40, got %d", product.Stock)
	}
}

func TestUpdateStock_InsufficientStock(t *testing.T) {
	mockRepo := repository.NewMockProductRepository()
	svc := NewProductService(mockRepo)

	req := model.CreateProductRequest{
		Name:  "Test Product",
		Price: 25.99,
		Stock: 5,
	}
	created, _ := svc.Create(req)

	// Try to reduce more than available
	err := svc.UpdateStock(created.ID, -10)
	if err == nil {
		t.Fatal("expected error for insufficient stock, got nil")
	}
}
