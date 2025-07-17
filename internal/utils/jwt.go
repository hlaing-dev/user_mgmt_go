package utils

import (
	"errors"
	"fmt"
	"time"

	"user_mgmt_go/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTManager handles JWT token operations
type JWTManager struct {
	secretKey    string
	tokenExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewJWTManager creates a new JWT manager instance
func NewJWTManager(secretKey string, tokenExpiry time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		tokenExpiry:   tokenExpiry,
		refreshExpiry: tokenExpiry * 7, // Refresh token lasts 7x longer than access token
	}
}

// GenerateTokenPair generates both access and refresh tokens for a user
func (j *JWTManager) GenerateTokenPair(user *models.User, role string) (*models.TokenPair, error) {
	// Generate access token
	accessToken, expiresAt, err := j.generateToken(user, role, j.tokenExpiry, models.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, _, err := j.generateToken(user, role, j.refreshExpiry, models.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// generateToken generates a JWT token with specified duration and type
func (j *JWTManager) generateToken(user *models.User, role string, duration time.Duration, tokenType models.TokenType) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(duration)

	// Create JWT claims
	claims := &models.JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),          // Unique token ID for revocation
			Subject:   user.ID.String(),             // User ID
			Audience:  jwt.ClaimStrings{string(tokenType)}, // Token type
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "user_mgmt_go",
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTManager) ValidateToken(tokenString string) (*models.JWTClaims, error) {
	// Parse token with claims
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, fmt.Errorf("malformed token")
		} else if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token has expired")
		} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, fmt.Errorf("token not valid yet")
		}
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(*models.JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Additional validation
	if !claims.IsValid() {
		return nil, fmt.Errorf("token claims validation failed")
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token using a valid refresh token
func (j *JWTManager) RefreshAccessToken(refreshTokenString string) (*models.RefreshTokenResponse, error) {
	// Validate refresh token
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if token is a refresh token
	if len(claims.Audience) == 0 || claims.Audience[0] != string(models.RefreshToken) {
		return nil, fmt.Errorf("provided token is not a refresh token")
	}

	// Create user object from claims
	user := &models.User{
		ID:    claims.UserID,
		Email: claims.Email,
		Name:  claims.Name,
	}

	// Generate new access token
	accessToken, expiresAt, err := j.generateToken(user, claims.Role, j.tokenExpiry, models.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new access token: %w", err)
	}

	return &models.RefreshTokenResponse{
		Token:     accessToken,
		ExpiresAt: expiresAt,
	}, nil
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is required")
	}

	// Check for Bearer prefix
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", fmt.Errorf("authorization header must start with 'Bearer '")
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		return "", fmt.Errorf("token is required")
	}

	return token, nil
}

// GetTokenExpiryTime returns the expiry time for access tokens
func (j *JWTManager) GetTokenExpiryTime() time.Duration {
	return j.tokenExpiry
}

// GetRefreshExpiryTime returns the expiry time for refresh tokens
func (j *JWTManager) GetRefreshExpiryTime() time.Duration {
	return j.refreshExpiry
}

// GenerateTokenForUser generates a token pair for a user with specified role
func (j *JWTManager) GenerateTokenForUser(userID uuid.UUID, email, name, role string) (*models.TokenPair, error) {
	user := &models.User{
		ID:    userID,
		Email: email,
		Name:  name,
	}
	
	return j.GenerateTokenPair(user, role)
}

// IsTokenExpired checks if a token is expired without full validation
func IsTokenExpired(tokenString string) bool {
	// Parse without validation to check expiry
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &models.JWTClaims{})
	if err != nil {
		return true
	}

	claims, ok := token.Claims.(*models.JWTClaims)
	if !ok {
		return true
	}

	return time.Now().Unix() > claims.ExpiresAt.Unix()
}

// GetTokenClaims extracts claims from a token without validation (for debugging)
func GetTokenClaims(tokenString string) (*models.JWTClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &models.JWTClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*models.JWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	return claims, nil
} 