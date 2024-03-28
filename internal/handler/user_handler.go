package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tanawat011/golang-boilerplate/internal/service"
	"github.com/tanawat011/golang-boilerplate/pkg/pagination"
	"github.com/tanawat011/golang-boilerplate/pkg/validator"
)

// UserHandler handles HTTP requests related to users
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Create handles POST /api/v1/users
func (h *UserHandler) Create(c *gin.Context) {
	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if errors := validator.ValidateStruct(req); errors != nil {
		ValidationErrorResponse(c, errors)
		return
	}

	user, err := h.userService.Create(req)
	if err != nil {
		switch err {
		case service.ErrEmailAlreadyExists:
			ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		case service.ErrUsernameExists:
			ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		default:
			ErrorResponse(c, http.StatusInternalServerError, "Failed to create user", err.Error())
		}
		return
	}

	SuccessResponse(c, http.StatusCreated, "User created successfully", user)
}

// GetByID handles GET /api/v1/users/:id
func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	user, err := h.userService.GetByID(id)
	if err != nil {
		if err == service.ErrUserNotFound {
			ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user", err.Error())
		return
	}

	SuccessResponse(c, http.StatusOK, "User retrieved successfully", user)
}

// GetAll handles GET /api/v1/users
func (h *UserHandler) GetAll(c *gin.Context) {
	page, limit := pagination.GetPaginationParams(c)

	users, total, err := h.userService.GetAll(page, limit)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve users", err.Error())
		return
	}

	PaginatedResponse(c, users, page, limit, total)
}

// Update handles PUT /api/v1/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if errors := validator.ValidateStruct(req); errors != nil {
		ValidationErrorResponse(c, errors)
		return
	}

	user, err := h.userService.Update(id, req)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
		case service.ErrEmailAlreadyExists, service.ErrUsernameExists:
			ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		default:
			ErrorResponse(c, http.StatusInternalServerError, "Failed to update user", err.Error())
		}
		return
	}

	SuccessResponse(c, http.StatusOK, "User updated successfully", user)
}

// Delete handles DELETE /api/v1/users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.userService.Delete(id); err != nil {
		if err == service.ErrUserNotFound {
			ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to delete user", err.Error())
		return
	}

	SuccessResponse(c, http.StatusOK, "User deleted successfully", nil)
}
