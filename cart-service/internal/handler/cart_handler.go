package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"cart-service/internal/model"
	"cart-service/internal/service"
)

// CartHandler handles HTTP requests for shopping carts.
type CartHandler struct {
	service service.CartService
}

// NewCartHandler creates a new CartHandler instance.
func NewCartHandler(svc service.CartService) *CartHandler {
	return &CartHandler{service: svc}
}

// GetCart returns the current user's cart.
// GET /api/v1/cart
func (h *CartHandler) GetCart(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	items, err := h.service.GetCart(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get cart"})
		return
	}

	// Calculate total
	total := 0.0
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": total,
	})
}

// AddItem adds an item to the cart.
// POST /api/v1/cart
func (h *CartHandler) AddItem(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req model.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddItem(c.Request.Context(), userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add item to cart"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "item added to cart"})
}

// UpdateQuantity updates the quantity of an item in the cart.
// PUT /api/v1/cart/:productId
func (h *CartHandler) UpdateQuantity(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	productID, err := strconv.ParseUint(c.Param("productId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	var req model.UpdateCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateQuantity(c.Request.Context(), userID, uint(productID), req.Quantity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "quantity updated"})
}

// RemoveItem removes an item from the cart.
// DELETE /api/v1/cart/:productId
func (h *CartHandler) RemoveItem(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	productID, err := strconv.ParseUint(c.Param("productId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	if err := h.service.RemoveItem(c.Request.Context(), userID, uint(productID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "item removed from cart"})
}

// ClearCart removes all items from the cart.
// DELETE /api/v1/cart
func (h *CartHandler) ClearCart(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	if err := h.service.ClearCart(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cart cleared"})
}
