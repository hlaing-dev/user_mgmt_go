package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"user_mgmt_go/internal/config"
	"user_mgmt_go/internal/models"
	"user_mgmt_go/internal/utils"
)

// RepositoryManager manages all repositories and database connections
type RepositoryManager struct {
	Database *Database
	Repos    *Repository
	config   *config.Config
}

// NewRepositoryManager creates a new repository manager with all dependencies
func NewRepositoryManager(cfg *config.Config) (*RepositoryManager, error) {
	// Initialize database connections
	database, err := NewDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize repositories
	userRepo := NewUserRepository(database.PostgreSQL)
	logRepo := NewUserLogRepository(database.MongoDB)

	repos := &Repository{
		User: userRepo,
		Log:  logRepo,
	}

	manager := &RepositoryManager{
		Database: database,
		Repos:    repos,
		config:   cfg,
	}

	// Create default admin user if it doesn't exist
	if err := manager.createDefaultAdmin(); err != nil {
		log.Printf("Warning: Failed to create default admin user: %v", err)
	}

	log.Println("✅ Repository manager initialized successfully")
	return manager, nil
}

// createDefaultAdmin creates the default admin user from config
func (rm *RepositoryManager) createDefaultAdmin() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if admin user already exists
	exists, err := rm.Repos.User.Exists(ctx, rm.config.Admin.Email)
	if err != nil {
		return fmt.Errorf("failed to check admin user existence: %w", err)
	}

	if exists {
		log.Printf("Admin user %s already exists, skipping creation", rm.config.Admin.Email)
		return nil
	}

	// Hash the admin password
	hashedPassword, err := utils.HashPassword(rm.config.Admin.Password)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	// Create admin user
	adminUser := &models.User{
		Name:     "Administrator",
		Email:    rm.config.Admin.Email,
		Password: hashedPassword,
	}

	if err := rm.Repos.User.Create(ctx, adminUser); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Log the admin user creation
	logEntry := models.NewUserLog(models.UserLogCreateRequest{
		UserID: &adminUser.ID,
		Event:  models.UserCreated,
		Action: "CREATE_ADMIN_USER",
		Details: map[string]interface{}{
			"email": adminUser.Email,
			"name":  adminUser.Name,
			"role":  "admin",
		},
	})

	if err := rm.Repos.Log.CreateAsync(logEntry); err != nil {
		log.Printf("Failed to log admin user creation: %v", err)
	}

	log.Printf("✅ Default admin user created: %s", rm.config.Admin.Email)
	return nil
}

// Close gracefully shuts down all repository connections
func (rm *RepositoryManager) Close() error {
	log.Println("🔄 Shutting down repository manager...")

	// Close log repository (stops async processor)
	if logRepo, ok := rm.Repos.Log.(*userLogRepository); ok {
		logRepo.Close()
	}

	// Close database connections
	if err := rm.Database.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
		return err
	}

	log.Println("✅ Repository manager shut down successfully")
	return nil
}

// HealthCheck performs a health check on all database connections
func (rm *RepositoryManager) HealthCheck() map[string]bool {
	pgHealthy, mongoHealthy := rm.Database.HealthCheck()
	
	return map[string]bool{
		"postgresql": pgHealthy,
		"mongodb":    mongoHealthy,
	}
}

// GetStats returns statistics about the repositories
func (rm *RepositoryManager) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get user count
	userCount, err := rm.Repos.User.Count(ctx, UserFilter{})
	if err != nil {
		log.Printf("Failed to get user count: %v", err)
		userCount = -1
	}
	stats["total_users"] = userCount

	// Get log count for last 30 days
	logCount, err := rm.Repos.Log.Count(ctx, models.LogFilterRequest{
		StartDate: &[]time.Time{time.Now().AddDate(0, 0, -30)}[0],
	})
	if err != nil {
		log.Printf("Failed to get log count: %v", err)
		logCount = -1
	}
	stats["logs_last_30_days"] = logCount

	// Get event statistics for last 7 days
	eventStats, err := rm.Repos.Log.GetEventStats(ctx, nil, 7)
	if err != nil {
		log.Printf("Failed to get event stats: %v", err)
		eventStats = make(map[models.LogEventType]int64)
	}
	stats["event_stats_last_7_days"] = eventStats

	return stats, nil
}

