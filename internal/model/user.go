package model

// User represents a user in the system
type User struct {
	Base
	Username string `json:"username" gorm:"uniqueIndex;not null;size:50"`
	Email    string `json:"email" gorm:"uniqueIndex;not null;size:100"`
	Password string `json:"-" gorm:"not null"` // "-" hides password from JSON
	FullName string `json:"full_name" gorm:"size:100"`
	Role     string `json:"role" gorm:"default:user;size:20"`
	IsActive bool   `json:"is_active" gorm:"default:true"`
}

// TableName overrides the table name
func (User) TableName() string {
	return "users"
}

// UserResponse is the response format for user data (without sensitive fields)
type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

// ToResponse converts User model to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		FullName: u.FullName,
		Role:     u.Role,
		IsActive: u.IsActive,
	}
}
