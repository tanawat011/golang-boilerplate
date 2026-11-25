// Package service holds the business logic. It knows nothing about HTTP.
package service

import "errors"

// Errors returned by services. Handlers map them to HTTP status codes.
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUsernameExists     = errors.New("username already exists")
	ErrAlreadyExists      = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountDisabled    = errors.New("account is disabled")
	ErrForbidden          = errors.New("not allowed")
	ErrPasswordTooLong    = errors.New("password must be at most 72 bytes")
)

// Actor is the authenticated caller of a service method.
type Actor struct {
	UserID string
	Role   string
}

// IsAdmin reports whether the actor has the admin role.
func (a Actor) IsAdmin() bool {
	return a.Role == "admin"
}

// canAccess reports whether the actor may read or change the given user.
func (a Actor) canAccess(userID string) bool {
	return a.IsAdmin() || a.UserID == userID
}
