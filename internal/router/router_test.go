package router_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
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
	"github.com/tanawat011/golang-boilerplate/internal/repository/repositorytest"
	"github.com/tanawat011/golang-boilerplate/internal/router"
	"github.com/tanawat011/golang-boilerplate/internal/service"
)

type pinger struct{ err error }

func (p *pinger) PingContext(context.Context) error { return p.err }

type app struct {
	t      *testing.T
	engine *gin.Engine
	repo   *repositorytest.UserRepository
	db     *pinger
}

func newApp(t *testing.T, mutate ...func(*config.Config)) *app {
	t.Helper()
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		App:  config.AppConfig{Mode: "test", Version: "test"},
		JWT:  config.JWTConfig{Secret: strings.Repeat("k", 40), Expiration: time.Hour, Issuer: "test"},
		Auth: config.AuthConfig{BcryptCost: bcrypt.MinCost, RateLimitPerMinute: 100},
	}
	for _, m := range mutate {
		m(cfg)
	}

	repo := repositorytest.NewUserRepository()
	tokens := auth.NewTokenManager(cfg.JWT.Secret, cfg.JWT.Expiration, cfg.JWT.Issuer)
	authSvc, err := service.NewAuthService(repo, tokens, cfg.Auth.BcryptCost)
	if err != nil {
		t.Fatal(err)
	}
	db := &pinger{}
	engine, err := router.New(router.Deps{
		Config: cfg,
		Logger: zap.NewNop(),
		Tokens: tokens,
		Auth:   authSvc,
		Users:  service.NewUserService(repo),
		DB:     db,
	})
	if err != nil {
		t.Fatal(err)
	}
	return &app{t: t, engine: engine, repo: repo, db: db}
}

type reply struct {
	Code   int
	Header http.Header
	Body   map[string]any
}

func (a *app) do(method, path, token string, body any, headers ...string) reply {
	a.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if s, ok := body.(string); ok {
			buf.WriteString(s)
		} else if err := json.NewEncoder(&buf).Encode(body); err != nil {
			a.t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	rec := httptest.NewRecorder()
	a.engine.ServeHTTP(rec, req)

	out := reply{Code: rec.Code, Header: rec.Header()}
	if rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &out.Body); err != nil {
			a.t.Fatalf("%s %s: response is not JSON: %s", method, path, rec.Body.String())
		}
	}
	return out
}

// signUp registers and logs in a user, returning its ID and token.
func (a *app) signUp(username string) (id, token string) {
	a.t.Helper()
	email := username + "@example.com"
	r := a.do("POST", "/api/v1/auth/register", "", map[string]string{
		"username": username, "email": email, "password": "correct-horse",
	})
	if r.Code != http.StatusCreated {
		a.t.Fatalf("register %s: %d %v", username, r.Code, r.Body)
	}
	id = r.Body["data"].(map[string]any)["id"].(string)

	r = a.do("POST", "/api/v1/auth/login", "", map[string]string{"email": email, "password": "correct-horse"})
	if r.Code != http.StatusOK {
		a.t.Fatalf("login %s: %d %v", username, r.Code, r.Body)
	}
	return id, r.Body["data"].(map[string]any)["access_token"].(string)
}

// adminToken creates an admin and returns its token (roles live in the token,
// so the login happens after the promotion).
func (a *app) adminToken() string {
	a.t.Helper()
	id, _ := a.signUp("root")
	a.repo.SetRole(id, model.RoleAdmin)
	r := a.do("POST", "/api/v1/auth/login", "", map[string]string{"email": "root@example.com", "password": "correct-horse"})
	return r.Body["data"].(map[string]any)["access_token"].(string)
}

func TestHealth(t *testing.T) {
	a := newApp(t)
	if r := a.do("GET", "/api/v1/health", "", nil); r.Code != http.StatusOK || r.Body["database"] != "up" {
		t.Errorf("healthy: %d %v", r.Code, r.Body)
	}
	a.db.err = errors.New("connection refused")
	if r := a.do("GET", "/api/v1/health", "", nil); r.Code != http.StatusServiceUnavailable {
		t.Errorf("db down: %d %v", r.Code, r.Body)
	}
}

