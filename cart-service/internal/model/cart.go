package model

// CartItem represents a single item in the shopping cart.
type CartItem struct {
	ProductID uint    `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

// AddToCartRequest is the payload for adding an item to cart.
type AddToCartRequest struct {
	ProductID uint    `json:"product_id" binding:"required"`
	Name      string  `json:"name" binding:"required"`
	Price     float64 `json:"price" binding:"required,gt=0"`
	Quantity  int     `json:"quantity" binding:"required,gt=0"`
}

// UpdateCartRequest is the payload for updating item quantity.
type UpdateCartRequest struct {
	Quantity int `json:"quantity" binding:"required,gt=0"`
}
