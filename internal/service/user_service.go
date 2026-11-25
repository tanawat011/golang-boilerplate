package service

import (
	"context"
	"errors"

	"github.com/tanawat011/golang-boilerplate/internal/model"
	"github.com/tanawat011/golang-boilerplate/internal/repository"
)

// UpdateUserRequest is the body of PATCH /users/:id. Empty fields are left
// unchanged. Role and IsActive may only be changed by an admin.
type UpdateUserRequest struct {
	Username string  `json:"username" validate:"omitempty,min=3,max=50,alphanum"`
	Email    string  `json:"email" validate:"omitempty,email,max=100"`
	FullName *string `json:"full_name" validate:"omitempty,max=100"`
	Role     string  `json:"role" validate:"omitempty,oneof=admin user"`
	IsActive *bool   `json:"is_active"`
}

// UserService manages user accounts.
type UserService interface {
	GetByID(ctx context.Context, actor Actor, id string) (*model.UserResponse, error)
	List(ctx context.Context, actor Actor, page, limit int) ([]model.UserResponse, int64, error)
	Update(ctx context.Context, actor Actor, id string, req UpdateUserRequest) (*model.UserResponse, error)
	Delete(ctx context.Context, actor Actor, id string) error
}

type userService struct {
	repo repository.UserRepository
}

// NewUserService creates a UserService.
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

// GetByID returns a user. Users may read themselves; admins may read anyone.
func (s *userService) GetByID(ctx context.Context, actor Actor, id string) (*model.UserResponse, error) {
	if !actor.canAccess(id) {
		return nil, ErrForbidden
	}
	user, err := s.find(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := user.ToResponse()
	return &resp, nil
}

// List returns a page of users, newest first. Admin only.
func (s *userService) List(ctx context.Context, actor Actor, page, limit int) ([]model.UserResponse, int64, error) {
	if !actor.IsAdmin() {
		return nil, 0, ErrForbidden
	}
	users, total, err := s.repo.FindAll(ctx, (page-1)*limit, limit)
	if err != nil {
		return nil, 0, err
	}
	out := make([]model.UserResponse, 0, len(users))
	for i := range users {
		out = append(out, users[i].ToResponse())
	}
	return out, total, nil
}

// Update changes a user. Users may change their own profile; only admins may
// change roles or activate / deactivate accounts.
func (s *userService) Update(ctx context.Context, actor Actor, id string, req UpdateUserRequest) (*model.UserResponse, error) {
	if !actor.canAccess(id) {
		return nil, ErrForbidden
	}
	if (req.Role != "" || req.IsActive != nil) && !actor.IsAdmin() {
		return nil, ErrForbidden
	}

	user, err := s.find(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Username != "" && req.Username != user.Username {
		if err := s.ensureFree(ctx, s.repo.FindByUsername, req.Username, id, ErrUsernameExists); err != nil {
			return nil, err
		}
		user.Username = req.Username
	}
	if req.Email != "" {
		email := normaliseEmail(req.Email)
		if email != user.Email {
			if err := s.ensureFree(ctx, s.repo.FindByEmail, email, id, ErrEmailAlreadyExists); err != nil {
				return nil, err
			}
			user.Email = email
		}
	}
	if req.FullName != nil {
		user.FullName = *req.FullName
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, user); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}
	resp := user.ToResponse()
	return &resp, nil
}

// Delete soft-deletes a user. Users may delete themselves; admins anyone.
func (s *userService) Delete(ctx context.Context, actor Actor, id string) error {
	if !actor.canAccess(id) {
		return ErrForbidden
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

func (s *userService) find(ctx context.Context, id string) (*model.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *userService) ensureFree(
	ctx context.Context,
	find func(context.Context, string) (*model.User, error),
	value, selfID string,
	taken error,
) error {
	existing, err := find(ctx, value)
	switch {
	case err == nil && existing.ID != selfID:
		return taken
	case err == nil, errors.Is(err, repository.ErrNotFound):
		return nil
	default:
		return err
	}
}
