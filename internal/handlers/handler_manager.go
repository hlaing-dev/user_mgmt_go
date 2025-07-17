package handlers

import (
	"user_mgmt_go/internal/middleware"
	"user_mgmt_go/internal/repository"
	"user_mgmt_go/internal/utils"

	"github.com/gin-gonic/gin"
)

// HandlerManager manages all API handlers
type HandlerManager struct {
	AuthHandler       *AuthHandler
	UserHandler       *UserHandler
	AdminHandler      *AdminHandler
	AdminPanelHandler *AdminPanelHandler
	LogHandler        *LogHandler
	
	middlewareManager *middleware.MiddlewareManager
}

// NewHandlerManager creates a new handler manager with all dependencies
func NewHandlerManager(
	jwtManager *utils.JWTManager,
	repoManager *repository.RepositoryManager,
	middlewareManager *middleware.MiddlewareManager,
) *HandlerManager {
	return &HandlerManager{
		AuthHandler: NewAuthHandler(
			jwtManager,
			repoManager.Repos.User,
			repoManager.Repos.Log,
		),
		UserHandler: NewUserHandler(
			repoManager.Repos.User,
			repoManager.Repos.Log,
		),
		AdminHandler: NewAdminHandler(
			repoManager.Repos.User,
			repoManager.Repos.Log,
			repoManager,
		),
		AdminPanelHandler: NewAdminPanelHandler(
			repoManager.Repos.User,
			repoManager.Repos.Log,
			repoManager,
		),
		LogHandler: NewLogHandler(
			repoManager.Repos.Log,
		),
		middlewareManager: middlewareManager,
	}
}

// SetupRoutes configures all API routes with appropriate middleware
func (hm *HandlerManager) SetupRoutes(router *gin.Engine) {
	// Setup global middleware
	hm.middlewareManager.SetupGlobalMiddleware(router)

	// Setup admin panel web interface routes
	hm.AdminPanelHandler.SetupAdminPanelRoutes(router, hm.middlewareManager)

	// API root
	api := router.Group("/api")

	// Setup route groups with pre-configured middleware
	hm.setupAuthRoutes(api)
	hm.setupUserRoutes(api)
	hm.setupAdminRoutes(api)
	hm.setupLogRoutes(api)
	hm.setupUtilityRoutes(api)
}

// setupAuthRoutes configures authentication-related routes
func (hm *HandlerManager) setupAuthRoutes(api *gin.RouterGroup) {
	auth := api.Group("/auth")
	
	// Public authentication endpoints
	auth.POST("/login", hm.AuthHandler.Login)
	auth.POST("/refresh", hm.AuthHandler.RefreshToken)
	
	// Protected authentication endpoints
	authProtected := auth.Group("")
	authProtected.Use(hm.middlewareManager.AuthMiddleware())
	{
		authProtected.POST("/logout", hm.AuthHandler.Logout)
		authProtected.GET("/profile", hm.AuthHandler.GetProfile)
		authProtected.POST("/change-password", hm.AuthHandler.ChangePassword)
	}
}

// setupUserRoutes configures user management routes
func (hm *HandlerManager) setupUserRoutes(api *gin.RouterGroup) {
	users := api.Group("/users")
	users.Use(hm.middlewareManager.AuthMiddleware())
	
	// User CRUD operations (admin required)
	{
		users.GET("", hm.middlewareManager.AdminRequiredMiddleware(), hm.UserHandler.ListUsers)
		users.POST("", hm.middlewareManager.AdminRequiredMiddleware(), hm.UserHandler.CreateUser)
		users.GET("/:id", hm.middlewareManager.SelfOrAdminMiddleware("id"), hm.UserHandler.GetUser)
		users.PUT("/:id", hm.middlewareManager.SelfOrAdminMiddleware("id"), hm.UserHandler.UpdateUser)
		users.DELETE("/:id", hm.middlewareManager.AdminRequiredMiddleware(), hm.UserHandler.DeleteUser)
	}
}

// setupAdminRoutes configures admin-specific routes
func (hm *HandlerManager) setupAdminRoutes(api *gin.RouterGroup) {
	admin := api.Group("/admin")
	admin.Use(hm.middlewareManager.AuthMiddleware())
	admin.Use(hm.middlewareManager.AdminRequiredMiddleware())
	
	// System management
	{
		admin.GET("/stats", hm.AdminHandler.GetSystemStats)
		admin.POST("/maintenance", hm.AdminHandler.RunMaintenance)
	}
	
	// Advanced user management
	{
		admin.GET("/users/deleted", hm.AdminHandler.GetDeletedUsers)
		admin.POST("/users/:id/restore", hm.AdminHandler.RestoreUser)
		admin.DELETE("/users/:id/permanent-delete", hm.AdminHandler.PermanentDeleteUser)
		admin.POST("/users/bulk-create", hm.AdminHandler.BulkCreateUsers)
	}
	
	// Admin log access
	{
		admin.GET("/logs", hm.AdminHandler.GetUserLogs)
	}
}

// setupLogRoutes configures log management routes
func (hm *HandlerManager) setupLogRoutes(api *gin.RouterGroup) {
	logs := api.Group("/logs")
	logs.Use(hm.middlewareManager.AuthMiddleware())
	
	// User's own activity logs
	{
		logs.GET("/my-activity", hm.LogHandler.GetUserLogs)
		logs.GET("/my-activity/summary", hm.LogHandler.GetUserActivity)
	}
	
	// Admin-only log operations
	logsAdmin := logs.Group("")
	logsAdmin.Use(hm.middlewareManager.AdminRequiredMiddleware())
	{
		logsAdmin.GET("/search", hm.LogHandler.SearchLogs)
		logsAdmin.GET("/stats", hm.LogHandler.GetEventStats)
		logsAdmin.GET("/:id", hm.LogHandler.GetLogDetails)
	}
	
	// Public (authenticated) endpoints
	{
		logs.GET("/event-types", hm.LogHandler.GetAvailableEventTypes)
	}
}

