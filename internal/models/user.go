package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents the user entity stored in PostgreSQL
type User struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name      string         `json:"name" gorm:"not null;size:255" binding:"required" example:"John Doe"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null;size:255" binding:"required,email" example:"john.doe@example.com"`
	Password  string         `json:"-" gorm:"not null;size:255" binding:"required,min=6"` // "-" means exclude from JSON
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"` // Soft delete support
}

// UserCreateRequest represents the request payload for creating a user
type UserCreateRequest struct {
	Name     string `json:"name" binding:"required" example:"John Doe"`
	Email    string `json:"email" binding:"required,email" example:"john.doe@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

// UserUpdateRequest represents the request payload for updating a user
type UserUpdateRequest struct {
	Name     *string `json:"name,omitempty" example:"John Doe"`
	Email    *string `json:"email,omitempty" binding:"omitempty,email" example:"john.doe@example.com"`
	Password *string `json:"password,omitempty" binding:"omitempty,min=6" example:"newpassword123"`
}

// UserResponse represents the response payload for user data
type UserResponse struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name      string    `json:"name" example:"John Doe"`
	Email     string    `json:"email" example:"john.doe@example.com"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// UsersListResponse represents the response payload for paginated user list
type UsersListResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int64          `json:"total" example:"100"`
	Page       int            `json:"page" example:"1"`
	PageSize   int            `json:"page_size" example:"10"`
	TotalPages int            `json:"total_pages" example:"10"`
}

// ToResponse converts User model to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// BeforeCreate is a GORM hook that runs before creating a user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
} 