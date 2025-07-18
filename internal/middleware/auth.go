package middleware

import (
	"net/http"

	"user_mgmt_go/internal/models"
	"user_mgmt_go/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(jwtManager *utils.JWTManager) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		var token string
		var err error

		// Extract token from Authorization header first
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// Extract token from "Bearer <token>" format
			token, err = utils.ExtractTokenFromHeader(authHeader)
			if err != nil {
				c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
					http.StatusUnauthorized,
					"Unauthorized",
					err.Error(),
					nil,
				))
				c.Abort()
				return
			}
		} else {
			// If no Authorization header, check for admin_token cookie (for admin panel)
			token, err = c.Cookie("admin_token")
			if err != nil || token == "" {
				c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
					http.StatusUnauthorized,
					"Unauthorized",
					"Authorization header is required",
					nil,
				))
				c.Abort()
				return
			}
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
				http.StatusUnauthorized,
				"Unauthorized",
				"Invalid or expired token",
				map[string]string{"error": err.Error()},
			))
			c.Abort()
			return
		}

		// Store user information in context for use in handlers
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_name", claims.Name)
		c.Set("user_role", claims.Role)
		c.Set("jwt_claims", claims)

		// Continue to next handler
		c.Next()
	})
}

// OptionalAuthMiddleware creates optional authentication middleware
// If token is present, it validates and sets user context
// If token is missing, it continues without authentication
func OptionalAuthMiddleware(jwtManager *utils.JWTManager) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Extract and validate token
		token, err := utils.ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.Next()
			return
		}

		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			c.Next()
			return
		}

		// Store user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_name", claims.Name)
		c.Set("user_role", claims.Role)
		c.Set("jwt_claims", claims)
		c.Set("authenticated", true)

		c.Next()
	})
}

// AdminRequiredMiddleware ensures the user has admin role
func AdminRequiredMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
				http.StatusUnauthorized,
				"Unauthorized",
				"Authentication required",
				nil,
			))
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok || role != "admin" {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(
				http.StatusForbidden,
				"Forbidden",
				"Admin access required",
				nil,
			))
			c.Abort()
			return
		}

		c.Next()
	})
}

// RoleRequiredMiddleware checks if user has any of the required roles
func RoleRequiredMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
				http.StatusUnauthorized,
				"Unauthorized",
				"Authentication required",
				nil,
			))
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
				http.StatusUnauthorized,
				"Unauthorized",
				"Invalid user role",
				nil,
			))
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			if role == requiredRole {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(
				http.StatusForbidden,
				"Forbidden",
				"Insufficient privileges",
				map[string]interface{}{
					"required_roles": requiredRoles,
					"user_role":      role,
				},
			))
			c.Abort()
			return
		}

		c.Next()
	})
}

// SelfOrAdminMiddleware allows access if user is accessing their own resource or is admin
func SelfOrAdminMiddleware(userIDParam string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
				http.StatusUnauthorized,
				"Unauthorized",
				"Authentication required",
				nil,
			))
			c.Abort()
			return
		}

		userRole, roleExists := c.Get("user_role")
		if !roleExists {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
				http.StatusUnauthorized,
				"Unauthorized",
				"Invalid authentication",
				nil,
			))
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
				http.StatusUnauthorized,
				"Unauthorized",
				"Invalid user role",
				nil,
			))
			c.Abort()
			return
		}

		// Admin can access any resource
		if role == "admin" {
			c.Next()
			return
		}

		// Check if user is accessing their own resource
		requestedUserID := c.Param(userIDParam)
		if requestedUserID == "" {
			requestedUserID = c.Query(userIDParam)
		}

		if requestedUserID == userID.(string) || requestedUserID == userID.(*models.User).ID.String() {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			http.StatusForbidden,
			"Forbidden",
			"Access denied: you can only access your own resources",
			nil,
		))
		c.Abort()
	})
}

// GetUserFromContext helper function to extract user information from context
func GetUserFromContext(c *gin.Context) (*models.JWTClaims, bool) {
	claims, exists := c.Get("jwt_claims")
	if !exists {
		return nil, false
	}

	userClaims, ok := claims.(*models.JWTClaims)
	return userClaims, ok
}

// IsAuthenticated checks if the request is authenticated
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("user_id")
	return exists
}

// IsAdmin checks if the authenticated user is an admin
func IsAdmin(c *gin.Context) bool {
	userRole, exists := c.Get("user_role")
	if !exists {
		return false
	}

	role, ok := userRole.(string)
	return ok && role == "admin"
}

// HasRole checks if the authenticated user has a specific role
func HasRole(c *gin.Context, role string) bool {
	userRole, exists := c.Get("user_role")
	if !exists {
		return false
	}

	currentRole, ok := userRole.(string)
	return ok && currentRole == role
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) (interface{}, bool) {
	return c.Get("user_id")
}

// GetUserEmail extracts user email from context
func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("user_email")
	if !exists {
		return "", false
	}

	userEmail, ok := email.(string)
	return userEmail, ok
}

// GetUserRole extracts user role from context
func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("user_role")
	if !exists {
		return "", false
	}

	userRole, ok := role.(string)
	return userRole, ok
} 