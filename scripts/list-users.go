package main

import (
	"context"
	"fmt"
	"log"

	"user_mgmt_go/internal/config"
	"user_mgmt_go/internal/repository"
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

	// List users
	ctx := context.Background()
	users, err := repoManager.Repos.User.List(ctx, repository.ListParams{
		Page:     1,
		PageSize: 10,
		SortBy:   "created_at",
		SortDir:  "desc",
	})
	if err != nil {
		log.Fatalf("Failed to list users: %v", err)
	}

	fmt.Printf("📋 Total users: %d\n\n", users.Total)
	
	for _, user := range users.Users {
		fmt.Printf("👤 User:\n")
		fmt.Printf("   ID: %s\n", user.ID)
		fmt.Printf("   Name: %s\n", user.Name)
		fmt.Printf("   Email: %s\n", user.Email)
		fmt.Printf("   Created: %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("   Updated: %s\n\n", user.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
} 