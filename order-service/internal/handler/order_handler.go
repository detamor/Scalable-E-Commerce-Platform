package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"order-service/internal/service"
)

// OrderHandler handles HTTP requests for orders.
type OrderHandler struct {
	service service.OrderService
}

// NewOrderHandler creates a new OrderHandler instance.
func NewOrderHandler(svc service.OrderService) *OrderHandler {
	return &OrderHandler{service: svc}
}

// Checkout places an order using the user's current cart.
// POST /api/v1/orders/checkout
func (h *OrderHandler) Checkout(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	email, _ := c.Get("email")
	emailStr, _ := email.(string)

	// Pass the full Authorization header so downstream services can authenticate
	authToken := c.GetHeader("Authorization")

	order, err := h.service.Checkout(userID, emailStr, authToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

// ListOrders returns all orders for the authenticated user.
// GET /api/v1/orders
func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	orders, err := h.service.ListOrders(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": orders})
}

// GetOrder returns a specific order by ID.
// GET /api/v1/orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	order, err := h.service.GetOrder(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}
