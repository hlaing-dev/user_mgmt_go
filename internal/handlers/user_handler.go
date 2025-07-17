package handlers

import (
	"net/http"
	"strconv"

	"user_mgmt_go/internal/middleware"
	"user_mgmt_go/internal/models"
	"user_mgmt_go/internal/repository"
	"user_mgmt_go/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler handles user-related requests
type UserHandler struct {
	userRepo repository.UserRepository
	logRepo  repository.UserLogRepository
}

// NewUserHandler creates a new user handler
func NewUserHandler(
	userRepo repository.UserRepository,
	logRepo repository.UserLogRepository,
) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
		logRepo:  logRepo,
	}
}

// ListUsers godoc
// @Summary List users
// @Description Get paginated list of users with optional filtering and search
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param sort_by query string false "Sort by field" default("created_at")
// @Param sort_dir query string false "Sort direction" default("desc") Enums(asc, desc)
// @Param search query string false "Search term"
// @Success 200 {object} models.UsersListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse pagination parameters
	params := repository.ListParams{
		Page:     1,
		PageSize: 10,
		SortBy:   "created_at",
		SortDir:  "desc",
	}

	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && page > 0 {
		params.Page = page
	}

	if pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10")); err == nil && pageSize > 0 && pageSize <= 100 {
		params.PageSize = pageSize
	}

	if sortBy := c.Query("sort_by"); sortBy != "" && repository.IsValidUserSortField(sortBy) {
		params.SortBy = sortBy
	}

	if sortDir := c.Query("sort_dir"); sortDir == "asc" || sortDir == "desc" {
		params.SortDir = sortDir
	}

	// Check for search parameter
	searchTerm := c.Query("search")
	
	var response *models.UsersListResponse
	var err error

	if searchTerm != "" {
		// Perform search
		response, err = h.userRepo.Search(c.Request.Context(), searchTerm, params)
	} else {
		// Regular list
		response, err = h.userRepo.List(c.Request.Context(), params)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"List Failed",
			"Failed to retrieve users",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get detailed information about a specific user
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	// Parse user ID
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			http.StatusBadRequest,
			"Invalid User ID",
			"Please provide a valid user ID",
			err.Error(),
		))
		return
	}

	// Get user from database
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			http.StatusNotFound,
			"User Not Found",
			"User with the specified ID was not found",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// CreateUser godoc
// @Summary Create new user
// @Description Create a new user account
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.UserCreateRequest true "User creation data"
// @Success 201 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.UserCreateRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			http.StatusBadRequest,
			"Invalid Request",
			"Please provide valid user data",
			err.Error(),
		))
		return
	}

	// Validate password strength
	if !utils.IsValidPassword(req.Password) {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			http.StatusBadRequest,
			"Invalid Password",
			"Password does not meet requirements (minimum 6 characters)",
			nil,
		))
		return
	}

	// Check if user already exists
	exists, err := h.userRepo.Exists(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"Database Error",
			"Failed to check user existence",
			err.Error(),
		))
		return
	}

	if exists {
		c.JSON(http.StatusConflict, models.NewErrorResponse(
			http.StatusConflict,
			"User Already Exists",
			"A user with this email already exists",
			nil,
		))
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"Password Hashing Failed",
			"Failed to process password",
			err.Error(),
		))
		return
	}

	// Create user object
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	// Save to database
	if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"Creation Failed",
			"Failed to create user",
			err.Error(),
		))
		return
	}

	// Log user creation
	h.logUserCreation(c, user)

	c.JSON(http.StatusCreated, user.ToResponse())
}

