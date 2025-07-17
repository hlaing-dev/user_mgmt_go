package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// LoginRequest represents the request payload for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"admin@example.com"`
	Password string `json:"password" binding:"required" example:"admin123"`
}

// LoginResponse represents the response payload for successful login
type LoginResponse struct {
	Token        string       `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string       `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt    time.Time    `json:"expires_at" example:"2023-12-31T23:59:59Z"`
	User         UserResponse `json:"user"`
}

// RefreshTokenRequest represents the request payload for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// RefreshTokenResponse represents the response payload for token refresh
type RefreshTokenResponse struct {
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt time.Time `json:"expires_at" example:"2023-12-31T23:59:59Z"`
}

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Name   string    `json:"name"`
	Role   string    `json:"role"` // "admin" or "user"
	jwt.RegisteredClaims
}

// TokenType represents different types of tokens
type TokenType string

const (
	AccessToken  TokenType = "access_token"
	RefreshToken TokenType = "refresh_token"
)

// TokenPair represents a pair of access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string      `json:"error" example:"Invalid credentials"`
	Message string      `json:"message" example:"The provided email or password is incorrect"`
	Code    int         `json:"code" example:"401"`
	Details interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a standard success response
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Operation completed successfully"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Page       int   `json:"page" example:"1"`
	PageSize   int   `json:"page_size" example:"10"`
	Total      int64 `json:"total" example:"100"`
	TotalPages int   `json:"total_pages" example:"10"`
}

// HealthCheckResponse represents the health check response
type HealthCheckResponse struct {
	Status    string    `json:"status" example:"healthy"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version" example:"1.0.0"`
	Services  struct {
		Database bool `json:"database" example:"true"`
		MongoDB  bool `json:"mongodb" example:"true"`
	} `json:"services"`
}

// ValidationError represents validation error details
type ValidationError struct {
	Field   string `json:"field" example:"email"`
	Tag     string `json:"tag" example:"required"`
	Value   string `json:"value" example:""`
	Message string `json:"message" example:"Email is required"`
}

// ValidationErrorResponse represents validation error response
type ValidationErrorResponse struct {
	Error   string            `json:"error" example:"Validation failed"`
	Message string            `json:"message" example:"Invalid input data"`
	Code    int               `json:"code" example:"400"`
	Errors  []ValidationError `json:"errors"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(code int, error, message string, details interface{}) *ErrorResponse {
	return &ErrorResponse{
		Error:   error,
		Message: message,
		Code:    code,
		Details: details,
	}
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(message string, data interface{}) *SuccessResponse {
	return &SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewValidationErrorResponse creates a new validation error response
func NewValidationErrorResponse(errors []ValidationError) *ValidationErrorResponse {
	return &ValidationErrorResponse{
		Error:   "Validation failed",
		Message: "Invalid input data",
		Code:    400,
		Errors:  errors,
	}
}

// IsAdmin checks if the user has admin role
func (c *JWTClaims) IsAdmin() bool {
	return c.Role == "admin"
}

// IsValid checks if the token claims are valid
func (c *JWTClaims) IsValid() bool {
	if c.UserID == uuid.Nil || c.Email == "" {
		return false
	}
	
	// Check if token is expired
	if time.Now().Unix() > c.ExpiresAt.Unix() {
		return false
	}
	
	return true
} 