package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tanawat011/golang-boilerplate/internal/service"
	"github.com/tanawat011/golang-boilerplate/pkg/validator"
)

// writeServiceError maps a service error to its HTTP status. Unknown errors
// become a 500 without details.
func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
	case errors.Is(err, service.ErrEmailAlreadyExists),
		errors.Is(err, service.ErrUsernameExists),
		errors.Is(err, service.ErrAlreadyExists):
		ErrorResponse(c, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, service.ErrInvalidCredentials):
		ErrorResponse(c, http.StatusUnauthorized, err.Error(), nil)
	case errors.Is(err, service.ErrAccountDisabled), errors.Is(err, service.ErrForbidden):
		ErrorResponse(c, http.StatusForbidden, err.Error(), nil)
	case errors.Is(err, service.ErrPasswordTooLong):
		ValidationErrorResponse(c, []validator.ValidationError{{Field: "password", Message: err.Error()}})
	default:
		internalError(c, err)
	}
}

// bindJSON decodes and validates the body into dst. It writes the error
// response itself and returns false when the request is invalid.
func bindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid JSON body", nil)
		return false
	}
	if errs := validator.ValidateStruct(dst); errs != nil {
		ValidationErrorResponse(c, errs)
		return false
	}
	return true
}

// userIDParam returns the :id path parameter when it is a UUID, and answers
// 404 otherwise (an ID that cannot exist is simply not found).
func userIDParam(c *gin.Context) (string, bool) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		ErrorResponse(c, http.StatusNotFound, service.ErrUserNotFound.Error(), nil)
		return "", false
	}
	return id, true
}
