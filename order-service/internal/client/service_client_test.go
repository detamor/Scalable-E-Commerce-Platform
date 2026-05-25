package client

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker_ClosedState(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 1*time.Second)

	// Successful calls should keep circuit closed
	for i := 0; i < 10; i++ {
		err := cb.Execute(func() error {
			return nil
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	}

	if cb.GetState() != StateClosed {
		t.Errorf("expected state CLOSED, got %s", cb.GetState())
	}
}

func TestCircuitBreaker_OpensAfterMaxFailures(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 1*time.Second)

	testErr := errors.New("service unavailable")

	// Cause 3 failures to trigger circuit open
	for i := 0; i < 3; i++ {
		_ = cb.Execute(func() error {
			return testErr
		})
	}

	if cb.GetState() != StateOpen {
		t.Errorf("expected state OPEN after %d failures, got %s", 3, cb.GetState())
	}

	// Next call should be rejected immediately without executing the function
	executed := false
	err := cb.Execute(func() error {
		executed = true
		return nil
	})

	if err == nil {
		t.Fatal("expected error when circuit is open, got nil")
	}

	if executed {
		t.Error("function should not be executed when circuit is open")
	}
}

func TestCircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 100*time.Millisecond)

	testErr := errors.New("service unavailable")

	// Open the circuit
	for i := 0; i < 3; i++ {
		_ = cb.Execute(func() error {
			return testErr
		})
	}

	if cb.GetState() != StateOpen {
		t.Fatalf("expected OPEN state, got %s", cb.GetState())
	}

	// Wait for timeout to elapse
	time.Sleep(150 * time.Millisecond)

	// Next call should transition to half-open and execute
	err := cb.Execute(func() error {
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error in half-open, got: %v", err)
	}
}

func TestCircuitBreaker_RecoverFromHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 100*time.Millisecond)
	cb.halfOpenMaxCalls = 2

	testErr := errors.New("service unavailable")

	// Open the circuit
	for i := 0; i < 3; i++ {
		_ = cb.Execute(func() error {
			return testErr
		})
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Make enough successful calls to close the circuit
	for i := 0; i < 2; i++ {
		err := cb.Execute(func() error {
			return nil
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	}

	if cb.GetState() != StateClosed {
		t.Errorf("expected state CLOSED after recovery, got %s", cb.GetState())
	}
}

func TestCircuitBreaker_HalfOpenFailureReopens(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 100*time.Millisecond)

	testErr := errors.New("service unavailable")

	// Open the circuit
	for i := 0; i < 3; i++ {
		_ = cb.Execute(func() error {
			return testErr
		})
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Fail during half-open should re-open the circuit
	_ = cb.Execute(func() error {
		return testErr
	})

	if cb.GetState() != StateOpen {
		t.Errorf("expected state OPEN after half-open failure, got %s", cb.GetState())
	}
}

func TestCircuitBreaker_ResetOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 1*time.Second)

	testErr := errors.New("service unavailable")

	// Cause 2 failures (not enough to open)
	for i := 0; i < 2; i++ {
		_ = cb.Execute(func() error {
			return testErr
		})
	}

	// A success should reset the failure counter
	_ = cb.Execute(func() error {
		return nil
	})

	if cb.GetState() != StateClosed {
		t.Errorf("expected CLOSED state, got %s", cb.GetState())
	}

	// Now 2 more failures should not open the circuit (counter was reset)
	for i := 0; i < 2; i++ {
		_ = cb.Execute(func() error {
			return testErr
		})
	}

	if cb.GetState() != StateClosed {
		t.Errorf("expected CLOSED state after reset, got %s", cb.GetState())
	}
}
