package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tanawat011/golang-boilerplate/internal/middleware"
	"github.com/tanawat011/golang-boilerplate/internal/service"
	"github.com/tanawat011/golang-boilerplate/pkg/pagination"
)

// UserHandler handles the /users endpoints. Every route requires Auth.
type UserHandler struct {
	users service.UserService
}

// NewUserHandler creates a UserHandler.
func NewUserHandler(users service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

// List handles GET /api/v1/users (admin only).
func (h *UserHandler) List(c *gin.Context) {
	actor, _ := middleware.ActorFrom(c)
	page, limit := pagination.GetPaginationParams(c)

	users, total, err := h.users.List(c.Request.Context(), actor, page, limit)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	PaginatedResponse(c, users, page, limit, total)
}

// GetByID handles GET /api/v1/users/:id.
func (h *UserHandler) GetByID(c *gin.Context) {
	id, ok := userIDParam(c)
	if !ok {
		return
	}
	actor, _ := middleware.ActorFrom(c)

	user, err := h.users.GetByID(c.Request.Context(), actor, id)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	SuccessResponse(c, http.StatusOK, "User retrieved successfully", user)
}

// Update handles PATCH /api/v1/users/:id.
func (h *UserHandler) Update(c *gin.Context) {
	id, ok := userIDParam(c)
	if !ok {
		return
	}
	var req service.UpdateUserRequest
	if !bindJSON(c, &req) {
		return
	}
	actor, _ := middleware.ActorFrom(c)

	user, err := h.users.Update(c.Request.Context(), actor, id, req)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	SuccessResponse(c, http.StatusOK, "User updated successfully", user)
}

// Delete handles DELETE /api/v1/users/:id.
func (h *UserHandler) Delete(c *gin.Context) {
	id, ok := userIDParam(c)
	if !ok {
		return
	}
	actor, _ := middleware.ActorFrom(c)

	if err := h.users.Delete(c.Request.Context(), actor, id); err != nil {
		writeServiceError(c, err)
		return
	}
	SuccessResponse(c, http.StatusOK, "User deleted successfully", nil)
}
