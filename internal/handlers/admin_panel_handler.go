package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"time"

	"user_mgmt_go/internal/middleware"
	"user_mgmt_go/internal/models"
	"user_mgmt_go/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminPanelHandler handles admin panel web interface
type AdminPanelHandler struct {
	userRepo    repository.UserRepository
	logRepo     repository.UserLogRepository
	repoManager *repository.RepositoryManager
	templates   *template.Template
}

// NewAdminPanelHandler creates a new admin panel handler
func NewAdminPanelHandler(
	userRepo repository.UserRepository,
	logRepo repository.UserLogRepository,
	repoManager *repository.RepositoryManager,
) *AdminPanelHandler {
	handler := &AdminPanelHandler{
		userRepo:    userRepo,
		logRepo:     logRepo,
		repoManager: repoManager,
	}
	
	// Load templates
	handler.loadTemplates()
	
	return handler
}

// loadTemplates loads all HTML templates
func (h *AdminPanelHandler) loadTemplates() {
	var err error
	h.templates, err = template.New("").Funcs(template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
	}).ParseGlob("templates/admin/*.html")
	
	if err != nil {
		// Templates not loaded, will create them
		h.templates = template.New("")
	}
}

// PageData represents common data for all admin pages
type PageData struct {
	Title       string
	CurrentUser *models.UserResponse
	CurrentTime time.Time
	Data        interface{}
}

// DashboardData represents data for the admin dashboard
type DashboardData struct {
	Stats       map[string]interface{}
	UserCount   int64
	LogCount    int64
	RecentUsers []models.UserResponse
	RecentLogs  []models.UserLogResponse
}

// Dashboard renders the admin dashboard
func (h *AdminPanelHandler) Dashboard(c *gin.Context) {
	user := h.getCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusTemporaryRedirect, "/admin/login")
		return
	}

	// Get system stats
	stats, _ := h.repoManager.GetStats(c.Request.Context())
	
	// Get recent users
	recentUsersResp, _ := h.userRepo.List(c.Request.Context(), repository.ListParams{
		Page: 1, PageSize: 5, SortBy: "created_at", SortDir: "desc",
	})
	
	// Get recent logs
	recentLogsResp, _ := h.logRepo.List(c.Request.Context(), models.LogFilterRequest{
		Page: 1, PageSize: 5,
	})

	data := DashboardData{
		Stats:       stats,
		UserCount:   recentUsersResp.Total,
		LogCount:    int64(len(recentLogsResp.Logs)),
		RecentUsers: recentUsersResp.Users,
		RecentLogs:  recentLogsResp.Logs,
	}

	pageData := PageData{
		Title:       "Admin Dashboard",
		CurrentUser: user,
		CurrentTime: time.Now(),
		Data:        data,
	}

	h.renderTemplate(c, "dashboard", pageData)
}

// Users renders the users management page
func (h *AdminPanelHandler) Users(c *gin.Context) {
	user := h.getCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusTemporaryRedirect, "/admin/login")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")

	params := repository.ListParams{
		Page: page, PageSize: pageSize, SortBy: "created_at", SortDir: "desc",
	}

	var usersResp *models.UsersListResponse
	var err error

	if search != "" {
		usersResp, err = h.userRepo.Search(c.Request.Context(), search, params)
	} else {
		usersResp, err = h.userRepo.List(c.Request.Context(), params)
	}

	if err != nil {
		usersResp = &models.UsersListResponse{}
	}

	pageData := PageData{
		Title:       "User Management",
		CurrentUser: user,
		CurrentTime: time.Now(),
		Data:        usersResp,
	}

	h.renderTemplate(c, "users", pageData)
}

