package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CartItem represents an item returned by the Cart Service.
type CartItem struct {
	ProductID uint    `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

// CartResponse is the response from Cart Service GET /api/v1/cart.
type CartResponse struct {
	Items []CartItem `json:"items"`
	Total float64    `json:"total"`
}

// PaymentRequest is the payload sent to Payment Service.
type PaymentRequest struct {
	OrderID uint    `json:"order_id"`
	Amount  float64 `json:"amount"`
}

// PaymentResponse is the response from Payment Service.
type PaymentResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}

// StockUpdateRequest is the payload sent to Product Service for stock updates.
type StockUpdateRequest struct {
	Quantity int `json:"quantity"`
}

// ServiceClient handles HTTP communication with other microservices.
type ServiceClient struct {
	httpClient        *http.Client
	cartServiceURL    string
	productServiceURL string
	paymentServiceURL string
}

// NewServiceClient creates a new ServiceClient instance.
func NewServiceClient(cartURL, productURL, paymentURL string) *ServiceClient {
	return &ServiceClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cartServiceURL:    cartURL,
		productServiceURL: productURL,
		paymentServiceURL: paymentURL,
	}
}

// GetCart fetches the user's cart from the Cart Service.
func (c *ServiceClient) GetCart(token string) (*CartResponse, error) {
	req, err := http.NewRequest("GET", c.cartServiceURL+"/api/v1/cart", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach cart service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cart service returned %d: %s", resp.StatusCode, string(body))
	}

	var cartResp CartResponse
	if err := json.NewDecoder(resp.Body).Decode(&cartResp); err != nil {
		return nil, fmt.Errorf("failed to decode cart response: %w", err)
	}

	return &cartResp, nil
}

// ClearCart clears the user's cart via the Cart Service.
func (c *ServiceClient) ClearCart(token string) error {
	req, err := http.NewRequest("DELETE", c.cartServiceURL+"/api/v1/cart", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reach cart service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cart service returned %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ReduceStock reduces a product's stock via the Product Service.
func (c *ServiceClient) ReduceStock(token string, productID uint, quantity int) error {
	payload := StockUpdateRequest{Quantity: -quantity} // Negative to reduce
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/api/v1/products/%d/stock", c.productServiceURL, productID)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reach product service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("product service returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// ProcessPayment sends a payment request to the Payment Service.
func (c *ServiceClient) ProcessPayment(orderID uint, amount float64) (*PaymentResponse, error) {
	payload := PaymentRequest{OrderID: orderID, Amount: amount}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", c.paymentServiceURL+"/api/v1/payments/process", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach payment service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("payment service returned %d: %s", resp.StatusCode, string(respBody))
	}

	var paymentResp PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return nil, fmt.Errorf("failed to decode payment response: %w", err)
	}

	return &paymentResp, nil
}
