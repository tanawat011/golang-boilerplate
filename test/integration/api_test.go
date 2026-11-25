//go:build integration

// Package integration drives the whole HTTP API against a real PostgreSQL.
package integration

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/tanawat011/golang-boilerplate/internal/auth"
	"github.com/tanawat011/golang-boilerplate/internal/config"
	"github.com/tanawat011/golang-boilerplate/internal/model"
	"github.com/tanawat011/golang-boilerplate/internal/repository"
	"github.com/tanawat011/golang-boilerplate/internal/router"
	"github.com/tanawat011/golang-boilerplate/internal/service"
	"github.com/tanawat011/golang-boilerplate/internal/testutil"
)

func TestAPIFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.NewDatabase(t)
	sqlDB, _ := db.DB()

	cfg := &config.Config{
		App:  config.AppConfig{Mode: "test", Version: "test"},
		JWT:  config.JWTConfig{Secret: strings.Repeat("k", 40), Expiration: time.Hour, Issuer: "it"},
		Auth: config.AuthConfig{BcryptCost: bcrypt.MinCost, RateLimitPerMinute: 1000},
	}
	repo := repository.NewUserRepository(db)
	tokens := auth.NewTokenManager(cfg.JWT.Secret, cfg.JWT.Expiration, cfg.JWT.Issuer)
	authSvc, err := service.NewAuthService(repo, tokens, cfg.Auth.BcryptCost)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := router.New(router.Deps{
		Config: cfg, Logger: zap.NewNop(), Tokens: tokens,
		Auth: authSvc, Users: service.NewUserService(repo), DB: sqlDB,
	})
	if err != nil {
		t.Fatal(err)
	}

	call := func(method, path, token string, body any) (int, map[string]any) {
		t.Helper()
		var buf bytes.Buffer
		if body != nil {
			_ = json.NewEncoder(&buf).Encode(body)
		}
		req := httptest.NewRequest(method, path, &buf)
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)
		var out map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &out)
		return rec.Code, out
	}
	data := func(body map[string]any) map[string]any { return body["data"].(map[string]any) }

	if code, body := call("GET", "/api/v1/health", "", nil); code != 200 || body["database"] != "up" {
		t.Fatalf("health: %d %v", code, body)
	}

	for _, name := range []string{"alice", "root"} {
		code, body := call("POST", "/api/v1/auth/register", "", map[string]string{
			"username": name, "email": name + "@example.com", "password": "correct-horse",
		})
		if code != 201 {
			t.Fatalf("register %s: %d %v", name, code, body)
		}
	}
	if code, _ := call("POST", "/api/v1/auth/register", "", map[string]string{
		"username": "alice", "email": "other@example.com", "password": "correct-horse",
	}); code != 409 {
		t.Errorf("duplicate username: %d", code)
	}

	// Promote root directly in the database, then log in to get an admin token.
	if err := db.Model(&model.User{}).Where("username = ?", "root").Update("role", model.RoleAdmin).Error; err != nil {
		t.Fatal(err)
	}
	login := func(email string) string {
		code, body := call("POST", "/api/v1/auth/login", "", map[string]string{"email": email, "password": "correct-horse"})
		if code != 200 {
			t.Fatalf("login %s: %d %v", email, code, body)
		}
		return data(body)["access_token"].(string)
	}
	alice, admin := login("alice@example.com"), login("root@example.com")

	code, body := call("GET", "/api/v1/auth/me", alice, nil)
	if code != 200 || data(body)["username"] != "alice" {
		t.Fatalf("me: %d %v", code, body)
	}
	aliceID := data(body)["id"].(string)

	if code, _ := call("GET", "/api/v1/users", alice, nil); code != 403 {
		t.Errorf("user lists users: %d", code)
	}
	code, body = call("GET", "/api/v1/users", admin, nil)
	if code != 200 || body["meta"].(map[string]any)["total"].(float64) != 2 {
		t.Errorf("admin list: %d %v", code, body)
	}

	code, body = call("PATCH", "/api/v1/users/"+aliceID, admin, map[string]any{"is_active": false})
	if code != 200 || data(body)["is_active"] != false {
		t.Fatalf("deactivate: %d %v", code, body)
	}
	if code, _ := call("POST", "/api/v1/auth/login", "", map[string]string{
		"email": "alice@example.com", "password": "correct-horse",
	}); code != 403 {
		t.Errorf("disabled account logs in: %d", code)
	}

	if code, _ := call("DELETE", "/api/v1/users/"+aliceID, admin, nil); code != 200 {
		t.Errorf("delete: %d", code)
	}
	if code, _ := call("GET", "/api/v1/users/"+aliceID, admin, nil); code != 404 {
		t.Errorf("deleted user: %d", code)
	}
}
