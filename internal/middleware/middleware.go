package middleware

import (
	"time"

	"user_mgmt_go/internal/config"
	"user_mgmt_go/internal/repository"
	"user_mgmt_go/internal/utils"

	"github.com/gin-gonic/gin"
)

// MiddlewareManager manages all middleware components
type MiddlewareManager struct {
	config      *config.Config
	jwtManager  *utils.JWTManager
	rateLimiter *RateLimiter
	repoManager *repository.RepositoryManager
}

// NewMiddlewareManager creates a new middleware manager
func NewMiddlewareManager(
	cfg *config.Config,
	jwtManager *utils.JWTManager,
	repoManager *repository.RepositoryManager,
) *MiddlewareManager {
	// Create rate limiter (100 requests per minute with burst of 20)
	rateLimiter := NewRateLimiter(time.Minute/100, 20)

	return &MiddlewareManager{
		config:      cfg,
		jwtManager:  jwtManager,
		rateLimiter: rateLimiter,
		repoManager: repoManager,
	}
}

// SetupGlobalMiddleware configures global middleware for the Gin router
func (mm *MiddlewareManager) SetupGlobalMiddleware(router *gin.Engine) {
	// Health check middleware (should be first)
	router.Use(mm.HealthCheckMiddleware())

	// Security headers
	router.Use(SecurityHeadersMiddleware())

	// CORS
	router.Use(CORSMiddleware(mm.config.CORS.AllowedOrigins))

	// Rate limiting
	router.Use(RateLimitMiddleware(mm.rateLimiter))

	// Request size limit (10MB)
	router.Use(RequestSizeLimitMiddleware(10 * 1024 * 1024))

	// Request timeout (30 seconds)
	router.Use(TimeoutMiddleware(30 * time.Second))

	// Recovery with logging
	router.Use(RecoveryMiddleware(mm.repoManager.Repos.Log))

	// Request logging (should be last to capture all request data)
	router.Use(RequestLoggingMiddleware(mm.repoManager.Repos.Log))
}

// AuthMiddleware returns the authentication middleware
func (mm *MiddlewareManager) AuthMiddleware() gin.HandlerFunc {
	return AuthMiddleware(mm.jwtManager)
}

// OptionalAuthMiddleware returns the optional authentication middleware
func (mm *MiddlewareManager) OptionalAuthMiddleware() gin.HandlerFunc {
	return OptionalAuthMiddleware(mm.jwtManager)
}

// AdminRequiredMiddleware returns the admin required middleware
func (mm *MiddlewareManager) AdminRequiredMiddleware() gin.HandlerFunc {
	return AdminRequiredMiddleware()
}

// RoleRequiredMiddleware returns the role required middleware
func (mm *MiddlewareManager) RoleRequiredMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return RoleRequiredMiddleware(requiredRoles...)
}

// SelfOrAdminMiddleware returns the self or admin middleware
func (mm *MiddlewareManager) SelfOrAdminMiddleware(userIDParam string) gin.HandlerFunc {
	return SelfOrAdminMiddleware(userIDParam)
}

// HealthCheckMiddleware returns the health check middleware
func (mm *MiddlewareManager) HealthCheckMiddleware() gin.HandlerFunc {
	return HealthCheckMiddleware(mm.repoManager)
}

// IPWhitelistMiddleware returns the IP whitelist middleware
func (mm *MiddlewareManager) IPWhitelistMiddleware(whitelist []string) gin.HandlerFunc {
	return IPWhitelistMiddleware(whitelist)
}

// StrictRateLimitMiddleware returns a stricter rate limiter for sensitive endpoints
func (mm *MiddlewareManager) StrictRateLimitMiddleware() gin.HandlerFunc {
	strictLimiter := NewRateLimiter(time.Minute/10, 5) // 10 requests per minute, burst of 5
	return RateLimitMiddleware(strictLimiter)
}

