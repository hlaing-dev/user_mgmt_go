package main

import (
	"context"
	"fmt"
	"log"

	"user_mgmt_go/internal/config"
	"user_mgmt_go/internal/models"
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

	// Check if admin user already exists
	ctx := context.Background()
	exists, err := repoManager.Repos.User.Exists(ctx, "admin@example.com")
	if err != nil {
		log.Fatalf("Failed to check if admin user exists: %v", err)
	}

	if exists {
		fmt.Println("Admin user already exists: admin@example.com")
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword("admin123")
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create admin user
	adminUser := &models.User{
		Name:     "System Administrator",
		Email:    "admin@example.com",
		Password: hashedPassword,
	}

	err = repoManager.Repos.User.Create(ctx, adminUser)
	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Printf("✅ Admin user created successfully!\n")
	fmt.Printf("   Email: admin@example.com\n")
	fmt.Printf("   Password: admin123\n")
	fmt.Printf("   User ID: %s\n", adminUser.ID)
	fmt.Printf("\n🔗 You can now login at: http://localhost:8080/admin/login\n")
} 