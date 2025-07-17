package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt with default cost
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}
	
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	
	return string(bytes), nil
}

// VerifyPassword compares a hashed password with a plain text password
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// IsValidPassword checks if a password meets minimum requirements
func IsValidPassword(password string) bool {
	// Minimum 6 characters
	if len(password) < 6 {
		return false
	}
	
	// Add more validation rules as needed:
	// - Contains uppercase letter
	// - Contains lowercase letter  
	// - Contains number
	// - Contains special character
	
	return true
} 