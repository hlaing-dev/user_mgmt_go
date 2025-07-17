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

// AdminHandler handles admin-specific requests
type AdminHandler struct {
	userRepo    repository.UserRepository
	logRepo     repository.UserLogRepository
	repoManager *repository.RepositoryManager
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	userRepo repository.UserRepository,
	logRepo repository.UserLogRepository,
	repoManager *repository.RepositoryManager,
) *AdminHandler {
	return &AdminHandler{
		userRepo:    userRepo,
		logRepo:     logRepo,
		repoManager: repoManager,
	}
}

// GetSystemStats godoc
// @Summary Get system statistics
// @Description Get comprehensive system statistics and metrics
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/stats [get]
func (h *AdminHandler) GetSystemStats(c *gin.Context) {
	// Get system statistics
	stats, err := h.repoManager.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"Stats Retrieval Failed",
			"Failed to retrieve system statistics",
			err.Error(),
		))
		return
	}

	// Get health status
	health := h.repoManager.HealthCheck()
	stats["database_health"] = health

	// Add admin-specific stats
	deletedUsers, err := h.userRepo.GetAllDeleted(c.Request.Context(), repository.ListParams{PageSize: 1})
	if err == nil {
		stats["deleted_users_count"] = deletedUsers.Total
	}

	c.JSON(http.StatusOK, stats)
}

// GetUserLogs godoc
// @Summary Get user activity logs
// @Description Get paginated user activity logs with filtering options
// @Tags admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param user_id query string false "Filter by user ID"
// @Param event query string false "Filter by event type"
// @Param start_date query string false "Start date (RFC3339)"
// @Param end_date query string false "End date (RFC3339)"
// @Param ip_address query string false "Filter by IP address"
// @Param action query string false "Filter by action"
// @Success 200 {object} models.UserLogsListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/logs [get]
func (h *AdminHandler) GetUserLogs(c *gin.Context) {
	// Parse filter parameters
	filter := models.LogFilterRequest{
		Page:     1,
		PageSize: 10,
	}

	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && page > 0 {
		filter.Page = page
	}

	if pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10")); err == nil && pageSize > 0 && pageSize <= 100 {
		filter.PageSize = pageSize
	}

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			filter.UserID = &userID
		}
	}

	if event := c.Query("event"); event != "" {
		eventType := models.LogEventType(event)
		if models.IsValidEventType(eventType) {
			filter.Event = &eventType
		}
	}

	if ipAddress := c.Query("ip_address"); ipAddress != "" {
		filter.IPAddress = ipAddress
	}

	if action := c.Query("action"); action != "" {
		filter.Action = action
	}

	// Note: For date filtering, you would parse start_date and end_date
	// from query parameters and convert them to time.Time

	// Get logs
	logs, err := h.logRepo.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"Logs Retrieval Failed",
			"Failed to retrieve user logs",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, logs)
}

// GetDeletedUsers godoc
// @Summary Get deleted users
// @Description Get list of soft-deleted users for potential restoration
// @Tags admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} models.UsersListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/users/deleted [get]
func (h *AdminHandler) GetDeletedUsers(c *gin.Context) {
	// Parse pagination parameters
	params := repository.ListParams{
		Page:     1,
		PageSize: 10,
		SortBy:   "deleted_at",
		SortDir:  "desc",
	}

	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && page > 0 {
		params.Page = page
	}

	if pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10")); err == nil && pageSize > 0 && pageSize <= 100 {
		params.PageSize = pageSize
	}

	// Get deleted users
	deletedUsers, err := h.userRepo.GetAllDeleted(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"Retrieval Failed",
			"Failed to retrieve deleted users",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, deletedUsers)
}

