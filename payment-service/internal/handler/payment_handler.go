package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"payment-service/internal/model"
	"payment-service/internal/service"
)

// PaymentHandler handles HTTP requests for payments.
type PaymentHandler struct {
	service service.PaymentService
}

// NewPaymentHandler creates a new PaymentHandler instance.
func NewPaymentHandler(svc service.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: svc}
}

// ProcessPayment handles payment processing.
// POST /api/v1/payments/process
func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	var req model.ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.ProcessPayment(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
