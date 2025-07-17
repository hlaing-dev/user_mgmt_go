package handlers

import (
	"net/http"
	"strconv"
	"time"

	"user_mgmt_go/internal/middleware"
	"user_mgmt_go/internal/models"
	"user_mgmt_go/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LogHandler handles log-related requests
type LogHandler struct {
	logRepo repository.UserLogRepository
}

// NewLogHandler creates a new log handler
func NewLogHandler(logRepo repository.UserLogRepository) *LogHandler {
	return &LogHandler{
		logRepo: logRepo,
	}
}

// GetUserLogs godoc
// @Summary Get user activity logs
// @Description Get activity logs for the authenticated user
// @Tags logs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param days query int false "Number of days to look back" default(30)
// @Success 200 {object} models.UserLogsListResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /logs/my-activity [get]
func (h *LogHandler) GetUserLogs(c *gin.Context) {
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

	// Parse pagination parameters
	params := repository.ListParams{
		Page:     1,
		PageSize: 10,
		SortBy:   "timestamp",
		SortDir:  "desc",
	}

	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && page > 0 {
		params.Page = page
	}

	if pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10")); err == nil && pageSize > 0 && pageSize <= 100 {
		params.PageSize = pageSize
	}

	// Get user activity logs
	logs, err := h.logRepo.GetByUserID(c.Request.Context(), userClaims.UserID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"Logs Retrieval Failed",
			"Failed to retrieve user activity logs",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, logs)
}

// GetUserActivity godoc
// @Summary Get user activity summary
// @Description Get recent activity summary for the authenticated user
// @Tags logs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param days query int false "Number of days to look back" default(7)
// @Success 200 {object} UserActivityResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /logs/my-activity/summary [get]
func (h *LogHandler) GetUserActivity(c *gin.Context) {
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

	// Parse days parameter
	days := 7
	if daysParam, err := strconv.Atoi(c.DefaultQuery("days", "7")); err == nil && daysParam > 0 && daysParam <= 365 {
		days = daysParam
	}

	// Get user activity
	activities, err := h.logRepo.GetUserActivity(c.Request.Context(), userClaims.UserID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"Activity Retrieval Failed",
			"Failed to retrieve user activity",
			err.Error(),
		))
		return
	}

	// Get event statistics
	eventStats, err := h.logRepo.GetEventStats(c.Request.Context(), &userClaims.UserID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"Stats Retrieval Failed",
			"Failed to retrieve activity statistics",
			err.Error(),
		))
		return
	}

	response := UserActivityResponse{
		UserID:         userClaims.UserID,
		DaysRequested:  days,
		TotalActivities: len(activities),
		EventStats:     eventStats,
		RecentActivity: activities,
	}

	c.JSON(http.StatusOK, response)
}

// SearchLogs godoc
// @Summary Search logs
// @Description Search through activity logs (admin only)
// @Tags logs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param user_id query string false "Filter by user ID"
// @Param event query string false "Filter by event type"
// @Param start_date query string false "Start date (RFC3339)"
// @Param end_date query string false "End date (RFC3339)"
// @Success 200 {object} models.UserLogsListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /logs/search [get]
func (h *LogHandler) SearchLogs(c *gin.Context) {
	// Check admin role
	if !middleware.IsAdmin(c) {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			http.StatusForbidden,
			"Forbidden",
			"Admin access required",
			nil,
		))
		return
	}

	// Get search query
	searchTerm := c.Query("q")
	if searchTerm == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			http.StatusBadRequest,
			"Missing Search Query",
			"Search query parameter 'q' is required",
			nil,
		))
		return
	}

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

	// Parse date filters
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}

	// Perform search
	logs, err := h.logRepo.SearchLogs(c.Request.Context(), searchTerm, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"Search Failed",
			"Failed to search logs",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, logs)
}