// RestoreUser godoc
// @Summary Restore deleted user
// @Description Restore a soft-deleted user account
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/users/{id}/restore [post]
func (h *AdminHandler) RestoreUser(c *gin.Context) {
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

	// Restore user
	if err := h.userRepo.RestoreDeleted(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			http.StatusNotFound,
			"Restore Failed",
			"Failed to restore user - user may not be deleted or may not exist",
			err.Error(),
		))
		return
	}

	// Log user restoration
	h.logUserRestoration(c, userID)

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		"User restored successfully",
		map[string]interface{}{
			"restored_user_id": userID,
		},
	))
}

// PermanentDeleteUser godoc
// @Summary Permanently delete user
// @Description Permanently delete a user from the database (irreversible)
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/users/{id}/permanent-delete [delete]
func (h *AdminHandler) PermanentDeleteUser(c *gin.Context) {
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

	// Check if admin is trying to permanently delete themselves
	if userClaims, exists := middleware.GetUserFromContext(c); exists {
		if userClaims.UserID == userID {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				http.StatusBadRequest,
				"Cannot Delete Self",
				"You cannot permanently delete your own account",
				nil,
			))
			return
		}
	}

	// Permanently delete user
	if err := h.userRepo.PermanentDelete(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			http.StatusNotFound,
			"Deletion Failed",
			"Failed to permanently delete user",
			err.Error(),
		))
		return
	}

	// Log permanent deletion
	h.logPermanentDeletion(c, userID)

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		"User permanently deleted",
		map[string]interface{}{
			"deleted_user_id": userID,
			"warning":         "This action is irreversible",
		},
	))
}

// BulkCreateUsers godoc
// @Summary Bulk create users
// @Description Create multiple users in a single operation
// @Tags admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body BulkCreateUsersRequest true "Bulk user creation data"
// @Success 201 {object} BulkCreateUsersResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/users/bulk-create [post]
func (h *AdminHandler) BulkCreateUsers(c *gin.Context) {
	var req BulkCreateUsersRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			http.StatusBadRequest,
			"Invalid Request",
			"Please provide valid bulk user data",
			err.Error(),
		))
		return
	}

	// Validate bulk size limits
	if len(req.Users) == 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			http.StatusBadRequest,
			"Empty Request",
			"No users provided for creation",
			nil,
		))
		return
	}

	if len(req.Users) > 100 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			http.StatusBadRequest,
			"Too Many Users",
			"Maximum 100 users can be created at once",
			nil,
		))
		return
	}

	var users []*models.User
	var results []BulkCreateResult
	var successCount, errorCount int

	// Process each user
	for i, userReq := range req.Users {
		result := BulkCreateResult{
			Index: i,
			Email: userReq.Email,
		}

		// Validate individual user
		if userReq.Name == "" || userReq.Email == "" || userReq.Password == "" {
			result.Success = false
			result.Error = "Missing required fields"
			errorCount++
			results = append(results, result)
			continue
		}

		// Check if user already exists
		exists, err := h.userRepo.Exists(c.Request.Context(), userReq.Email)
		if err != nil || exists {
			result.Success = false
			if err != nil {
				result.Error = "Database error"
			} else {
				result.Error = "User already exists"
			}
			errorCount++
			results = append(results, result)
			continue
		}

		// Hash password
		hashedPassword, err := h.hashPassword(userReq.Password)
		if err != nil {
			result.Success = false
			result.Error = "Password hashing failed"
			errorCount++
			results = append(results, result)
			continue
		}

		// Create user object
		user := &models.User{
			Name:     userReq.Name,
			Email:    userReq.Email,
			Password: hashedPassword,
		}

		users = append(users, user)
		result.Success = true
		successCount++
		results = append(results, result)
	}

	// Perform bulk creation for valid users
	if len(users) > 0 {
		if err := h.userRepo.CreateBatch(c.Request.Context(), users); err != nil {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				http.StatusInternalServerError,
				"Bulk Creation Failed",
				"Failed to create users in batch",
				err.Error(),
			))
			return
		}

		// Update results with created user IDs
		userIndex := 0
		for i := range results {
			if results[i].Success {
				results[i].UserID = &users[userIndex].ID
				userIndex++
			}
		}

		// Log bulk creation
		h.logBulkCreation(c, users)
	}

	response := BulkCreateUsersResponse{
		TotalProcessed: len(req.Users),
		SuccessCount:   successCount,
		ErrorCount:     errorCount,
		Results:        results,
	}

	status := http.StatusCreated
	if errorCount > 0 {
		status = http.StatusPartialContent
	}

	c.JSON(status, response)
}

