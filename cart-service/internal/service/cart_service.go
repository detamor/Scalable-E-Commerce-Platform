package service

import (
	"context"

	"cart-service/internal/model"
	"cart-service/internal/repository"
)

// CartService defines business logic for shopping carts.
type CartService interface {
	GetCart(ctx context.Context, userID uint) ([]model.CartItem, error)
	AddItem(ctx context.Context, userID uint, req model.AddToCartRequest) error
	UpdateQuantity(ctx context.Context, userID uint, productID uint, quantity int) error
	RemoveItem(ctx context.Context, userID uint, productID uint) error
	ClearCart(ctx context.Context, userID uint) error
}

type cartService struct {
	repo repository.CartRepository
}

// NewCartService creates a new CartService instance.
func NewCartService(repo repository.CartRepository) CartService {
	return &cartService{repo: repo}
}

func (s *cartService) GetCart(ctx context.Context, userID uint) ([]model.CartItem, error) {
	return s.repo.GetCart(ctx, userID)
}

func (s *cartService) AddItem(ctx context.Context, userID uint, req model.AddToCartRequest) error {
	item := model.CartItem{
		ProductID: req.ProductID,
		Name:      req.Name,
		Price:     req.Price,
		Quantity:  req.Quantity,
	}
	return s.repo.AddItem(ctx, userID, item)
}

func (s *cartService) UpdateQuantity(ctx context.Context, userID uint, productID uint, quantity int) error {
	return s.repo.UpdateQuantity(ctx, userID, productID, quantity)
}

func (s *cartService) RemoveItem(ctx context.Context, userID uint, productID uint) error {
	return s.repo.RemoveItem(ctx, userID, productID)
}

func (s *cartService) ClearCart(ctx context.Context, userID uint) error {
	return s.repo.ClearCart(ctx, userID)
}
