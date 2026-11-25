package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Pinger checks that a dependency is reachable. *sql.DB satisfies it.
type Pinger interface {
	PingContext(ctx context.Context) error
}

// HealthHandler reports whether the app and its database are up.
type HealthHandler struct {
	version string
	db      Pinger
}

// NewHealthHandler creates a HealthHandler.
func NewHealthHandler(version string, db Pinger) *HealthHandler {
	return &HealthHandler{version: version, db: db}
}

// Check handles GET /api/v1/health. It answers 503 when the database does
// not respond within two seconds, so it works as a readiness probe.
func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "unavailable",
			"version":  h.version,
			"database": "down",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"version":  h.version,
		"database": "up",
	})
}
