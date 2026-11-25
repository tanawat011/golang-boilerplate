package model

import "time"

// User roles.
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// User represents a user in the system.
type User struct {
	Base
	Username string `gorm:"uniqueIndex;not null;size:50"`
	Email    string `gorm:"uniqueIndex;not null;size:100"`
	Password string `gorm:"not null"` // bcrypt hash, never serialised
	FullName string `gorm:"size:100"`
	Role     string `gorm:"not null;default:user;size:20"`
	IsActive bool   `gorm:"not null;default:true"`
}

// TableName overrides the table name.
func (User) TableName() string {
	return "users"
}

// UserResponse is the public representation of a user (no password hash).
type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts a User to its public representation.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		FullName:  u.FullName,
		Role:      u.Role,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
