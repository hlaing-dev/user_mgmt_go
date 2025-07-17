package models

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LogEventType represents the type of log event
type LogEventType string

const (
	// User-related events
	UserCreated LogEventType = "USER_CREATED"
	UserUpdated LogEventType = "USER_UPDATED"
	UserDeleted LogEventType = "USER_DELETED"
	UserLogin   LogEventType = "USER_LOGIN"
	
	// Admin-related events
	AdminLogin     LogEventType = "ADMIN_LOGIN"
	AdminLogout    LogEventType = "ADMIN_LOGOUT"
	
	// Authentication events
	LoginSuccess   LogEventType = "LOGIN_SUCCESS"
	LoginFailed    LogEventType = "LOGIN_FAILED"
	TokenRefresh   LogEventType = "TOKEN_REFRESH"
	
	// System events
	SystemError         LogEventType = "SYSTEM_ERROR"
	ValidationLogError  LogEventType = "VALIDATION_ERROR"
)

// UserLog represents the log entry stored in MongoDB
type UserLog struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    *string            `json:"user_id,omitempty" bson:"user_id,omitempty"` // UUID as string, nullable for system events
	Event     LogEventType       `json:"event" bson:"event"`
	Data      LogData            `json:"data" bson:"data"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
	IPAddress string             `json:"ip_address,omitempty" bson:"ip_address,omitempty"`
	UserAgent string             `json:"user_agent,omitempty" bson:"user_agent,omitempty"`
}

// LogData contains the actual log data with flexible structure
type LogData struct {
	Action      string                 `json:"action" bson:"action"`
	Details     map[string]interface{} `json:"details" bson:"details"`
	OldValues   map[string]interface{} `json:"old_values,omitempty" bson:"old_values,omitempty"`     // For update operations
	NewValues   map[string]interface{} `json:"new_values,omitempty" bson:"new_values,omitempty"`     // For update operations
	Error       string                 `json:"error,omitempty" bson:"error,omitempty"`               // For error events
	Duration    int64                  `json:"duration,omitempty" bson:"duration,omitempty"`         // Request duration in milliseconds
	StatusCode  int                    `json:"status_code,omitempty" bson:"status_code,omitempty"`   // HTTP status code
}

// UserLogCreateRequest represents the request to create a log entry
type UserLogCreateRequest struct {
	UserID    *uuid.UUID             `json:"user_id,omitempty"`
	Event     LogEventType           `json:"event"`
	Action    string                 `json:"action"`
	Details   map[string]interface{} `json:"details,omitempty"`
	OldValues map[string]interface{} `json:"old_values,omitempty"`
	NewValues map[string]interface{} `json:"new_values,omitempty"`
	Error     string                 `json:"error,omitempty"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
}

// UserLogResponse represents the response payload for log data
type UserLogResponse struct {
	ID        string       `json:"id"`
	UserID    *uuid.UUID   `json:"user_id,omitempty"`
	Event     LogEventType `json:"event"`
	Data      LogData      `json:"data"`
	Timestamp time.Time    `json:"timestamp"`
	IPAddress string       `json:"ip_address,omitempty"`
	UserAgent string       `json:"user_agent,omitempty"`
}

// UserLogsListResponse represents the response payload for paginated log list
type UserLogsListResponse struct {
	Logs       []UserLogResponse `json:"logs"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// LogFilterRequest represents the request payload for filtering logs
type LogFilterRequest struct {
	UserID     *uuid.UUID    `json:"user_id,omitempty" form:"user_id"`
	Event      *LogEventType `json:"event,omitempty" form:"event"`
	StartDate  *time.Time    `json:"start_date,omitempty" form:"start_date"`
	EndDate    *time.Time    `json:"end_date,omitempty" form:"end_date"`
	IPAddress  string        `json:"ip_address,omitempty" form:"ip_address"`
	Action     string        `json:"action,omitempty" form:"action"`
	Page       int           `json:"page" form:"page" binding:"omitempty,min=1"`
	PageSize   int           `json:"page_size" form:"page_size" binding:"omitempty,min=1,max=100"`
}

// NewUserLog creates a new UserLog instance
func NewUserLog(req UserLogCreateRequest) *UserLog {
	var userIDStr *string
	if req.UserID != nil {
		userIDString := req.UserID.String()
		userIDStr = &userIDString
	}

	return &UserLog{
		UserID: userIDStr,
		Event:  req.Event,
		Data: LogData{
			Action:    req.Action,
			Details:   req.Details,
			OldValues: req.OldValues,
			NewValues: req.NewValues,
			Error:     req.Error,
		},
		Timestamp: time.Now(),
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
	}
}

// ToResponse converts UserLog model to UserLogResponse
func (ul *UserLog) ToResponse() UserLogResponse {
	var userID *uuid.UUID
	if ul.UserID != nil {
		if parsedUUID, err := uuid.Parse(*ul.UserID); err == nil {
			userID = &parsedUUID
		}
	}

	return UserLogResponse{
		ID:        ul.ID.Hex(),
		UserID:    userID,
		Event:     ul.Event,
		Data:      ul.Data,
		Timestamp: ul.Timestamp,
		IPAddress: ul.IPAddress,
		UserAgent: ul.UserAgent,
	}
}

// CollectionName returns the MongoDB collection name
func (UserLog) CollectionName() string {
	return "user_logs"
}

// GetValidEventTypes returns all valid event types
func GetValidEventTypes() []LogEventType {
	return []LogEventType{
		UserCreated,
		UserUpdated,
		UserDeleted,
		UserLogin,
		AdminLogin,
		AdminLogout,
		LoginSuccess,
		LoginFailed,
		TokenRefresh,
		SystemError,
		ValidationLogError,
	}
}

// IsValidEventType checks if an event type is valid
func IsValidEventType(event LogEventType) bool {
	validTypes := GetValidEventTypes()
	for _, validType := range validTypes {
		if event == validType {
			return true
		}
	}
	return false
} 