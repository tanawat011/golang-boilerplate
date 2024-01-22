package config

import (
	"log"
	"sync"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name    string `mapstructure:"APP_NAME"`
	Port    string `mapstructure:"APP_PORT"`
	Mode    string `mapstructure:"APP_MODE"`
	Version string `mapstructure:"APP_VERSION"`
}

// DatabaseConfig holds database-specific configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	Name     string `mapstructure:"DB_NAME"`
	SSLMode  string `mapstructure:"DB_SSL_MODE"`
}

// RedisConfig holds Redis-specific configuration
type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     string `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

// JWTConfig holds JWT-specific configuration
type JWTConfig struct {
	Secret          string `mapstructure:"JWT_SECRET"`
	ExpirationHours int    `mapstructure:"JWT_EXPIRATION_HOURS"`
	Issuer          string `mapstructure:"JWT_ISSUER"`
}

var (
	once     sync.Once
	instance *Config
)

// Load reads configuration from environment variables and .env file
func Load() *Config {
	once.Do(func() {
		viper.SetConfigFile(".env")
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err != nil {
			log.Printf("Warning: .env file not found, using environment variables: %v", err)
		}

		// Set default values
		setDefaults()

		instance = &Config{
			App: AppConfig{
				Name:    viper.GetString("APP_NAME"),
				Port:    viper.GetString("APP_PORT"),
				Mode:    viper.GetString("APP_MODE"),
				Version: viper.GetString("APP_VERSION"),
			},
			Database: DatabaseConfig{
				Host:     viper.GetString("DB_HOST"),
				Port:     viper.GetString("DB_PORT"),
				User:     viper.GetString("DB_USER"),
				Password: viper.GetString("DB_PASSWORD"),
				Name:     viper.GetString("DB_NAME"),
				SSLMode:  viper.GetString("DB_SSL_MODE"),
			},
			Redis: RedisConfig{
				Host:     viper.GetString("REDIS_HOST"),
				Port:     viper.GetString("REDIS_PORT"),
				Password: viper.GetString("REDIS_PASSWORD"),
				DB:       viper.GetInt("REDIS_DB"),
			},
			JWT: JWTConfig{
				Secret:          viper.GetString("JWT_SECRET"),
				ExpirationHours: viper.GetInt("JWT_EXPIRATION_HOURS"),
				Issuer:          viper.GetString("JWT_ISSUER"),
			},
		}
	})

	return instance
}

// setDefaults sets default values for configuration
func setDefaults() {
	viper.SetDefault("APP_NAME", "golang-boilerplate")
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("APP_MODE", "development")
	viper.SetDefault("APP_VERSION", "1.0.0")

	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "golang_boilerplate")
	viper.SetDefault("DB_SSL_MODE", "disable")

	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)

	viper.SetDefault("JWT_SECRET", "your-secret-key-change-in-production")
	viper.SetDefault("JWT_EXPIRATION_HOURS", 24)
	viper.SetDefault("JWT_ISSUER", "golang-boilerplate")
}

// GetDSN returns the database connection string
func (d *DatabaseConfig) GetDSN() string {
	return "host=" + d.Host +
		" user=" + d.User +
		" password=" + d.Password +
		" dbname=" + d.Name +
		" port=" + d.Port +
		" sslmode=" + d.SSLMode +
		" TimeZone=Asia/Bangkok"
}