// setupUtilityRoutes configures utility and meta routes
func (hm *HandlerManager) setupUtilityRoutes(api *gin.RouterGroup) {
	// Public endpoints
	api.GET("/ping", hm.handlePing)
	api.GET("/version", hm.handleVersion)
	
	// Health check is handled by global middleware
	// but we can add a more detailed version here
	api.GET("/health/detailed", hm.middlewareManager.AuthMiddleware(), hm.handleDetailedHealth)
}

// Utility handlers

// handlePing provides a simple ping endpoint
func (hm *HandlerManager) handlePing(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
		"status":  "ok",
	})
}

// handleVersion provides version information
func (hm *HandlerManager) handleVersion(c *gin.Context) {
	c.JSON(200, gin.H{
		"version":     "1.0.0",
		"api_version": "v1",
		"service":     "user_mgmt_go",
		"description": "User Management System API",
	})
}

// handleDetailedHealth provides detailed health information (authenticated)
func (hm *HandlerManager) handleDetailedHealth(c *gin.Context) {
	// This would typically include more detailed health checks
	// For now, we'll return basic information
	c.JSON(200, gin.H{
		"status":    "healthy",
		"timestamp": gin.H{"time": "now"},
		"services": gin.H{
			"database": true,
			"mongodb":  true,
		},
		"version": "1.0.0",
	})
}

// GetRouteSummary returns a summary of all available routes
func (hm *HandlerManager) GetRouteSummary() map[string][]RouteInfo {
	return map[string][]RouteInfo{
		"Authentication": {
			{Method: "POST", Path: "/api/auth/login", Description: "Admin login", Auth: "Public"},
			{Method: "POST", Path: "/api/auth/refresh", Description: "Refresh token", Auth: "Public"},
			{Method: "POST", Path: "/api/auth/logout", Description: "User logout", Auth: "Required"},
			{Method: "GET", Path: "/api/auth/profile", Description: "Get user profile", Auth: "Required"},
			{Method: "POST", Path: "/api/auth/change-password", Description: "Change password", Auth: "Required"},
		},
		"User Management": {
			{Method: "GET", Path: "/api/users", Description: "List users", Auth: "Admin"},
			{Method: "POST", Path: "/api/users", Description: "Create user", Auth: "Admin"},
			{Method: "GET", Path: "/api/users/:id", Description: "Get user", Auth: "Self or Admin"},
			{Method: "PUT", Path: "/api/users/:id", Description: "Update user", Auth: "Self or Admin"},
			{Method: "DELETE", Path: "/api/users/:id", Description: "Delete user", Auth: "Admin"},
		},
		"Admin Operations": {
			{Method: "GET", Path: "/api/admin/stats", Description: "System statistics", Auth: "Admin"},
			{Method: "POST", Path: "/api/admin/maintenance", Description: "Run maintenance", Auth: "Admin"},
			{Method: "GET", Path: "/api/admin/users/deleted", Description: "Get deleted users", Auth: "Admin"},
			{Method: "POST", Path: "/api/admin/users/:id/restore", Description: "Restore user", Auth: "Admin"},
			{Method: "DELETE", Path: "/api/admin/users/:id/permanent-delete", Description: "Permanent delete", Auth: "Admin"},
			{Method: "POST", Path: "/api/admin/users/bulk-create", Description: "Bulk create users", Auth: "Admin"},
			{Method: "GET", Path: "/api/admin/logs", Description: "Get all logs", Auth: "Admin"},
		},
		"Logs & Monitoring": {
			{Method: "GET", Path: "/api/logs/my-activity", Description: "User's activity logs", Auth: "Required"},
			{Method: "GET", Path: "/api/logs/my-activity/summary", Description: "Activity summary", Auth: "Required"},
			{Method: "GET", Path: "/api/logs/search", Description: "Search logs", Auth: "Admin"},
			{Method: "GET", Path: "/api/logs/stats", Description: "Event statistics", Auth: "Admin"},
			{Method: "GET", Path: "/api/logs/:id", Description: "Log details", Auth: "Required"},
			{Method: "GET", Path: "/api/logs/event-types", Description: "Available event types", Auth: "Required"},
		},
		"Utilities": {
			{Method: "GET", Path: "/api/ping", Description: "Simple ping", Auth: "Public"},
			{Method: "GET", Path: "/api/version", Description: "Version info", Auth: "Public"},
			{Method: "GET", Path: "/health", Description: "Health check", Auth: "Public"},
			{Method: "GET", Path: "/api/health/detailed", Description: "Detailed health", Auth: "Required"},
		},
	}
}

// RouteInfo represents information about an API route
type RouteInfo struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description"`
	Auth        string `json:"auth"`
}

// SetupDocumentationRoute adds a route that shows all available endpoints
func (hm *HandlerManager) SetupDocumentationRoute(router *gin.Engine) {
	router.GET("/api/docs/routes", func(c *gin.Context) {
		routes := hm.GetRouteSummary()
		c.JSON(200, gin.H{
			"title":       "User Management API",
			"version":     "1.0.0",
			"description": "Complete API documentation for user management system",
			"routes":      routes,
		})
	})
}

// Close performs cleanup for all handlers
func (hm *HandlerManager) Close() {
	// Currently no cleanup needed for handlers
	// This is a placeholder for future cleanup requirements
}

// ValidateRoutes can be used to validate that all routes are properly configured
func (hm *HandlerManager) ValidateRoutes() error {
	// This could include validation logic for routes
	// For now, we'll just return nil
	return nil
} 