package repository

import (
	"context"

	"user_mgmt_go/internal/models"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// List operations with pagination and filtering
	List(ctx context.Context, params ListParams) (*models.UsersListResponse, error)
	Count(ctx context.Context, filter UserFilter) (int64, error)
	
	// Bulk operations
	CreateBatch(ctx context.Context, users []*models.User) error
	DeleteBatch(ctx context.Context, ids []uuid.UUID) error
	
	// Search and filtering
	Search(ctx context.Context, query string, params ListParams) (*models.UsersListResponse, error)
	Exists(ctx context.Context, email string) (bool, error)
	
	// Admin operations
	GetAllDeleted(ctx context.Context, params ListParams) (*models.UsersListResponse, error)
	RestoreDeleted(ctx context.Context, id uuid.UUID) error
	PermanentDelete(ctx context.Context, id uuid.UUID) error
}

// UserLogRepository defines the interface for logging operations
type UserLogRepository interface {
	// Basic log operations
	Create(ctx context.Context, log *models.UserLog) error
	CreateAsync(log *models.UserLog) error // Asynchronous logging
	GetByID(ctx context.Context, id string) (*models.UserLog, error)
	
	// List operations with advanced filtering
	List(ctx context.Context, filter models.LogFilterRequest) (*models.UserLogsListResponse, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, params ListParams) (*models.UserLogsListResponse, error)
	GetByEvent(ctx context.Context, event models.LogEventType, params ListParams) (*models.UserLogsListResponse, error)
	
	// Analytics and reporting
	Count(ctx context.Context, filter models.LogFilterRequest) (int64, error)
	GetEventStats(ctx context.Context, userID *uuid.UUID, days int) (map[models.LogEventType]int64, error)
	GetUserActivity(ctx context.Context, userID uuid.UUID, days int) ([]models.UserLogResponse, error)
	
	// Maintenance operations
	DeleteOldLogs(ctx context.Context, olderThanDays int) (int64, error)
	BulkCreate(ctx context.Context, logs []*models.UserLog) error
	
	// Search operations
	SearchLogs(ctx context.Context, searchTerm string, filter models.LogFilterRequest) (*models.UserLogsListResponse, error)
}

// Repository aggregates all repository interfaces
type Repository struct {
	User UserRepository
	Log  UserLogRepository
}

// ListParams defines common pagination and sorting parameters
type ListParams struct {
	Page     int    `json:"page" form:"page" binding:"omitempty,min=1"`
	PageSize int    `json:"page_size" form:"page_size" binding:"omitempty,min=1,max=100"`
	SortBy   string `json:"sort_by" form:"sort_by"`
	SortDir  string `json:"sort_dir" form:"sort_dir" binding:"omitempty,oneof=asc desc"`
}

// UserFilter defines filtering options for user queries
type UserFilter struct {
	Email     string    `json:"email" form:"email"`
	Name      string    `json:"name" form:"name"`
	CreatedAt *TimeRange `json:"created_at" form:"created_at"`
	UpdatedAt *TimeRange `json:"updated_at" form:"updated_at"`
	IsDeleted *bool     `json:"is_deleted" form:"is_deleted"`
}

// TimeRange defines a time range filter
type TimeRange struct {
	From *string `json:"from" form:"from"` // ISO 8601 format
	To   *string `json:"to" form:"to"`     // ISO 8601 format
}

// SetDefaults sets default values for ListParams
func (lp *ListParams) SetDefaults() {
	if lp.Page <= 0 {
		lp.Page = 1
	}
	if lp.PageSize <= 0 {
		lp.PageSize = 10
	}
	if lp.PageSize > 100 {
		lp.PageSize = 100
	}
	if lp.SortBy == "" {
		lp.SortBy = "created_at"
	}
	if lp.SortDir == "" {
		lp.SortDir = "desc"
	}
}

// GetOffset calculates the offset for pagination
func (lp *ListParams) GetOffset() int {
	return (lp.Page - 1) * lp.PageSize
}

// GetLimit returns the page size as limit
func (lp *ListParams) GetLimit() int {
	return lp.PageSize
}

// Validate validates the ListParams
func (lp *ListParams) Validate() error {
	if lp.Page < 1 {
		lp.Page = 1
	}
	if lp.PageSize < 1 || lp.PageSize > 100 {
		lp.PageSize = 10
	}
	if lp.SortDir != "asc" && lp.SortDir != "desc" {
		lp.SortDir = "desc"
	}
	return nil
}

// CalculateTotalPages calculates total pages based on total count and page size
func CalculateTotalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	return int((total + int64(pageSize) - 1) / int64(pageSize))
}

// IsValidSortField checks if a sort field is valid for users
func IsValidUserSortField(field string) bool {
	validFields := map[string]bool{
		"id":         true,
		"name":       true,
		"email":      true,
		"created_at": true,
		"updated_at": true,
	}
	return validFields[field]
}

// IsValidLogSortField checks if a sort field is valid for logs
func IsValidLogSortField(field string) bool {
	validFields := map[string]bool{
		"timestamp": true,
		"event":     true,
		"user_id":   true,
	}
	return validFields[field]
} 