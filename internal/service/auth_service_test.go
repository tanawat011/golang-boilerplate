package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/tanawat011/golang-boilerplate/internal/auth"
	"github.com/tanawat011/golang-boilerplate/internal/model"
	"github.com/tanawat011/golang-boilerplate/internal/repository/repositorytest"
	"github.com/tanawat011/golang-boilerplate/internal/service"
)

// Test keys are built at run time: a literal would look like a leaked
// secret to secret scanners.
var secret = strings.Repeat("k", 40)

func newAuth(t *testing.T) (service.AuthService, *repositorytest.UserRepository, *auth.TokenManager) {
	t.Helper()
	repo := repositorytest.NewUserRepository()
	tokens := auth.NewTokenManager(secret, time.Hour, "test")
	svc, err := service.NewAuthService(repo, tokens, bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	return svc, repo, tokens
}

func register(t *testing.T, svc service.AuthService, username, email string) *model.UserResponse {
	t.Helper()
	u, err := svc.Register(context.Background(), service.RegisterRequest{
		Username: username, Email: email, Password: "correct-horse",
	})
	if err != nil {
		t.Fatalf("Register(%s): %v", username, err)
	}
	return u
}

func TestRegisterHashesPasswordAndNormalisesEmail(t *testing.T) {
	svc, repo, _ := newAuth(t)

	u := register(t, svc, "alice", "  Alice@Example.COM ")
	if u.Email != "alice@example.com" || u.Role != model.RoleUser || !u.IsActive {
		t.Errorf("unexpected user: %+v", u)
	}

	stored, _ := repo.Get(u.ID)
	if stored.Password == "correct-horse" {
		t.Fatal("password stored in plain text")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte("correct-horse")); err != nil {
		t.Errorf("stored hash does not match: %v", err)
	}
}

func TestRegisterRejectsDuplicates(t *testing.T) {
	svc, _, _ := newAuth(t)
	register(t, svc, "alice", "alice@example.com")
	ctx := context.Background()

	_, err := svc.Register(ctx, service.RegisterRequest{Username: "bob", Email: "ALICE@example.com", Password: "password1"})
	if !errors.Is(err, service.ErrEmailAlreadyExists) {
		t.Errorf("same email: err = %v", err)
	}
	_, err = svc.Register(ctx, service.RegisterRequest{Username: "alice", Email: "other@example.com", Password: "password1"})
	if !errors.Is(err, service.ErrUsernameExists) {
		t.Errorf("same username: err = %v", err)
	}
}

func TestRegisterRefusesPasswordsBcryptWouldTruncate(t *testing.T) {
	svc, _, _ := newAuth(t)
	// 25 three-byte characters: 25 runes, 75 bytes.
	_, err := svc.Register(context.Background(), service.RegisterRequest{
		Username: "alice", Email: "alice@example.com", Password: strings.Repeat("ก", 25),
	})
	if !errors.Is(err, service.ErrPasswordTooLong) {
		t.Errorf("err = %v, want ErrPasswordTooLong", err)
	}
}

func TestLogin(t *testing.T) {
	svc, repo, tokens := newAuth(t)
	u := register(t, svc, "alice", "alice@example.com")
	ctx := context.Background()

	res, err := svc.Login(ctx, service.LoginRequest{Email: "Alice@example.com", Password: "correct-horse"})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	claims, err := tokens.Parse(res.AccessToken)
	if err != nil {
		t.Fatalf("token does not parse: %v", err)
	}
	if claims.Subject != u.ID || claims.Role != model.RoleUser || res.TokenType != "Bearer" {
		t.Errorf("unexpected login result: %+v / %+v", res, claims)
	}

	t.Run("wrong password", func(t *testing.T) {
		_, err := svc.Login(ctx, service.LoginRequest{Email: "alice@example.com", Password: "wrong"})
		if !errors.Is(err, service.ErrInvalidCredentials) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("unknown email gets the same error", func(t *testing.T) {
		_, err := svc.Login(ctx, service.LoginRequest{Email: "nobody@example.com", Password: "correct-horse"})
		if !errors.Is(err, service.ErrInvalidCredentials) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("disabled account", func(t *testing.T) {
		stored, _ := repo.Get(u.ID)
		stored.IsActive = false
		if err := repo.Update(ctx, &stored); err != nil {
			t.Fatal(err)
		}
		_, err := svc.Login(ctx, service.LoginRequest{Email: "alice@example.com", Password: "correct-horse"})
		if !errors.Is(err, service.ErrAccountDisabled) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestRepositoryFailuresAreNotHidden(t *testing.T) {
	svc, repo, _ := newAuth(t)
	boom := errors.New("database is down")
	repo.Err = boom

	if _, err := svc.Register(context.Background(), service.RegisterRequest{
		Username: "alice", Email: "alice@example.com", Password: "password1",
	}); !errors.Is(err, boom) {
		t.Errorf("Register err = %v, want the repository error", err)
	}
	if _, err := svc.Login(context.Background(), service.LoginRequest{
		Email: "alice@example.com", Password: "password1",
	}); !errors.Is(err, boom) {
		t.Errorf("Login err = %v, want the repository error", err)
	}
}
