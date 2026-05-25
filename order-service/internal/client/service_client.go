package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"sync"
	"time"
)

// ──────────────────────────────────────────────
// Circuit Breaker Implementation
// ──────────────────────────────────────────────

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	StateClosed   CircuitState = iota // Normal operation
	StateOpen                         // Blocking requests
	StateHalfOpen                     // Testing recovery
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF-OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	mu               sync.Mutex
	state            CircuitState
	failureCount     int
	successCount     int
	maxFailures      int           // Failures before opening circuit
	timeout          time.Duration // How long to stay open
	halfOpenMaxCalls int           // Max test calls in half-open state
	lastFailureTime  time.Time
	name             string
}

// NewCircuitBreaker creates a new CircuitBreaker with configurable thresholds.
func NewCircuitBreaker(name string, maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		state:            StateClosed,
		maxFailures:      maxFailures,
		timeout:          timeout,
		halfOpenMaxCalls: 2,
	}
}

// Execute runs the given function through the circuit breaker.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()

	switch cb.state {
	case StateOpen:
		// Check if timeout has elapsed to transition to half-open
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
			log.Printf("[CircuitBreaker:%s] State: OPEN -> HALF-OPEN", cb.name)
		} else {
			cb.mu.Unlock()
			return fmt.Errorf("[CircuitBreaker:%s] circuit is OPEN - service unavailable, try again later", cb.name)
		}

	case StateHalfOpen:
		// Allow limited requests through
	}

	cb.mu.Unlock()

	// Execute the function
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		cb.lastFailureTime = time.Now()

		if cb.state == StateHalfOpen || cb.failureCount >= cb.maxFailures {
			cb.state = StateOpen
			log.Printf("[CircuitBreaker:%s] State: -> OPEN (failures: %d)", cb.name, cb.failureCount)
		}
		return err
	}

	// Success
	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.halfOpenMaxCalls {
			cb.state = StateClosed
			cb.failureCount = 0
			cb.successCount = 0
			log.Printf("[CircuitBreaker:%s] State: HALF-OPEN -> CLOSED (recovered)", cb.name)
		}
	} else {
		cb.failureCount = 0 // Reset on success in closed state
	}

	return nil
}

// GetState returns the current state of the circuit breaker.
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// ──────────────────────────────────────────────
// Retry with Exponential Backoff
// ──────────────────────────────────────────────

// RetryConfig holds retry configuration.
type RetryConfig struct {
	MaxRetries  int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// DefaultRetryConfig returns sensible default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  200 * time.Millisecond,
		MaxDelay:   2 * time.Second,
	}
}

// withRetry executes a function with retry and exponential backoff.
func withRetry(cfg RetryConfig, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if attempt < cfg.MaxRetries {
			delay := time.Duration(math.Pow(2, float64(attempt))) * cfg.BaseDelay
			if delay > cfg.MaxDelay {
				delay = cfg.MaxDelay
			}
			log.Printf("[Retry] Attempt %d/%d failed: %v. Retrying in %v...", attempt+1, cfg.MaxRetries, lastErr, delay)
			time.Sleep(delay)
		}
	}
	return fmt.Errorf("all %d retries exhausted: %w", cfg.MaxRetries, lastErr)
}

// ──────────────────────────────────────────────
// Data Models
// ──────────────────────────────────────────────

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

// ──────────────────────────────────────────────
// Service Client with Circuit Breaker & Retry
// ──────────────────────────────────────────────

// ServiceClient handles HTTP communication with other microservices.
type ServiceClient struct {
	httpClient        *http.Client
	cartServiceURL    string
	productServiceURL string
	paymentServiceURL string

	// Circuit breakers for each downstream service
	cartBreaker    *CircuitBreaker
	productBreaker *CircuitBreaker
	paymentBreaker *CircuitBreaker

	// Retry configuration
	retryConfig RetryConfig
}

// NewServiceClient creates a new ServiceClient with circuit breakers and retry.
func NewServiceClient(cartURL, productURL, paymentURL string) *ServiceClient {
	return &ServiceClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cartServiceURL:    cartURL,
		productServiceURL: productURL,
		paymentServiceURL: paymentURL,

		// Circuit breakers: open after 5 consecutive failures, reset after 30 seconds
		cartBreaker:    NewCircuitBreaker("cart-service", 5, 30*time.Second),
		productBreaker: NewCircuitBreaker("product-service", 5, 30*time.Second),
		paymentBreaker: NewCircuitBreaker("payment-service", 5, 30*time.Second),

		retryConfig: DefaultRetryConfig(),
	}
}

// GetCart fetches the user's cart from the Cart Service.
func (c *ServiceClient) GetCart(token string) (*CartResponse, error) {
	var cartResp CartResponse

	err := c.cartBreaker.Execute(func() error {
		return withRetry(c.retryConfig, func() error {
			req, err := http.NewRequest("GET", c.cartServiceURL+"/api/v1/cart", nil)
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

			if err := json.NewDecoder(resp.Body).Decode(&cartResp); err != nil {
				return fmt.Errorf("failed to decode cart response: %w", err)
			}

			return nil
		})
	})

	if err != nil {
		return nil, err
	}
	return &cartResp, nil
}

// ClearCart clears the user's cart via the Cart Service.
func (c *ServiceClient) ClearCart(token string) error {
	return c.cartBreaker.Execute(func() error {
		return withRetry(c.retryConfig, func() error {
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
		})
	})
}

// ReduceStock reduces a product's stock via the Product Service.
func (c *ServiceClient) ReduceStock(token string, productID uint, quantity int) error {
	return c.productBreaker.Execute(func() error {
		return withRetry(c.retryConfig, func() error {
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
		})
	})
}

// ProcessPayment sends a payment request to the Payment Service.
func (c *ServiceClient) ProcessPayment(orderID uint, amount float64) (*PaymentResponse, error) {
	var paymentResp PaymentResponse

	err := c.paymentBreaker.Execute(func() error {
		return withRetry(c.retryConfig, func() error {
			payload := PaymentRequest{OrderID: orderID, Amount: amount}
			body, _ := json.Marshal(payload)

			req, err := http.NewRequest("POST", c.paymentServiceURL+"/api/v1/payments/process", bytes.NewBuffer(body))
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := c.httpClient.Do(req)
			if err != nil {
				return fmt.Errorf("failed to reach payment service: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				respBody, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("payment service returned %d: %s", resp.StatusCode, string(respBody))
			}

			if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
				return fmt.Errorf("failed to decode payment response: %w", err)
			}

			return nil
		})
	})

	if err != nil {
		return nil, err
	}
	return &paymentResp, nil
}
