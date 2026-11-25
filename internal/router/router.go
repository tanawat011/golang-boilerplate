// Package router wires middleware and handlers into a Gin engine.
package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/tanawat011/golang-boilerplate/internal/auth"
	"github.com/tanawat011/golang-boilerplate/internal/config"
	"github.com/tanawat011/golang-boilerplate/internal/handler"
	"github.com/tanawat011/golang-boilerplate/internal/middleware"
	"github.com/tanawat011/golang-boilerplate/internal/model"
	"github.com/tanawat011/golang-boilerplate/internal/service"
)

// maxBodyBytes caps every request body (1 MiB).
const maxBodyBytes = 1 << 20

// Deps are the dependencies the router needs.
type Deps struct {
	Config *config.Config
	Logger *zap.Logger
	Tokens *auth.TokenManager
	Auth   service.AuthService
	Users  service.UserService
	DB     handler.Pinger
}

// New builds the Gin engine with every route.
func New(d Deps) (*gin.Engine, error) {
	if d.Config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	// Only proxies listed in TRUSTED_PROXIES may set X-Forwarded-For; the
	// rate limiter keys on the resulting client IP.
	if err := r.SetTrustedProxies(d.Config.App.TrustedProxies); err != nil {
		return nil, err
	}

	r.Use(
		middleware.Recovery(d.Logger),
		middleware.Logger(d.Logger),
		middleware.SecureHeaders(d.Config.IsProduction()),
		middleware.CORS(d.Config.App.CORSOrigins),
		middleware.BodyLimit(maxBodyBytes),
	)

	health := handler.NewHealthHandler(d.Config.App.Version, d.DB)
	authH := handler.NewAuthHandler(d.Auth, d.Users)
	users := handler.NewUserHandler(d.Users)
	authRequired := middleware.Auth(d.Tokens)
	authLimit := middleware.NewRateLimiter(d.Config.Auth.RateLimitPerMinute).Middleware()

	v1 := r.Group("/api/v1")
	v1.GET("/health", health.Check)

	authGroup := v1.Group("/auth")
	authGroup.POST("/register", authLimit, authH.Register)
	authGroup.POST("/login", authLimit, authH.Login)
	authGroup.GET("/me", authRequired, authH.Me)

	userGroup := v1.Group("/users", authRequired)
	userGroup.GET("", middleware.RequireRole(model.RoleAdmin), users.List)
	userGroup.GET("/:id", users.GetByID)
	userGroup.PATCH("/:id", users.Update)
	userGroup.DELETE("/:id", users.Delete)

	r.NoRoute(func(c *gin.Context) {
		handler.ErrorResponse(c, http.StatusNotFound, "Route not found", nil)
	})

	return r, nil
}
