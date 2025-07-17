package handlers

import (
	"net/http"

	"user_mgmt_go/internal/middleware"
	"user_mgmt_go/internal/models"
	"user_mgmt_go/internal/repository"
	"user_mgmt_go/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	jwtManager  *utils.JWTManager
	userRepo    repository.UserRepository
	logRepo     repository.UserLogRepository
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(
	jwtManager *utils.JWTManager,
	userRepo repository.UserRepository,
	logRepo repository.UserLogRepository,
) *AuthHandler {
	return &AuthHandler{
		jwtManager: jwtManager,
		userRepo:   userRepo,
		logRepo:    logRepo,
	}
}

// Login godoc
// @Summary Admin login
// @Description Authenticate admin user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			http.StatusBadRequest,
			"Invalid Request",
			"Please provide valid email and password",
			err.Error(),
		))
		return
	}

	// Get user by email
	user, err := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		// Log failed login attempt
		h.logFailedLogin(c, req.Email, "User not found")
		
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			http.StatusUnauthorized,
			"Invalid Credentials",
			"Invalid email or password",
			nil,
		))
		return
	}

	// Verify password
	if err := utils.VerifyPassword(user.Password, req.Password); err != nil {
		// Log failed login attempt
		h.logFailedLogin(c, req.Email, "Invalid password")
		
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			http.StatusUnauthorized,
			"Invalid Credentials", 
			"Invalid email or password",
			nil,
		))
		return
	}

	// Determine user role (for this assignment, we'll use admin for the configured admin email)
	role := "user"
	if req.Email == "admin@example.com" { // This should come from config
		role = "admin"
	}

	// Generate JWT tokens
	tokenPair, err := h.jwtManager.GenerateTokenPair(user, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"Token Generation Failed",
			"Failed to generate authentication tokens",
			err.Error(),
		))
		return
	}

	// Log successful login
	h.logSuccessfulLogin(c, user)

	// Return login response
	response := models.LoginResponse{
		Token:        tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		User:         user.ToResponse(),
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Generate new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} models.RefreshTokenResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			http.StatusBadRequest,
			"Invalid Request",
			"Please provide a valid refresh token",
			err.Error(),
		))
		return
	}

	// Generate new access token using refresh token
	response, err := h.jwtManager.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			http.StatusUnauthorized,
			"Invalid Refresh Token",
			"The provided refresh token is invalid or expired",
			err.Error(),
		))
		return
	}

	// Log token refresh
	h.logTokenRefresh(c)

	c.JSON(http.StatusOK, response)
}

// Logout godoc
// @Summary User logout
// @Description Logout user (client-side token invalidation)
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get user from context
	userClaims, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			http.StatusUnauthorized,
			"Unauthorized",
			"Authentication required",
			nil,
		))
		return
	}

	// Log logout
	h.logUserLogout(c, userClaims)

	// Note: In a production system, you would typically:
	// 1. Add token to a blacklist/revocation list
	// 2. Store revoked tokens in Redis with expiry
	// 3. Check blacklist in JWT middleware
	// For this assignment, we'll rely on client-side token removal

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		"Logout successful",
		map[string]string{
			"message": "Please remove the token from client storage",
		},
	))
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get authenticated user's profile information
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.UserResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Get user from context
	userClaims, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			http.StatusUnauthorized,
			"Unauthorized",
			"Authentication required",
			nil,
		))
		return
	}

	// Get full user data from database
	user, err := h.userRepo.GetByID(c.Request.Context(), userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			http.StatusNotFound,
			"User Not Found",
			"User profile not found",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// ChangePassword godoc
// @Summary Change user password
// @Description Change authenticated user's password
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "Password change request"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			http.StatusBadRequest,
			"Invalid Request",
			"Please provide valid password data",
			err.Error(),
		))
		return
	}

	// Get user from context
	userClaims, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			http.StatusUnauthorized,
			"Unauthorized",
			"Authentication required",
			nil,
		))
		return
	}

	// Get user from database
	user, err := h.userRepo.GetByID(c.Request.Context(), userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			http.StatusNotFound,
			"User Not Found",
			"User not found",
			err.Error(),
		))
		return
	}

	// Verify current password
	if err := utils.VerifyPassword(user.Password, req.CurrentPassword); err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			http.StatusUnauthorized,
			"Invalid Password",
			"Current password is incorrect",
			nil,
		))
		return
	}

	// Validate new password
	if !utils.IsValidPassword(req.NewPassword) {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			http.StatusBadRequest,
			"Invalid Password",
			"New password does not meet requirements",
			nil,
		))
		return
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"Password Hashing Failed",
			"Failed to process new password",
			err.Error(),
		))
		return
	}

	// Update password in database
	updates := map[string]interface{}{
		"password": hashedPassword,
	}
	
	if err := h.userRepo.Update(c.Request.Context(), user.ID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"Update Failed",
			"Failed to update password",
			err.Error(),
		))
		return
	}

	// Log password change
	h.logPasswordChange(c, user)

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		"Password changed successfully",
		nil,
	))
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required" example:"oldpassword123"`
	NewPassword     string `json:"new_password" binding:"required,min=6" example:"newpassword123"`
}

// Helper methods for logging

func (h *AuthHandler) logFailedLogin(c *gin.Context, email, reason string) {
	logEntry := models.NewUserLog(models.UserLogCreateRequest{
		Event:  models.LoginFailed,
		Action: "LOGIN_FAILED",
		Details: map[string]interface{}{
			"email":      email,
			"reason":     reason,
			"ip_address": c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		},
		Error:     reason,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})

	h.logRepo.CreateAsync(logEntry)
}

func (h *AuthHandler) logSuccessfulLogin(c *gin.Context, user *models.User) {
	logEntry := models.NewUserLog(models.UserLogCreateRequest{
		UserID: &user.ID,
		Event:  models.LoginSuccess,
		Action: "LOGIN_SUCCESS",
		Details: map[string]interface{}{
			"email":      user.Email,
			"name":       user.Name,
			"ip_address": c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		},
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})

	h.logRepo.CreateAsync(logEntry)
}

func (h *AuthHandler) logTokenRefresh(c *gin.Context) {
	logEntry := models.NewUserLog(models.UserLogCreateRequest{
		Event:  models.TokenRefresh,
		Action: "TOKEN_REFRESH",
		Details: map[string]interface{}{
			"ip_address": c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		},
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})

	h.logRepo.CreateAsync(logEntry)
}

func (h *AuthHandler) logUserLogout(c *gin.Context, user *models.JWTClaims) {
	logEntry := models.NewUserLog(models.UserLogCreateRequest{
		UserID: &user.UserID,
		Event:  models.AdminLogout,
		Action: "USER_LOGOUT",
		Details: map[string]interface{}{
			"email":      user.Email,
			"name":       user.Name,
			"ip_address": c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		},
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})

	h.logRepo.CreateAsync(logEntry)
}

func (h *AuthHandler) logPasswordChange(c *gin.Context, user *models.User) {
	logEntry := models.NewUserLog(models.UserLogCreateRequest{
		UserID: &user.ID,
		Event:  models.UserUpdated,
		Action: "PASSWORD_CHANGE",
		Details: map[string]interface{}{
			"email":      user.Email,
			"name":       user.Name,
			"ip_address": c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		},
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})

	h.logRepo.CreateAsync(logEntry)
} 