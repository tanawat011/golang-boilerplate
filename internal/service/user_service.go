package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/tanawat011/golang-boilerplate/internal/model"
	"github.com/tanawat011/golang-boilerplate/internal/repository"
)

// Common errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"max=100"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Username string `json:"username" validate:"omitempty,min=3,max=50"`
	Email    string `json:"email" validate:"omitempty,email"`
	FullName string `json:"full_name" validate:"omitempty,max=100"`
	Role     string `json:"role" validate:"omitempty,oneof=admin user"`
	IsActive *bool  `json:"is_active"`
}

// UserService defines the interface for user business logic
type UserService interface {
	Create(req CreateUserRequest) (*model.UserResponse, error)
	GetByID(id string) (*model.UserResponse, error)
	GetAll(page, limit int) ([]model.UserResponse, int64, error)
	Update(id string, req UpdateUserRequest) (*model.UserResponse, error)
	Delete(id string) error
}

type userService struct {
	repo repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

// Create registers a new user with hashed password
func (s *userService) Create(req CreateUserRequest) (*model.UserResponse, error) {
	// Check if email already exists
	existing, _ := s.repo.FindByEmail(req.Email)
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Check if username already exists
	existing, _ = s.repo.FindByUsername(req.Username)
	if existing != nil {
		return nil, ErrUsernameExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		FullName: req.FullName,
		Role:     "user",
		IsActive: true,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

// GetByID retrieves a user by their ID
func (s *userService) GetByID(id string) (*model.UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	response := user.ToResponse()
	return &response, nil
}

// GetAll retrieves all users with pagination
func (s *userService) GetAll(page, limit int) ([]model.UserResponse, int64, error) {
	offset := (page - 1) * limit

	users, total, err := s.repo.FindAll(offset, limit)
	if err != nil {
		return nil, 0, err
	}

	var responses []model.UserResponse
	for _, user := range users {
		responses = append(responses, user.ToResponse())
	}

	return responses, total, nil
}

// Update modifies an existing user
func (s *userService) Update(id string, req UpdateUserRequest) (*model.UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if req.Username != "" {
		existing, _ := s.repo.FindByUsername(req.Username)
		if existing != nil && existing.ID != id {
			return nil, ErrUsernameExists
		}
		user.Username = req.Username
	}

	if req.Email != "" {
		existing, _ := s.repo.FindByEmail(req.Email)
		if existing != nil && existing.ID != id {
			return nil, ErrEmailAlreadyExists
		}
		user.Email = req.Email
	}

	if req.FullName != "" {
		user.FullName = req.FullName
	}

	if req.Role != "" {
		user.Role = req.Role
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

// Delete removes a user by their ID
func (s *userService) Delete(id string) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return ErrUserNotFound
	}

	return s.repo.Delete(id)
}