// RunMaintenance performs routine maintenance tasks
func (rm *RepositoryManager) RunMaintenance(ctx context.Context) error {
	log.Println("🔄 Running repository maintenance...")

	// Delete old logs (older than 90 days)
	deletedCount, err := rm.Repos.Log.DeleteOldLogs(ctx, 90)
	if err != nil {
		log.Printf("Failed to delete old logs: %v", err)
	} else {
		log.Printf("Deleted %d old log entries", deletedCount)
	}

	// Log maintenance completion
	logEntry := models.NewUserLog(models.UserLogCreateRequest{
		Event:  models.SystemError, // Using SystemError as maintenance event
		Action: "SYSTEM_MAINTENANCE",
		Details: map[string]interface{}{
			"deleted_logs": deletedCount,
			"timestamp":    time.Now(),
		},
	})

	if err := rm.Repos.Log.CreateAsync(logEntry); err != nil {
		log.Printf("Failed to log maintenance: %v", err)
	}

	log.Println("✅ Repository maintenance completed")
	return nil
}

// SeedTestData creates test data for development/testing (only in debug mode)
func (rm *RepositoryManager) SeedTestData(ctx context.Context) error {
	if rm.config.Server.GinMode != "debug" {
		return fmt.Errorf("test data seeding only allowed in debug mode")
	}

	log.Println("🌱 Seeding test data...")

	// Create test users
	testUsers := []*models.User{
		{
			Name:     "John Doe",
			Email:    "john.doe@example.com",
			Password: "hashed_password_here", // This should be properly hashed
		},
		{
			Name:     "Jane Smith",
			Email:    "jane.smith@example.com", 
			Password: "hashed_password_here",
		},
		{
			Name:     "Bob Johnson",
			Email:    "bob.johnson@example.com",
			Password: "hashed_password_here",
		},
	}

	// Hash passwords
	for _, user := range testUsers {
		hashedPassword, err := utils.HashPassword("testpassword123")
		if err != nil {
			return fmt.Errorf("failed to hash test password: %w", err)
		}
		user.Password = hashedPassword
	}

	// Create users in batch
	if err := rm.Repos.User.CreateBatch(ctx, testUsers); err != nil {
		return fmt.Errorf("failed to create test users: %w", err)
	}

	// Create test logs for each user
	for _, user := range testUsers {
		logEntry := models.NewUserLog(models.UserLogCreateRequest{
			UserID: &user.ID,
			Event:  models.UserCreated,
			Action: "CREATE_TEST_USER",
			Details: map[string]interface{}{
				"email": user.Email,
				"name":  user.Name,
				"test":  true,
			},
		})

		if err := rm.Repos.Log.CreateAsync(logEntry); err != nil {
			log.Printf("Failed to log test user creation: %v", err)
		}
	}

	log.Printf("✅ Created %d test users with logs", len(testUsers))
	return nil
}

// Backup creates a backup of critical data (placeholder for actual implementation)
func (rm *RepositoryManager) Backup(ctx context.Context, backupPath string) error {
	// This is a placeholder for backup functionality
	// In a real implementation, you would:
	// 1. Create PostgreSQL dump
	// 2. Export MongoDB collections
	// 3. Compress and store backups
	
	log.Printf("📦 Backup functionality not implemented yet. Would backup to: %s", backupPath)
	return nil
}

// Import imports data from backup (placeholder for actual implementation)
func (rm *RepositoryManager) Import(ctx context.Context, backupPath string) error {
	// This is a placeholder for import functionality
	// In a real implementation, you would:
	// 1. Restore PostgreSQL from dump
	// 2. Import MongoDB collections
	// 3. Verify data integrity
	
	log.Printf("📥 Import functionality not implemented yet. Would import from: %s", backupPath)
	return nil
} 