// LoggingOnlyMiddleware returns a middleware that only logs without other security measures
func (mm *MiddlewareManager) LoggingOnlyMiddleware() gin.HandlerFunc {
	return RequestLoggingMiddleware(mm.repoManager.Repos.Log)
}

// AuthChain returns a chain of authentication and authorization middleware
type AuthChain struct {
	mm *MiddlewareManager
}

// NewAuthChain creates a new authentication chain
func (mm *MiddlewareManager) NewAuthChain() *AuthChain {
	return &AuthChain{mm: mm}
}

// RequireAuth adds authentication requirement
func (ac *AuthChain) RequireAuth() *AuthChain {
	return ac
}

// RequireAdmin adds admin requirement
func (ac *AuthChain) RequireAdmin() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		ac.mm.AuthMiddleware(),
		ac.mm.AdminRequiredMiddleware(),
	}
}

// RequireRole adds specific role requirement
func (ac *AuthChain) RequireRole(roles ...string) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		ac.mm.AuthMiddleware(),
		ac.mm.RoleRequiredMiddleware(roles...),
	}
}

// RequireSelfOrAdmin adds self or admin requirement
func (ac *AuthChain) RequireSelfOrAdmin(userIDParam string) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		ac.mm.AuthMiddleware(),
		ac.mm.SelfOrAdminMiddleware(userIDParam),
	}
}

// RequireAuthOnly adds only authentication requirement
func (ac *AuthChain) RequireAuthOnly() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		ac.mm.AuthMiddleware(),
	}
}

// OptionalAuth adds optional authentication
func (ac *AuthChain) OptionalAuth() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		ac.mm.OptionalAuthMiddleware(),
	}
}

// PublicEndpoint returns empty middleware chain for public endpoints
func (ac *AuthChain) PublicEndpoint() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}

// StrictSecurity adds extra security for sensitive endpoints
func (ac *AuthChain) StrictSecurity() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		ac.mm.StrictRateLimitMiddleware(),
		ac.mm.AuthMiddleware(),
		ac.mm.AdminRequiredMiddleware(),
	}
}

// MiddlewareConfig holds configuration for different endpoint types
type MiddlewareConfig struct {
	Public        []gin.HandlerFunc
	Authenticated []gin.HandlerFunc
	Admin         []gin.HandlerFunc
	SelfOrAdmin   []gin.HandlerFunc
	Strict        []gin.HandlerFunc
}

// GetMiddlewareConfig returns pre-configured middleware sets
func (mm *MiddlewareManager) GetMiddlewareConfig() MiddlewareConfig {
	authChain := mm.NewAuthChain()

	return MiddlewareConfig{
		Public:        authChain.PublicEndpoint(),
		Authenticated: authChain.RequireAuthOnly(),
		Admin:         authChain.RequireAdmin(),
		SelfOrAdmin:   authChain.RequireSelfOrAdmin("id"),
		Strict:        authChain.StrictSecurity(),
	}
}

// SetupRouteGroups creates pre-configured route groups with appropriate middleware
func (mm *MiddlewareManager) SetupRouteGroups(router *gin.Engine) *RouteGroups {
	config := mm.GetMiddlewareConfig()

	return &RouteGroups{
		Public:        router.Group("/api"),
		Authenticated: router.Group("/api", config.Authenticated...),
		Admin:         router.Group("/api/admin", config.Admin...),
		SelfOrAdmin:   router.Group("/api/users", config.SelfOrAdmin...),
	}
}

// RouteGroups holds pre-configured route groups
type RouteGroups struct {
	Public        *gin.RouterGroup
	Authenticated *gin.RouterGroup
	Admin         *gin.RouterGroup
	SelfOrAdmin   *gin.RouterGroup
}

// Close cleans up middleware resources
func (mm *MiddlewareManager) Close() {
	// Currently no cleanup needed for middleware
	// This is a placeholder for future cleanup requirements
} 