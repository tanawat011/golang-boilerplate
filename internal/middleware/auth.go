// Package middleware holds the Gin middleware used by the router.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/tanawat011/golang-boilerplate/internal/auth"
	"github.com/tanawat011/golang-boilerplate/internal/service"
)

const actorKey = "actor"

// Auth verifies the "Authorization: Bearer <token>" header and stores the
// caller as a service.Actor in the request context.
func Auth(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		scheme, token, ok := strings.Cut(c.GetHeader("Authorization"), " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
			abort(c, http.StatusUnauthorized, "Authorization header must be 'Bearer <token>'")
			return
		}

		claims, err := tokens.Parse(strings.TrimSpace(token))
		if err != nil {
			abort(c, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		c.Set(actorKey, service.Actor{UserID: claims.Subject, Role: claims.Role})
		c.Next()
	}
}

// RequireRole allows the request only when the authenticated actor has the
// given role. Use it after Auth.
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		actor, ok := ActorFrom(c)
		if !ok || actor.Role != role {
			abort(c, http.StatusForbidden, "Insufficient permissions")
			return
		}
		c.Next()
	}
}

// ActorFrom returns the actor stored by Auth.
func ActorFrom(c *gin.Context) (service.Actor, bool) {
	v, ok := c.Get(actorKey)
	if !ok {
		return service.Actor{}, false
	}
	actor, ok := v.(service.Actor)
	return actor, ok
}

func abort(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"success": false, "message": message})
}
