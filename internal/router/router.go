package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/tanawat011/golang-boilerplate/internal/config"
	"github.com/tanawat011/golang-boilerplate/internal/handler"
	"github.com/tanawat011/golang-boilerplate/internal/middleware"
	"github.com/tanawat011/golang-boilerplate/internal/repository"
	"github.com/tanawat011/golang-boilerplate/internal/service"
)

// Setup initializes the Gin router with all routes and middleware
func Setup(cfg *config.Config, db *gorm.DB, logger *zap.Logger) *gin.Engine {
	// Set Gin mode based on configuration
	if cfg.App.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middleware
	r.Use(middleware.CORS())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recovery(logger))

	// Initialize layers
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(cfg.App.Version)
	userHandler := handler.NewUserHandler(userService)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Health check (public)
		v1.GET("/health", healthHandler.Check)

		// User routes
		users := v1.Group("/users")
		{
			users.POST("", userHandler.Create)
			users.GET("", userHandler.GetAll)
			users.GET("/:id", userHandler.GetByID)

			// Protected routes
			protected := users.Group("")
			protected.Use(middleware.Auth())
			{
				protected.PUT("/:id", userHandler.Update)
				protected.DELETE("/:id", userHandler.Delete)
			}
		}
	}

	return r
}