// RunMaintenance godoc
// @Summary Run system maintenance
// @Description Run system maintenance tasks (log cleanup, etc.)
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/maintenance [post]
func (h *AdminHandler) RunMaintenance(c *gin.Context) {
	if err := h.repoManager.RunMaintenance(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"Maintenance Failed",
			"Failed to run system maintenance",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		"System maintenance completed successfully",
		nil,
	))
}

// Request/Response types for bulk operations

type BulkCreateUsersRequest struct {
	Users []models.UserCreateRequest `json:"users" binding:"required"`
}

type BulkCreateUsersResponse struct {
	TotalProcessed int                `json:"total_processed"`
	SuccessCount   int                `json:"success_count"`
	ErrorCount     int                `json:"error_count"`
	Results        []BulkCreateResult `json:"results"`
}

type BulkCreateResult struct {
	Index   int        `json:"index"`
	Email   string     `json:"email"`
	UserID  *uuid.UUID `json:"user_id,omitempty"`
	Success bool       `json:"success"`
	Error   string     `json:"error,omitempty"`
}

// Helper methods

func (h *AdminHandler) hashPassword(password string) (string, error) {
	// Import utils package at the top and use the actual hashing function
	return utils.HashPassword(password)
}

func (h *AdminHandler) logUserRestoration(c *gin.Context, userID uuid.UUID) {
	// Get admin from context
	var adminID *uuid.UUID
	if userClaims, exists := middleware.GetUserFromContext(c); exists {
		adminID = &userClaims.UserID
	}

	logEntry := models.NewUserLog(models.UserLogCreateRequest{
		UserID: adminID,
		Event:  models.UserUpdated,
		Action: "RESTORE_USER",
		Details: map[string]interface{}{
			"restored_user_id": userID,
			"ip_address":       c.ClientIP(),
			"user_agent":       c.Request.UserAgent(),
		},
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})

	h.logRepo.CreateAsync(logEntry)
}

func (h *AdminHandler) logPermanentDeletion(c *gin.Context, userID uuid.UUID) {
	// Get admin from context
	var adminID *uuid.UUID
	if userClaims, exists := middleware.GetUserFromContext(c); exists {
		adminID = &userClaims.UserID
	}

	logEntry := models.NewUserLog(models.UserLogCreateRequest{
		UserID: adminID,
		Event:  models.UserDeleted,
		Action: "PERMANENT_DELETE_USER",
		Details: map[string]interface{}{
			"permanently_deleted_user_id": userID,
			"ip_address":                  c.ClientIP(),
			"user_agent":                  c.Request.UserAgent(),
			"warning":                     "IRREVERSIBLE_ACTION",
		},
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})

	h.logRepo.CreateAsync(logEntry)
}

func (h *AdminHandler) logBulkCreation(c *gin.Context, users []*models.User) {
	// Get admin from context
	var adminID *uuid.UUID
	if userClaims, exists := middleware.GetUserFromContext(c); exists {
		adminID = &userClaims.UserID
	}

	userIDs := make([]uuid.UUID, len(users))
	for i, user := range users {
		userIDs[i] = user.ID
	}

	logEntry := models.NewUserLog(models.UserLogCreateRequest{
		UserID: adminID,
		Event:  models.UserCreated,
		Action: "BULK_CREATE_USERS",
		Details: map[string]interface{}{
			"created_user_count": len(users),
			"created_user_ids":   userIDs,
			"ip_address":         c.ClientIP(),
			"user_agent":         c.Request.UserAgent(),
		},
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})

	h.logRepo.CreateAsync(logEntry)
} 