func TestRegisterValidation(t *testing.T) {
	a := newApp(t)

	r := a.do("POST", "/api/v1/auth/register", "", map[string]string{"username": "x", "email": "nope", "password": "short"})
	if r.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid body: %d %v", r.Code, r.Body)
	}
	fields := map[string]bool{}
	for _, e := range r.Body["error"].([]any) {
		fields[e.(map[string]any)["field"].(string)] = true
	}
	for _, f := range []string{"username", "email", "password"} {
		if !fields[f] {
			t.Errorf("no validation error for %q: %v", f, r.Body["error"])
		}
	}

	if r := a.do("POST", "/api/v1/auth/register", "", "{not json"); r.Code != http.StatusBadRequest {
		t.Errorf("malformed JSON: %d", r.Code)
	}

	a.signUp("alice")
	r = a.do("POST", "/api/v1/auth/register", "", map[string]string{
		"username": "alice2", "email": "alice@example.com", "password": "correct-horse",
	})
	if r.Code != http.StatusConflict {
		t.Errorf("duplicate email: %d %v", r.Code, r.Body)
	}
}

func TestRegisterResponseHasNoPassword(t *testing.T) {
	a := newApp(t)
	r := a.do("POST", "/api/v1/auth/register", "", map[string]string{
		"username": "alice", "email": "alice@example.com", "password": "correct-horse",
	})
	for key := range r.Body["data"].(map[string]any) {
		if strings.Contains(strings.ToLower(key), "password") {
			t.Errorf("response leaks %q", key)
		}
	}
}

func TestLoginAndMe(t *testing.T) {
	a := newApp(t)
	id, token := a.signUp("alice")

	r := a.do("POST", "/api/v1/auth/login", "", map[string]string{"email": "alice@example.com", "password": "wrong-password"})
	if r.Code != http.StatusUnauthorized {
		t.Errorf("wrong password: %d", r.Code)
	}

	if r := a.do("GET", "/api/v1/auth/me", "", nil); r.Code != http.StatusUnauthorized {
		t.Errorf("no token: %d", r.Code)
	}
	if r := a.do("GET", "/api/v1/auth/me", "not-a-token", nil); r.Code != http.StatusUnauthorized {
		t.Errorf("bad token: %d", r.Code)
	}
	r = a.do("GET", "/api/v1/auth/me", token, nil)
	if r.Code != http.StatusOK || r.Body["data"].(map[string]any)["id"] != id {
		t.Errorf("me: %d %v", r.Code, r.Body)
	}
}

func TestUserRoutesEnforceOwnershipAndRoles(t *testing.T) {
	a := newApp(t)
	aliceID, alice := a.signUp("alice")
	bobID, _ := a.signUp("bob")
	admin := a.adminToken()

	checks := []struct {
		name         string
		method, path string
		token        string
		body         any
		want         int
	}{
		{"list needs a token", "GET", "/api/v1/users", "", nil, http.StatusUnauthorized},
		{"list is admin only", "GET", "/api/v1/users", alice, nil, http.StatusForbidden},
		{"admin lists", "GET", "/api/v1/users?page=1&limit=2", admin, nil, http.StatusOK},
		{"read self", "GET", "/api/v1/users/" + aliceID, alice, nil, http.StatusOK},
		{"read someone else", "GET", "/api/v1/users/" + bobID, alice, nil, http.StatusForbidden},
		{"admin reads anyone", "GET", "/api/v1/users/" + bobID, admin, nil, http.StatusOK},
		{"not a UUID", "GET", "/api/v1/users/42", admin, nil, http.StatusNotFound},
		{"self-promotion", "PATCH", "/api/v1/users/" + aliceID, alice, map[string]string{"role": "admin"}, http.StatusForbidden},
		{"bad role value", "PATCH", "/api/v1/users/" + bobID, admin, map[string]string{"role": "root"}, http.StatusUnprocessableEntity},
		{"edit self", "PATCH", "/api/v1/users/" + aliceID, alice, map[string]string{"full_name": "Alice"}, http.StatusOK},
		{"admin promotes", "PATCH", "/api/v1/users/" + bobID, admin, map[string]string{"role": "admin"}, http.StatusOK},
		{"delete someone else", "DELETE", "/api/v1/users/" + bobID, alice, nil, http.StatusForbidden},
		{"delete self", "DELETE", "/api/v1/users/" + aliceID, alice, nil, http.StatusOK},
		{"deleted is gone", "GET", "/api/v1/users/" + aliceID, admin, nil, http.StatusNotFound},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if r := a.do(c.method, c.path, c.token, c.body); r.Code != c.want {
				t.Errorf("%s %s: got %d, want %d (%v)", c.method, c.path, r.Code, c.want, r.Body)
			}
		})
	}

	r := a.do("GET", "/api/v1/users?page=1&limit=2", admin, nil)
	meta := r.Body["meta"].(map[string]any)
	if meta["total"].(float64) != 2 || meta["total_pages"].(float64) != 1 {
		t.Errorf("pagination meta after one delete: %v", meta)
	}
}

