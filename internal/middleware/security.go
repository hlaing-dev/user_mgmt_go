package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"user_mgmt_go/internal/models"
	"user_mgmt_go/internal/repository"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	visitors map[string]*Visitor
	mutex    sync.RWMutex
	rate     time.Duration
	burst    int
}

// Visitor holds rate limiting information for each visitor
type Visitor struct {
	limiter   chan struct{}
	lastSeen  time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate time.Duration, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     rate,
		burst:    burst,
	}

	// Start cleanup goroutine
	go rl.cleanupVisitors()

	return rl
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	visitor, exists := rl.visitors[ip]
	if !exists {
		visitor = &Visitor{
			limiter:  make(chan struct{}, rl.burst),
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = visitor
	}

	visitor.lastSeen = time.Now()

	select {
	case visitor.limiter <- struct{}{}:
		// Release token after rate duration
		go func() {
			time.Sleep(rl.rate)
			<-visitor.limiter
		}()
		return true
	default:
		return false
	}
}

// cleanupVisitors removes old visitors to prevent memory leaks
func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		rl.mutex.Lock()
		for ip, visitor := range rl.visitors {
			if time.Since(visitor.lastSeen) > time.Hour {
				delete(rl.visitors, ip)
			}
		}
		rl.mutex.Unlock()
	}
}

// RateLimitMiddleware creates rate limiting middleware
func RateLimitMiddleware(rateLimiter *RateLimiter) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		ip := c.ClientIP()
		
		if !rateLimiter.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, models.NewErrorResponse(
				http.StatusTooManyRequests,
				"Rate Limit Exceeded",
				"Too many requests from this IP address",
				map[string]interface{}{
					"retry_after": "60 seconds",
					"ip":          ip,
				},
			))
			c.Abort()
			return
		}

		c.Next()
	})
}

// RequestLoggingMiddleware logs incoming requests and responses
func RequestLoggingMiddleware(logRepo repository.UserLogRepository) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Get user ID from context if authenticated (for potential future use)
		if userIDRaw, exists := c.Get("user_id"); exists {
			_ = userIDRaw // Placeholder for future logging enhancement
		}

		// Create log entry asynchronously
		logEntry := &models.UserLog{
			Event:     models.SystemError, // Using as general system event
			Data: models.LogData{
				Action: "HTTP_REQUEST",
				Details: map[string]interface{}{
					"method":      method,
					"path":        path,
					"status_code": statusCode,
					"duration_ms": duration.Milliseconds(),
					"user_agent":  userAgent,
					"ip_address":  clientIP,
				},
				StatusCode: statusCode,
				Duration:   duration.Milliseconds(),
			},
			Timestamp: time.Now(),
			IPAddress: clientIP,
			UserAgent: userAgent,
		}

		// Log the request (non-blocking)
		if err := logRepo.CreateAsync(logEntry); err != nil {
			log.Printf("Failed to log request: %v", err)
		}
	})
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		// Prevent clickjacking attacks
		c.Header("X-Frame-Options", "DENY")
		
		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Force HTTPS in production
		if gin.Mode() == gin.ReleaseMode {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		// Prevent information disclosure
		c.Header("X-Powered-By", "")
		c.Header("Server", "")
		
		// Content Security Policy - relaxed for Swagger UI and Admin Panel
		path := c.Request.URL.Path
		if path == "/swagger/index.html" || 
		   path == "/swagger/" || 
		   strings.HasPrefix(path, "/swagger/") {
			// Relaxed CSP for Swagger UI
			c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; connect-src 'self'")
		} else if strings.HasPrefix(path, "/admin/") {
			// Relaxed CSP for Admin Panel - allow CDN resources
			c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; img-src 'self' data:; font-src 'self' data: https://cdn.jsdelivr.net; connect-src 'self'")
		} else {
			// Strict CSP for other routes
			c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'")
		}
		
		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	})
}

// CORSMiddleware creates CORS middleware with custom configuration
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	config := cors.DefaultConfig()
	
	if len(allowedOrigins) > 0 {
		config.AllowOrigins = allowedOrigins
	} else {
		config.AllowAllOrigins = false
		config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:3001"}
	}
	
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	return cors.New(config)
}

// RequestSizeLimitMiddleware limits the size of request bodies
func RequestSizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, models.NewErrorResponse(
				http.StatusRequestEntityTooLarge,
				"Request Too Large",
				"Request body too large",
				map[string]interface{}{
					"max_size_bytes": maxSize,
					"received_bytes": c.Request.ContentLength,
				},
			))
			c.Abort()
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	})
}

// TimeoutMiddleware adds request timeout
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		
		// Set the new context with timeout
		c.Request = c.Request.WithContext(ctx)
		
		// Set timeout
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		done := make(chan struct{})
		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			// Request completed within timeout
			return
		case <-timer.C:
			// Request timed out
			c.JSON(http.StatusRequestTimeout, models.NewErrorResponse(
				http.StatusRequestTimeout,
				"Request Timeout",
				"Request processing timeout",
				map[string]interface{}{
					"timeout_seconds": timeout.Seconds(),
				},
			))
			c.Abort()
			return
		}
	})
}

// RecoveryMiddleware provides panic recovery with logging
func RecoveryMiddleware(logRepo repository.UserLogRepository) gin.HandlerFunc {
	return gin.RecoveryWithWriter(gin.DefaultWriter, func(c *gin.Context, err interface{}) {
		// Log the panic
		logEntry := &models.UserLog{
			Event: models.SystemError,
			Data: models.LogData{
				Action: "PANIC_RECOVERY",
				Error:  fmt.Sprintf("Panic: %v", err),
				Details: map[string]interface{}{
					"method":     c.Request.Method,
					"path":       c.Request.URL.Path,
					"ip_address": c.ClientIP(),
					"user_agent": c.Request.UserAgent(),
				},
				StatusCode: http.StatusInternalServerError,
			},
			Timestamp: time.Now(),
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		}

		if err := logRepo.CreateAsync(logEntry); err != nil {
			log.Printf("Failed to log panic: %v", err)
		}

		// Return error response
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"Internal Server Error",
			"An unexpected error occurred",
			nil,
		))
	})
}

// HealthCheckMiddleware provides a simple health check endpoint
func HealthCheckMiddleware(repoManager *repository.RepositoryManager) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if c.Request.URL.Path == "/health" && c.Request.Method == "GET" {
			// Check database health
			health := repoManager.HealthCheck()
			
			status := "healthy"
			httpStatus := http.StatusOK
			
			// If any database is unhealthy, mark as unhealthy
			for _, isHealthy := range health {
				if !isHealthy {
					status = "unhealthy"
					httpStatus = http.StatusServiceUnavailable
					break
				}
			}

			response := models.HealthCheckResponse{
				Status:    status,
				Timestamp: time.Now(),
				Version:   "1.0.0",
			}
			response.Services.Database = health["postgresql"]
			response.Services.MongoDB = health["mongodb"]

			c.JSON(httpStatus, response)
			c.Abort()
			return
		}

		c.Next()
	})
}

// IPWhitelistMiddleware allows only whitelisted IP addresses
func IPWhitelistMiddleware(whitelist []string) gin.HandlerFunc {
	allowedIPs := make(map[string]bool)
	for _, ip := range whitelist {
		allowedIPs[ip] = true
	}

	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		if len(allowedIPs) > 0 && !allowedIPs[clientIP] {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(
				http.StatusForbidden,
				"IP Not Allowed",
				"Your IP address is not allowed to access this resource",
				map[string]string{"ip": clientIP},
			))
			c.Abort()
			return
		}

		c.Next()
	})
} 