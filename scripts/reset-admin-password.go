package main

import (
	"context"
	"fmt"
	"log"

	"user_mgmt_go/internal/config"
	"user_mgmt_go/internal/repository"
	"user_mgmt_go/internal/utils"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize repository manager
	repoManager, err := repository.NewRepositoryManager(&cfg)
	if err != nil {
		log.Fatalf("Failed to initialize repository manager: %v", err)
	}
	defer repoManager.Close()

	ctx := context.Background()

	// Get admin user
	adminUser, err := repoManager.Repos.User.GetByEmail(ctx, "admin@example.com")
	if err != nil {
		log.Fatalf("Failed to get admin user: %v", err)
	}

	// Hash new password
	newPassword := "admin123"
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Update password
	updateData := map[string]interface{}{
		"password": hashedPassword,
	}
	err = repoManager.Repos.User.Update(ctx, adminUser.ID, updateData)
	if err != nil {
		log.Fatalf("Failed to update admin password: %v", err)
	}

	fmt.Printf("✅ Admin password reset successfully!\n")
	fmt.Printf("   Email: admin@example.com\n")
	fmt.Printf("   New Password: %s\n", newPassword)
	fmt.Printf("   User ID: %s\n", adminUser.ID)
	fmt.Printf("\n🔗 You can now login at: http://localhost:8080/admin/login\n")
} 