// Logs renders the logs management page
func (h *AdminPanelHandler) Logs(c *gin.Context) {
	user := h.getCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusTemporaryRedirect, "/admin/login")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	userID := c.Query("user_id")
	event := c.Query("event")

	filter := models.LogFilterRequest{
		Page: page, PageSize: pageSize,
	}

	if userID != "" {
		if parsedUserID, err := uuid.Parse(userID); err == nil {
			filter.UserID = &parsedUserID
		}
	}

	if event != "" {
		eventType := models.LogEventType(event)
		if models.IsValidEventType(eventType) {
			filter.Event = &eventType
		}
	}

	logsResp, err := h.logRepo.List(c.Request.Context(), filter)
	if err != nil {
		logsResp = &models.UserLogsListResponse{}
	}

	pageData := PageData{
		Title:       "Activity Logs",
		CurrentUser: user,
		CurrentTime: time.Now(),
		Data:        logsResp,
	}

	h.renderTemplate(c, "logs", pageData)
}

// Stats renders the system statistics page
func (h *AdminPanelHandler) Stats(c *gin.Context) {
	user := h.getCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusTemporaryRedirect, "/admin/login")
		return
	}

	stats, _ := h.repoManager.GetStats(c.Request.Context())
	health := h.repoManager.HealthCheck()
	
	eventStats, _ := h.logRepo.GetEventStats(c.Request.Context(), nil, 30)

	data := map[string]interface{}{
		"system_stats": stats,
		"health":       health,
		"event_stats":  eventStats,
	}

	pageData := PageData{
		Title:       "System Statistics",
		CurrentUser: user,
		CurrentTime: time.Now(),
		Data:        data,
	}

	h.renderTemplate(c, "stats", pageData)
}

// DeletedUsers renders the deleted users management page
func (h *AdminPanelHandler) DeletedUsers(c *gin.Context) {
	user := h.getCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusTemporaryRedirect, "/admin/login")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	params := repository.ListParams{
		Page: page, PageSize: pageSize, SortBy: "deleted_at", SortDir: "desc",
	}

	deletedUsers, err := h.userRepo.GetAllDeleted(c.Request.Context(), params)
	if err != nil {
		deletedUsers = &models.UsersListResponse{}
	}

	pageData := PageData{
		Title:       "Deleted Users",
		CurrentUser: user,
		CurrentTime: time.Now(),
		Data:        deletedUsers,
	}

	h.renderTemplate(c, "deleted-users", pageData)
}

// Login renders the admin login page
func (h *AdminPanelHandler) Login(c *gin.Context) {
	// Check if already logged in
	if user := h.getCurrentUser(c); user != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/admin/dashboard")
		return
	}

	pageData := PageData{
		Title:       "Admin Login",
		CurrentTime: time.Now(),
	}

	h.renderTemplate(c, "login", pageData)
}

// Helper methods

func (h *AdminPanelHandler) getCurrentUser(c *gin.Context) *models.UserResponse {
	userClaims, exists := middleware.GetUserFromContext(c)
	if !exists {
		return nil
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userClaims.UserID)
	if err != nil {
		return nil
	}

	response := user.ToResponse()
	return &response
}

func (h *AdminPanelHandler) renderTemplate(c *gin.Context, templateName string, data PageData) {
	// Try with .html extension first
	templateWithExt := templateName + ".html"
	
	// If templates are not loaded, render JSON fallback
	if h.templates.Lookup(templateWithExt) == nil {
		c.JSON(http.StatusOK, data)
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(c.Writer, templateWithExt, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// SetupAdminPanelRoutes sets up the admin panel routes
func (h *AdminPanelHandler) SetupAdminPanelRoutes(router *gin.Engine, middlewareManager *middleware.MiddlewareManager) {
	admin := router.Group("/admin")

	// Public login page
	admin.GET("/login", h.Login)

	// Protected admin pages
	protected := admin.Group("")
	protected.Use(middlewareManager.AuthMiddleware())
	protected.Use(middlewareManager.AdminRequiredMiddleware())
	{
		protected.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusTemporaryRedirect, "/admin/dashboard")
		})
		protected.GET("/dashboard", h.Dashboard)
		protected.GET("/users", h.Users)
		protected.GET("/logs", h.Logs)
		protected.GET("/stats", h.Stats)
		protected.GET("/deleted-users", h.DeletedUsers)
	}

	// Serve static files for admin panel
	router.Static("/admin/static", "./templates/admin/static")
} 