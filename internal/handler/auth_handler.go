package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tanawat011/golang-boilerplate/internal/middleware"
	"github.com/tanawat011/golang-boilerplate/internal/service"
)

// AuthHandler handles registration, login and the current user.
type AuthHandler struct {
	auth  service.AuthService
	users service.UserService
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(auth service.AuthService, users service.UserService) *AuthHandler {
	return &AuthHandler{auth: auth, users: users}
}

// Register handles POST /api/v1/auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if !bindJSON(c, &req) {
		return
	}
	user, err := h.auth.Register(c.Request.Context(), req)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	SuccessResponse(c, http.StatusCreated, "User registered successfully", user)
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.auth.Login(c.Request.Context(), req)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	SuccessResponse(c, http.StatusOK, "Logged in successfully", result)
}

// Me handles GET /api/v1/auth/me.
func (h *AuthHandler) Me(c *gin.Context) {
	actor, _ := middleware.ActorFrom(c)
	user, err := h.users.GetByID(c.Request.Context(), actor, actor.UserID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	SuccessResponse(c, http.StatusOK, "Current user", user)
}