// UpdateUser godoc
// @Summary Update user
// @Description Update user information
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body models.UserUpdateRequest true "User update data"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// Parse user ID
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			http.StatusBadRequest,
			"Invalid User ID",
			"Please provide a valid user ID",
			err.Error(),
		))
		return
	}

	var req models.UserUpdateRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			http.StatusBadRequest,
			"Invalid Request",
			"Please provide valid update data",
			err.Error(),
		))
		return
	}

	// Get existing user for logging old values
	existingUser, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			http.StatusNotFound,
			"User Not Found",
			"User with the specified ID was not found",
			err.Error(),
		))
		return
	}

	// Build update map
	updates := make(map[string]interface{})
	oldValues := make(map[string]interface{})
	newValues := make(map[string]interface{})

	if req.Name != nil && *req.Name != existingUser.Name {
		updates["name"] = *req.Name
		oldValues["name"] = existingUser.Name
		newValues["name"] = *req.Name
	}

	if req.Email != nil && *req.Email != existingUser.Email {
		// Check if new email already exists
		exists, err := h.userRepo.Exists(c.Request.Context(), *req.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				http.StatusInternalServerError,
				"Database Error",
				"Failed to check email existence",
				err.Error(),
			))
			return
		}

		if exists {
			c.JSON(http.StatusConflict, models.NewErrorResponse(
				http.StatusConflict,
				"Email Already Exists",
				"Another user with this email already exists",
				nil,
			))
			return
		}

		updates["email"] = *req.Email
		oldValues["email"] = existingUser.Email
		newValues["email"] = *req.Email
	}

	if req.Password != nil {
		// Validate password strength
		if !utils.IsValidPassword(*req.Password) {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				http.StatusBadRequest,
				"Invalid Password",
				"Password does not meet requirements (minimum 6 characters)",
				nil,
			))
			return
		}

		// Hash new password
		hashedPassword, err := utils.HashPassword(*req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				http.StatusInternalServerError,
				"Password Hashing Failed",
				"Failed to process password",
				err.Error(),
			))
			return
		}

		updates["password"] = hashedPassword
		oldValues["password"] = "[REDACTED]"
		newValues["password"] = "[REDACTED]"
	}

	// Check if there are any updates
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			http.StatusBadRequest,
			"No Updates",
			"No valid updates provided",
			nil,
		))
		return
	}

	// Perform update
	if err := h.userRepo.Update(c.Request.Context(), userID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"Update Failed",
			"Failed to update user",
			err.Error(),
		))
		return
	}

	// Get updated user
	updatedUser, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"Retrieval Failed",
			"Failed to retrieve updated user",
			err.Error(),
		))
		return
	}

	// Log user update
	h.logUserUpdate(c, updatedUser, oldValues, newValues)

	c.JSON(http.StatusOK, updatedUser.ToResponse())
}

// DeleteUser godoc
// @Summary Delete user
// @Description Soft delete a user account
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// Parse user ID
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			http.StatusBadRequest,
			"Invalid User ID",
			"Please provide a valid user ID",
			err.Error(),
		))
		return
	}

	// Get user before deletion for logging
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			http.StatusNotFound,
			"User Not Found",
			"User with the specified ID was not found",
			err.Error(),
		))
		return
	}

	// Check if user is trying to delete themselves
	if userClaims, exists := middleware.GetUserFromContext(c); exists {
		if userClaims.UserID == userID {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				http.StatusBadRequest,
				"Cannot Delete Self",
				"You cannot delete your own account",
				nil,
			))
			return
		}
	}

	// Perform soft delete
	if err := h.userRepo.Delete(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"Deletion Failed",
			"Failed to delete user",
			err.Error(),
		))
		return
	}

	// Log user deletion
	h.logUserDeletion(c, user)

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		"User deleted successfully",
		map[string]interface{}{
			"deleted_user_id": userID,
			"deleted_email":   user.Email,
		},
	))
}

// Helper methods for logging

func (h *UserHandler) logUserCreation(c *gin.Context, user *models.User) {
	// Get creator from context
	var creatorID *uuid.UUID
	if userClaims, exists := middleware.GetUserFromContext(c); exists {
		creatorID = &userClaims.UserID
	}

	logEntry := models.NewUserLog(models.UserLogCreateRequest{
		UserID: creatorID,
		Event:  models.UserCreated,
		Action: "CREATE_USER",
		Details: map[string]interface{}{
			"created_user_id":    user.ID,
			"created_user_email": user.Email,
			"created_user_name":  user.Name,
			"ip_address":         c.ClientIP(),
			"user_agent":         c.Request.UserAgent(),
		},
		NewValues: map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})

	h.logRepo.CreateAsync(logEntry)
}

func (h *UserHandler) logUserUpdate(c *gin.Context, user *models.User, oldValues, newValues map[string]interface{}) {
	// Get updater from context
	var updaterID *uuid.UUID
	if userClaims, exists := middleware.GetUserFromContext(c); exists {
		updaterID = &userClaims.UserID
	}

	logEntry := models.NewUserLog(models.UserLogCreateRequest{
		UserID: updaterID,
		Event:  models.UserUpdated,
		Action: "UPDATE_USER",
		Details: map[string]interface{}{
			"updated_user_id":    user.ID,
			"updated_user_email": user.Email,
			"updated_user_name":  user.Name,
			"ip_address":         c.ClientIP(),
			"user_agent":         c.Request.UserAgent(),
		},
		OldValues: oldValues,
		NewValues: newValues,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})

	h.logRepo.CreateAsync(logEntry)
}

func (h *UserHandler) logUserDeletion(c *gin.Context, user *models.User) {
	// Get deleter from context
	var deleterID *uuid.UUID
	if userClaims, exists := middleware.GetUserFromContext(c); exists {
		deleterID = &userClaims.UserID
	}

	logEntry := models.NewUserLog(models.UserLogCreateRequest{
		UserID: deleterID,
		Event:  models.UserDeleted,
		Action: "DELETE_USER",
		Details: map[string]interface{}{
			"deleted_user_id":    user.ID,
			"deleted_user_email": user.Email,
			"deleted_user_name":  user.Name,
			"ip_address":         c.ClientIP(),
			"user_agent":         c.Request.UserAgent(),
		},
		OldValues: map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})

	h.logRepo.CreateAsync(logEntry)
} 