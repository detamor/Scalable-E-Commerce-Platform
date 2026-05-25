package repository

import (
	"errors"

	"user-service/internal/model"
)

// MockUserRepository is a mock implementation of UserRepository for testing.
type MockUserRepository struct {
	Users    map[string]*model.User // indexed by email
	UserByID map[uint]*model.User   // indexed by ID
	NextID   uint
	CreateFn func(user *model.User) error
}

// NewMockUserRepository creates a new MockUserRepository for testing.
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users:    make(map[string]*model.User),
		UserByID: make(map[uint]*model.User),
		NextID:   1,
	}
}

func (m *MockUserRepository) Create(user *model.User) error {
	if m.CreateFn != nil {
		return m.CreateFn(user)
	}
	if _, exists := m.Users[user.Email]; exists {
		return errors.New("email already exists")
	}
	user.ID = m.NextID
	m.NextID++
	m.Users[user.Email] = user
	m.UserByID[user.ID] = user
	return nil
}

func (m *MockUserRepository) FindByEmail(email string) (*model.User, error) {
	user, exists := m.Users[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MockUserRepository) FindByID(id uint) (*model.User, error) {
	user, exists := m.UserByID[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}