func TestSecurityHeaders(t *testing.T) {
	r := newApp(t).do("GET", "/api/v1/health", "", nil)
	for h, want := range map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "no-referrer",
	} {
		if got := r.Header.Get(h); got != want {
			t.Errorf("%s = %q, want %q", h, got, want)
		}
	}
	if r.Header.Get("Strict-Transport-Security") != "" {
		t.Error("HSTS sent outside production")
	}

	prod := newApp(t, func(c *config.Config) { c.App.Mode = config.ModeProduction })
	if prod.do("GET", "/api/v1/health", "", nil).Header.Get("Strict-Transport-Security") == "" {
		t.Error("no HSTS in production")
	}
}

func TestCORS(t *testing.T) {
	closed := newApp(t)
	r := closed.do("GET", "/api/v1/health", "", nil, "Origin", "https://evil.example")
	if got := r.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("no origins configured, but Allow-Origin = %q", got)
	}

	open := newApp(t, func(c *config.Config) { c.App.CORSOrigins = []string{"https://app.example"} })
	r = open.do("GET", "/api/v1/health", "", nil, "Origin", "https://app.example")
	if got := r.Header.Get("Access-Control-Allow-Origin"); got != "https://app.example" {
		t.Errorf("allowed origin: Allow-Origin = %q", got)
	}
	if got := r.Header.Get("Access-Control-Allow-Credentials"); got != "" {
		t.Errorf("credentials allowed: %q", got)
	}
	r = open.do("GET", "/api/v1/health", "", nil, "Origin", "https://evil.example")
	if r.Code != http.StatusForbidden {
		t.Errorf("unknown origin: %d", r.Code)
	}
}

func TestAuthRoutesAreRateLimited(t *testing.T) {
	a := newApp(t, func(c *config.Config) { c.Auth.RateLimitPerMinute = 3 })
	body := map[string]string{"email": "nobody@example.com", "password": "whatever"}

	for i := range 3 {
		if r := a.do("POST", "/api/v1/auth/login", "", body); r.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: %d", i+1, r.Code)
		}
	}
	r := a.do("POST", "/api/v1/auth/login", "", body)
	if r.Code != http.StatusTooManyRequests || r.Header.Get("Retry-After") == "" {
		t.Errorf("4th attempt: %d, Retry-After %q", r.Code, r.Header.Get("Retry-After"))
	}
	// Health is not rate limited.
	if r := a.do("GET", "/api/v1/health", "", nil); r.Code != http.StatusOK {
		t.Errorf("health after limit: %d", r.Code)
	}
}

func TestBodyLimit(t *testing.T) {
	a := newApp(t)
	huge := `{"username":"alice","email":"a@example.com","password":"` + strings.Repeat("x", 2<<20) + `"}`
	if r := a.do("POST", "/api/v1/auth/register", "", huge); r.Code != http.StatusBadRequest {
		t.Errorf("2 MiB body: %d", r.Code)
	}
}

func TestUnknownRoute(t *testing.T) {
	if r := newApp(t).do("GET", "/nope", "", nil); r.Code != http.StatusNotFound {
		t.Errorf("unknown route: %d", r.Code)
	}
}
