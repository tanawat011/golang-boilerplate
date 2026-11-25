package config

import (
	"strings"
	"testing"
	"time"
)

// Test keys are built at run time: a literal would look like a leaked
// secret to secret scanners.
var goodSecret = strings.Repeat("g", 40)

// isolate runs the test in an empty directory (no .env) with the given env.
func isolate(t *testing.T, env map[string]string) {
	t.Helper()
	t.Chdir(t.TempDir())
	for _, k := range []string{
		"APP_MODE", "JWT_SECRET", "JWT_EXPIRATION", "BCRYPT_COST",
		"AUTH_RATE_LIMIT_PER_MINUTE", "CORS_ALLOWED_ORIGINS", "DB_PASSWORD", "DB_HOST",
	} {
		t.Setenv(k, "")
	}
	for k, v := range env {
		t.Setenv(k, v)
	}
}

func TestLoadDefaults(t *testing.T) {
	isolate(t, map[string]string{"JWT_SECRET": goodSecret, "DB_HOST": "localhost"})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.App.Port != "8080" || cfg.Database.Port != "15432" {
		t.Errorf("unexpected defaults: port %s, db port %s", cfg.App.Port, cfg.Database.Port)
	}
	if cfg.JWT.Expiration != time.Hour {
		t.Errorf("JWT expiration = %v, want 1h", cfg.JWT.Expiration)
	}
	if cfg.Auth.BcryptCost != 12 {
		t.Errorf("bcrypt cost = %d, want 12", cfg.Auth.BcryptCost)
	}
	if len(cfg.App.CORSOrigins) != 0 {
		t.Errorf("CORS origins = %v, want none", cfg.App.CORSOrigins)
	}
}

func TestLoadReadsLists(t *testing.T) {
	isolate(t, map[string]string{
		"JWT_SECRET":           goodSecret,
		"DB_HOST":              "localhost",
		"CORS_ALLOWED_ORIGINS": " https://a.example , https://b.example,,",
	})
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(cfg.App.CORSOrigins, "|"); got != "https://a.example|https://b.example" {
		t.Errorf("CORS origins = %q", got)
	}
}

func TestLoadRejectsUnsafeConfig(t *testing.T) {
	cases := map[string]struct {
		env  map[string]string
		want string
	}{
		"missing secret": {
			env:  map[string]string{},
			want: "JWT_SECRET is required",
		},
		"short secret": {
			env:  map[string]string{"JWT_SECRET": "too-short"},
			want: "at least 32 characters",
		},
		"placeholder in production": {
			env: map[string]string{
				"APP_MODE":   ModeProduction,
				"JWT_SECRET": "change-me-to-a-random-string-of-at-least-32-characters",
			},
			want: "example value",
		},
		"weak bcrypt cost": {
			env:  map[string]string{"JWT_SECRET": goodSecret, "BCRYPT_COST": "4"},
			want: "BCRYPT_COST",
		},
		"bad duration": {
			env:  map[string]string{"JWT_SECRET": goodSecret, "JWT_EXPIRATION": "soon"},
			want: "JWT_EXPIRATION",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			env := map[string]string{"DB_HOST": "localhost"}
			for k, v := range tc.env {
				env[k] = v
			}
			isolate(t, env)
			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Load() error = %v, want it to mention %q", err, tc.want)
			}
		})
	}
}

func TestPlaceholderSecretAllowedInDevelopment(t *testing.T) {
	isolate(t, map[string]string{
		"DB_HOST":    "localhost",
		"JWT_SECRET": "dev-only-insecure-jwt-secret-do-not-use-in-production",
	})
	if _, err := Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
}

func TestDSNEscapesPassword(t *testing.T) {
	d := DatabaseConfig{
		Host: "db", Port: "5432", User: "app", Password: "p@ss word/?#",
		Name: "app", SSLMode: "require", TimeZone: "UTC",
	}
	want := "postgres://app:p%40ss%20word%2F%3F%23@db:5432/app?TimeZone=UTC&sslmode=require"
	if got := d.DSN(); got != want {
		t.Errorf("DSN() = %s\nwant      %s", got, want)
	}
}
