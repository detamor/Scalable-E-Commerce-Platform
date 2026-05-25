package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	"cart-service/internal/model"
)

// CartRepository defines data access operations for shopping carts.
type CartRepository interface {
	GetCart(ctx context.Context, userID uint) ([]model.CartItem, error)
	AddItem(ctx context.Context, userID uint, item model.CartItem) error
	UpdateQuantity(ctx context.Context, userID uint, productID uint, quantity int) error
	RemoveItem(ctx context.Context, userID uint, productID uint) error
	ClearCart(ctx context.Context, userID uint) error
}

type cartRepository struct {
	client *redis.Client
}

// NewCartRepository creates a new CartRepository instance.
func NewCartRepository(client *redis.Client) CartRepository {
	return &cartRepository{client: client}
}

func cartKey(userID uint) string {
	return fmt.Sprintf("cart:%d", userID)
}

func (r *cartRepository) GetCart(ctx context.Context, userID uint) ([]model.CartItem, error) {
	data, err := r.client.Get(ctx, cartKey(userID)).Bytes()
	if err == redis.Nil {
		return []model.CartItem{}, nil
	}
	if err != nil {
		return nil, err
	}

	var items []model.CartItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *cartRepository) AddItem(ctx context.Context, userID uint, item model.CartItem) error {
	items, err := r.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	// Check if item already exists, if so update quantity
	found := false
	for i, existing := range items {
		if existing.ProductID == item.ProductID {
			items[i].Quantity += item.Quantity
			found = true
			break
		}
	}

	if !found {
		items = append(items, item)
	}

	return r.saveCart(ctx, userID, items)
}

func (r *cartRepository) UpdateQuantity(ctx context.Context, userID uint, productID uint, quantity int) error {
	items, err := r.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	for i, item := range items {
		if item.ProductID == productID {
			items[i].Quantity = quantity
			return r.saveCart(ctx, userID, items)
		}
	}

	return fmt.Errorf("product %d not found in cart", productID)
}

func (r *cartRepository) RemoveItem(ctx context.Context, userID uint, productID uint) error {
	items, err := r.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	filtered := make([]model.CartItem, 0)
	for _, item := range items {
		if item.ProductID != productID {
			filtered = append(filtered, item)
		}
	}

	return r.saveCart(ctx, userID, filtered)
}

func (r *cartRepository) ClearCart(ctx context.Context, userID uint) error {
	return r.client.Del(ctx, cartKey(userID)).Err()
}

func (r *cartRepository) saveCart(ctx context.Context, userID uint, items []model.CartItem) error {
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, cartKey(userID), data, 0).Err()
}
