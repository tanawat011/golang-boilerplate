// Package handler is the HTTP layer: it binds requests, calls services and
// writes JSON responses.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the envelope of every JSON response.
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   any    `json:"error,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

// Meta contains pagination metadata.
type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

// SuccessResponse sends a success response.
func SuccessResponse(c *gin.Context, statusCode int, message string, data any) {
	c.JSON(statusCode, Response{Success: true, Message: message, Data: data})
}

// PaginatedResponse sends a page of results with its metadata.
func PaginatedResponse(c *gin.Context, data any, page, limit int, total int64) {
	totalPages := (total + int64(limit) - 1) / int64(limit)
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta:    &Meta{Page: page, Limit: limit, Total: total, TotalPages: totalPages},
	})
}

// ErrorResponse sends an error response. err is shown to the client, so it
// must never carry internal details.
func ErrorResponse(c *gin.Context, statusCode int, message string, err any) {
	c.JSON(statusCode, Response{Success: false, Message: message, Error: err})
}

// ValidationErrorResponse sends a 422 with one entry per invalid field.
func ValidationErrorResponse(c *gin.Context, errors any) {
	ErrorResponse(c, http.StatusUnprocessableEntity, "Validation failed", errors)
}

// internalError logs err for the operator and sends a generic 500.
func internalError(c *gin.Context, err error) {
	_ = c.Error(err)
	ErrorResponse(c, http.StatusInternalServerError, "Internal server error", nil)
}
