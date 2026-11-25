package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/tanawat011/golang-boilerplate/internal/auth"
	"github.com/tanawat011/golang-boilerplate/internal/model"
	"github.com/tanawat011/golang-boilerplate/internal/repository"
)

// maxPasswordBytes is bcrypt's input limit; longer passwords are refused
// rather than silently truncated.
const maxPasswordBytes = 72

// RegisterRequest is the body of POST /auth/register.
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Email    string `json:"email" validate:"required,email,max=100"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	FullName string `json:"full_name" validate:"max=100"`
}

// LoginRequest is the body of POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResult is returned after a successful login.
type LoginResult struct {
	AccessToken string             `json:"access_token"`
	TokenType   string             `json:"token_type"`
	ExpiresAt   time.Time          `json:"expires_at"`
	User        model.UserResponse `json:"user"`
}

// AuthService registers users and exchanges credentials for tokens.
type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (*model.UserResponse, error)
	Login(ctx context.Context, req LoginRequest) (*LoginResult, error)
}

type authService struct {
	repo       repository.UserRepository
	tokens     *auth.TokenManager
	bcryptCost int
	dummyHash  []byte
}

// NewAuthService creates an AuthService.
func NewAuthService(repo repository.UserRepository, tokens *auth.TokenManager, bcryptCost int) (AuthService, error) {
	// Compared against when the email is unknown, so a failed login takes
	// the same time whether or not the account exists.
	dummy, err := bcrypt.GenerateFromPassword([]byte("timing-equaliser"), bcryptCost)
	if err != nil {
		return nil, err
	}
	return &authService{repo: repo, tokens: tokens, bcryptCost: bcryptCost, dummyHash: dummy}, nil
}

func (s *authService) Register(ctx context.Context, req RegisterRequest) (*model.UserResponse, error) {
	if len(req.Password) > maxPasswordBytes {
		return nil, ErrPasswordTooLong
	}
	email := normaliseEmail(req.Email)

	if _, err := s.repo.FindByEmail(ctx, email); err == nil {
		return nil, ErrEmailAlreadyExists
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	if _, err := s.repo.FindByUsername(ctx, req.Username); err == nil {
		return nil, ErrUsernameExists
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcryptCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: req.Username,
		Email:    email,
		Password: string(hash),
		FullName: req.FullName,
		Role:     model.RoleUser,
		IsActive: true,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			// Lost a race with another registration, or the address belongs
			// to a deleted account.
			return nil, ErrAlreadyExists
		}
		return nil, err
	}

	resp := user.ToResponse()
	return &resp, nil
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (*LoginResult, error) {
	user, err := s.repo.FindByEmail(ctx, normaliseEmail(req.Email))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(req.Password))
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrAccountDisabled
	}

	token, expiresAt, err := s.tokens.Issue(user.ID, user.Role)
	if err != nil {
		return nil, err
	}
	return &LoginResult{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
		User:        user.ToResponse(),
	}, nil
}

func normaliseEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