// GetEventStats godoc
// @Summary Get event statistics
// @Description Get statistics about different event types (admin only)
// @Tags logs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param days query int false "Number of days to look back" default(30)
// @Param user_id query string false "Filter by user ID"
// @Success 200 {object} EventStatsResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /logs/stats [get]
func (h *LogHandler) GetEventStats(c *gin.Context) {
	// Check admin role
	if !middleware.IsAdmin(c) {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			http.StatusForbidden,
			"Forbidden",
			"Admin access required",
			nil,
		))
		return
	}

	// Parse parameters
	days := 30
	if daysParam, err := strconv.Atoi(c.DefaultQuery("days", "30")); err == nil && daysParam > 0 && daysParam <= 365 {
		days = daysParam
	}

	var userID *uuid.UUID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if parsedUserID, err := uuid.Parse(userIDStr); err == nil {
			userID = &parsedUserID
		}
	}

	// Get event statistics
	eventStats, err := h.logRepo.GetEventStats(c.Request.Context(), userID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"Stats Retrieval Failed",
			"Failed to retrieve event statistics",
			err.Error(),
		))
		return
	}

	// Calculate total events
	var totalEvents int64
	for _, count := range eventStats {
		totalEvents += count
	}

	response := EventStatsResponse{
		DaysRequested: days,
		UserID:        userID,
		TotalEvents:   totalEvents,
		EventStats:    eventStats,
		GeneratedAt:   time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// GetLogDetails godoc
// @Summary Get log entry details
// @Description Get detailed information about a specific log entry
// @Tags logs
// @Security BearerAuth
// @Produce json
// @Param id path string true "Log entry ID"
// @Success 200 {object} models.UserLogResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /logs/{id} [get]
func (h *LogHandler) GetLogDetails(c *gin.Context) {
	logID := c.Param("id")
	if logID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			http.StatusBadRequest,
			"Invalid Log ID",
			"Log ID is required",
			nil,
		))
		return
	}

	// Get log entry
	logEntry, err := h.logRepo.GetByID(c.Request.Context(), logID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			http.StatusNotFound,
			"Log Not Found",
			"Log entry with the specified ID was not found",
			err.Error(),
		))
		return
	}

	// Check if user can access this log entry
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

	// Allow access if user is admin or if it's their own log entry
	if !middleware.IsAdmin(c) && logEntry.UserID != nil && *logEntry.UserID != userClaims.UserID.String() {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			http.StatusForbidden,
			"Forbidden",
			"You can only access your own log entries",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, logEntry.ToResponse())
}

// Response types

type UserActivityResponse struct {
	UserID          uuid.UUID                          `json:"user_id"`
	DaysRequested   int                                `json:"days_requested"`
	TotalActivities int                                `json:"total_activities"`
	EventStats      map[models.LogEventType]int64      `json:"event_stats"`
	RecentActivity  []models.UserLogResponse           `json:"recent_activity"`
}

type EventStatsResponse struct {
	DaysRequested int                           `json:"days_requested"`
	UserID        *uuid.UUID                    `json:"user_id,omitempty"`
	TotalEvents   int64                         `json:"total_events"`
	EventStats    map[models.LogEventType]int64 `json:"event_stats"`
	GeneratedAt   time.Time                     `json:"generated_at"`
}

// GetAvailableEventTypes godoc
// @Summary Get available event types
// @Description Get list of all available log event types
// @Tags logs
// @Security BearerAuth
// @Produce json
// @Success 200 {object} AvailableEventTypesResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /logs/event-types [get]
func (h *LogHandler) GetAvailableEventTypes(c *gin.Context) {
	eventTypes := models.GetValidEventTypes()
	
	response := AvailableEventTypesResponse{
		EventTypes:  eventTypes,
		TotalCount:  len(eventTypes),
		GeneratedAt: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

type AvailableEventTypesResponse struct {
	EventTypes  []models.LogEventType `json:"event_types"`
	TotalCount  int                   `json:"total_count"`
	GeneratedAt time.Time             `json:"generated_at"`
} 