// Package config loads application settings from environment variables
// (and an optional .env file) and validates them before the app starts.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// ModeProduction is the APP_MODE value that turns on production behaviour
// (release-mode Gin, JSON logs, HSTS, strict secret checks).
const ModeProduction = "production"

// minJWTSecretLength is the shortest HS256 secret the app accepts (256 bits).
const minJWTSecretLength = 32

// placeholderSecrets are the example values shipped in .env.example and
// docker-compose.yml. They are fine for local development but refused in
// production mode.
var placeholderSecrets = map[string]bool{
	"change-me-to-a-random-string-of-at-least-32-characters": true,
	"dev-only-insecure-jwt-secret-do-not-use-in-production":  true,
}

// Config holds all configuration for the application.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Auth     AuthConfig
}

// AppConfig holds HTTP server and runtime settings.
type AppConfig struct {
	Name           string
	Port           string
	Mode           string
	Version        string
	CORSOrigins    []string
	TrustedProxies []string
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	TimeZone string
}

// JWTConfig holds access-token settings.
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
	Issuer     string
}

// AuthConfig holds password hashing and brute-force protection settings.
type AuthConfig struct {
	BcryptCost         int
	RateLimitPerMinute int
}

// IsProduction reports whether the app runs in production mode.
func (c *Config) IsProduction() bool {
	return c.App.Mode == ModeProduction
}

// Load reads configuration from the environment. Values in a .env file in the
// working directory are used when present; real environment variables win.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()
	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		// A missing .env is normal in containers; anything else is a real error.
		if !errors.As(err, &notFound) && !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("read .env: %w", err)
		}
	}

	expiration, err := time.ParseDuration(v.GetString("JWT_EXPIRATION"))
	if err != nil {
		return nil, fmt.Errorf("JWT_EXPIRATION: %w", err)
	}

	cfg := &Config{
		App: AppConfig{
			Name:           v.GetString("APP_NAME"),
			Port:           v.GetString("APP_PORT"),
			Mode:           v.GetString("APP_MODE"),
			Version:        v.GetString("APP_VERSION"),
			CORSOrigins:    splitList(v.GetString("CORS_ALLOWED_ORIGINS")),
			TrustedProxies: splitList(v.GetString("TRUSTED_PROXIES")),
		},
		Database: DatabaseConfig{
			Host:     v.GetString("DB_HOST"),
			Port:     v.GetString("DB_PORT"),
			User:     v.GetString("DB_USER"),
			Password: v.GetString("DB_PASSWORD"),
			Name:     v.GetString("DB_NAME"),
			SSLMode:  v.GetString("DB_SSL_MODE"),
			TimeZone: v.GetString("DB_TIMEZONE"),
		},
		JWT: JWTConfig{
			Secret:     v.GetString("JWT_SECRET"),
			Expiration: expiration,
			Issuer:     v.GetString("JWT_ISSUER"),
		},
		Auth: AuthConfig{
			BcryptCost:         v.GetInt("BCRYPT_COST"),
			RateLimitPerMinute: v.GetInt("AUTH_RATE_LIMIT_PER_MINUTE"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate checks that the configuration is safe to run with.
func (c *Config) Validate() error {
	var problems []string

	switch {
	case c.JWT.Secret == "":
		problems = append(problems, "JWT_SECRET is required")
	case len(c.JWT.Secret) < minJWTSecretLength:
		problems = append(problems, fmt.Sprintf("JWT_SECRET must be at least %d characters", minJWTSecretLength))
	case c.IsProduction() && placeholderSecrets[c.JWT.Secret]:
		problems = append(problems, "JWT_SECRET is still the example value; set a random secret for production")
	}
	if c.JWT.Expiration <= 0 {
		problems = append(problems, "JWT_EXPIRATION must be positive")
	}
	if c.Auth.BcryptCost < 10 || c.Auth.BcryptCost > 31 {
		problems = append(problems, "BCRYPT_COST must be between 10 and 31")
	}
	if c.Auth.RateLimitPerMinute <= 0 {
		problems = append(problems, "AUTH_RATE_LIMIT_PER_MINUTE must be positive")
	}
	if c.Database.Host == "" || c.Database.Name == "" {
		problems = append(problems, "DB_HOST and DB_NAME are required")
	}

	if len(problems) > 0 {
		return fmt.Errorf("invalid configuration: %s", strings.Join(problems, "; "))
	}
	return nil
}

// DSN returns the PostgreSQL connection URL. Every part is escaped, so
// passwords may contain any character.
func (d *DatabaseConfig) DSN() string {
	q := url.Values{}
	q.Set("sslmode", d.SSLMode)
	if d.TimeZone != "" {
		q.Set("TimeZone", d.TimeZone)
	}
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(d.User, d.Password),
		Host:     d.Host + ":" + d.Port,
		Path:     "/" + d.Name,
		RawQuery: q.Encode(),
	}
	return u.String()
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("APP_NAME", "golang-boilerplate")
	v.SetDefault("APP_PORT", "8080")
	v.SetDefault("APP_MODE", "development")
	v.SetDefault("APP_VERSION", "dev")
	v.SetDefault("CORS_ALLOWED_ORIGINS", "")
	v.SetDefault("TRUSTED_PROXIES", "")

	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "15432")
	v.SetDefault("DB_USER", "postgres")
	v.SetDefault("DB_PASSWORD", "")
	v.SetDefault("DB_NAME", "golang_boilerplate")
	v.SetDefault("DB_SSL_MODE", "disable")
	v.SetDefault("DB_TIMEZONE", "UTC")

	v.SetDefault("JWT_SECRET", "")
	v.SetDefault("JWT_EXPIRATION", "1h")
	v.SetDefault("JWT_ISSUER", "golang-boilerplate")

	v.SetDefault("BCRYPT_COST", 12)
	v.SetDefault("AUTH_RATE_LIMIT_PER_MINUTE", 10)
}

func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
