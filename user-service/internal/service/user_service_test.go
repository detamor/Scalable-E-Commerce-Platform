package service

import (
	"testing"

	"user-service/internal/model"
	"user-service/internal/repository"
)

const testJWTSecret = "test-secret-key"

func TestRegister_Success(t *testing.T) {
	mockRepo := repository.NewMockUserRepository()
	svc := NewUserService(mockRepo, testJWTSecret)

	req := model.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	resp, err := svc.Register(req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp.Token == "" {
		t.Fatal("expected a JWT token, got empty string")
	}

	if resp.User.Email != req.Email {
		t.Errorf("expected email %s, got %s", req.Email, resp.User.Email)
	}

	if resp.User.Name != req.Name {
		t.Errorf("expected name %s, got %s", req.Name, resp.User.Name)
	}

	// Verify password is hashed (not stored in plain text)
	if resp.User.Password == req.Password {
		t.Error("password should be hashed, not stored in plain text")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	mockRepo := repository.NewMockUserRepository()
	svc := NewUserService(mockRepo, testJWTSecret)

	req := model.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	// First registration should succeed
	_, err := svc.Register(req)
	if err != nil {
		t.Fatalf("first register should succeed, got: %v", err)
	}

	// Second registration with same email should fail
	_, err = svc.Register(req)
	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}

	if err.Error() != "email already registered" {
		t.Errorf("expected 'email already registered', got: %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	mockRepo := repository.NewMockUserRepository()
	svc := NewUserService(mockRepo, testJWTSecret)

	// First register a user
	registerReq := model.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	_, err := svc.Register(registerReq)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Now login
	loginReq := model.LoginRequest{
		Email:    "john@example.com",
		Password: "password123",
	}

	resp, err := svc.Login(loginReq)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp.Token == "" {
		t.Fatal("expected a JWT token, got empty string")
	}

	if resp.User.Email != loginReq.Email {
		t.Errorf("expected email %s, got %s", loginReq.Email, resp.User.Email)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	mockRepo := repository.NewMockUserRepository()
	svc := NewUserService(mockRepo, testJWTSecret)

	// Register
	registerReq := model.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	_, err := svc.Register(registerReq)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Login with wrong password
	loginReq := model.LoginRequest{
		Email:    "john@example.com",
		Password: "wrongpassword",
	}

	_, err = svc.Login(loginReq)
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}

	if err.Error() != "invalid email or password" {
		t.Errorf("expected 'invalid email or password', got: %v", err)
	}
}

func TestLogin_NonExistentUser(t *testing.T) {
	mockRepo := repository.NewMockUserRepository()
	svc := NewUserService(mockRepo, testJWTSecret)

	loginReq := model.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	_, err := svc.Login(loginReq)
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
}

func TestGetProfile_Success(t *testing.T) {
	mockRepo := repository.NewMockUserRepository()
	svc := NewUserService(mockRepo, testJWTSecret)

	// Register a user first
	registerReq := model.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	resp, err := svc.Register(registerReq)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Get profile
	user, err := svc.GetProfile(resp.User.ID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if user.Email != registerReq.Email {
		t.Errorf("expected email %s, got %s", registerReq.Email, user.Email)
	}

	if user.Name != registerReq.Name {
		t.Errorf("expected name %s, got %s", registerReq.Name, user.Name)
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	mockRepo := repository.NewMockUserRepository()
	svc := NewUserService(mockRepo, testJWTSecret)

	_, err := svc.GetProfile(999)
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
}
