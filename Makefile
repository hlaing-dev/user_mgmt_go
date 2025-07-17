# User Management System - Development Makefile

.PHONY: help build run test clean docker-up docker-down docker-logs deps lint fmt vet check install-tools setup dev migrate

# Default target
.DEFAULT_GOAL := help

# Variables
APP_NAME := user_mgmt_go
BINARY_NAME := user_mgmt_server
BUILD_DIR := ./build
DOCKER_COMPOSE := docker-compose
GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

## help: Show this help message
help:
	@echo "$(BLUE)User Management System - Available Commands:$(NC)"
	@echo ""
	@echo "$(GREEN)Development:$(NC)"
	@echo "  setup          - Initial project setup (install tools, deps)"
	@echo "  dev            - Start development server with hot reload"
	@echo "  build          - Build the application binary"
	@echo "  run            - Build and run the application"
	@echo "  test           - Run tests with coverage"
	@echo "  lint           - Run golangci-lint"
	@echo "  fmt            - Format Go code"
	@echo "  vet            - Run go vet"
	@echo "  check          - Run all checks (fmt, vet, lint, test)"
	@echo ""
	@echo "$(GREEN)Docker:$(NC)"
	@echo "  docker-up      - Start databases with Docker Compose"
	@echo "  docker-down    - Stop Docker Compose services"
	@echo "  docker-logs    - Show Docker Compose logs"
	@echo "  docker-reset   - Reset Docker volumes and restart"
	@echo ""
	@echo "$(GREEN)Database:$(NC)"
	@echo "  migrate        - Run database migrations"
	@echo "  db-status      - Check database connection status"
	@echo ""
	@echo "$(GREEN)Utilities:$(NC)"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  install-tools  - Install development tools"
	@echo "  swagger        - Generate Swagger API documentation"

## setup: Initial project setup
setup: install-tools deps
	@echo "$(GREEN)✅ Project setup complete!$(NC)"
	@echo "$(YELLOW)Next steps:$(NC)"
	@echo "  1. Run 'make docker-up' to start databases"
	@echo "  2. Run 'make dev' to start the development server"

## install-tools: Install development tools
install-tools:
	@echo "$(BLUE)📦 Installing development tools...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/cosmtrek/air@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)✅ Development tools installed$(NC)"

## deps: Download and tidy dependencies
deps:
	@echo "$(BLUE)📦 Downloading dependencies...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✅ Dependencies updated$(NC)"

## dev: Start development server with hot reload
dev:
	@echo "$(BLUE)🚀 Starting development server with hot reload...$(NC)"
	@echo "$(YELLOW)Make sure databases are running: make docker-up$(NC)"
	@air -c .air.toml

## build: Build the application
build:
	@echo "$(BLUE)🔨 Building application...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/server/main.go
	@echo "$(GREEN)✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## run: Build and run the application
run: build
	@echo "$(BLUE)🚀 Running application...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME)

## test: Run tests with coverage
test:
	@echo "$(BLUE)🧪 Running tests...$(NC)"
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Tests complete. Coverage report: coverage.html$(NC)"

## lint: Run golangci-lint
lint:
	@echo "$(BLUE)🔍 Running linter...$(NC)"
	@golangci-lint run ./...
	@echo "$(GREEN)✅ Linting complete$(NC)"

## fmt: Format Go code
fmt:
	@echo "$(BLUE)📝 Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✅ Code formatted$(NC)"

## vet: Run go vet
vet:
	@echo "$(BLUE)🔍 Running go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✅ Vet complete$(NC)"

## check: Run all checks
check: fmt vet lint test
	@echo "$(GREEN)✅ All checks passed!$(NC)"

## docker-up: Start databases with Docker Compose
docker-up:
	@echo "$(BLUE)🐳 Starting databases...$(NC)"
	@$(DOCKER_COMPOSE) up -d postgres mongodb
	@echo "$(GREEN)✅ Databases started$(NC)"
	@echo "$(YELLOW)Waiting for databases to be ready...$(NC)"
	@sleep 10
	@make db-status

## docker-down: Stop Docker Compose services
docker-down:
	@echo "$(BLUE)🐳 Stopping Docker services...$(NC)"
	@$(DOCKER_COMPOSE) down
	@echo "$(GREEN)✅ Docker services stopped$(NC)"

## docker-logs: Show Docker Compose logs
docker-logs:
	@echo "$(BLUE)📋 Docker Compose logs:$(NC)"
	@$(DOCKER_COMPOSE) logs -f

## docker-reset: Reset Docker volumes and restart
docker-reset:
	@echo "$(BLUE)🔄 Resetting Docker environment...$(NC)"
	@$(DOCKER_COMPOSE) down -v
	@$(DOCKER_COMPOSE) up -d postgres mongodb
	@echo "$(GREEN)✅ Docker environment reset$(NC)"

## migrate: Run database migrations
migrate:
	@echo "$(BLUE)🗃️  Running database migrations...$(NC)"
	@go run cmd/server/main.go -migrate
	@echo "$(GREEN)✅ Migrations complete$(NC)"

## db-status: Check database connection status
db-status:
	@echo "$(BLUE)🔍 Checking database status...$(NC)"
	@echo "PostgreSQL:"
	@docker exec user_mgmt_postgres pg_isready -U postgres || echo "$(RED)❌ PostgreSQL not ready$(NC)"
	@echo "MongoDB:"
	@docker exec user_mgmt_mongodb mongosh --eval "db.adminCommand('ping')" --quiet || echo "$(RED)❌ MongoDB not ready$(NC)"

## clean: Clean build artifacts
clean:
	@echo "$(BLUE)🧹 Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@go clean
	@echo "$(GREEN)✅ Clean complete$(NC)"

# Development watch configuration
.air.toml:
	@echo "$(BLUE)📝 Creating Air configuration...$(NC)"
	@echo 'root = "."' > .air.toml
	@echo 'testdata_dir = "testdata"' >> .air.toml
	@echo 'tmp_dir = "tmp"' >> .air.toml
	@echo '' >> .air.toml
	@echo '[build]' >> .air.toml
	@echo '  args_bin = []' >> .air.toml
	@echo '  bin = "./tmp/main"' >> .air.toml
	@echo '  cmd = "go build -o ./tmp/main cmd/server/main.go"' >> .air.toml
	@echo '  delay = 1000' >> .air.toml
	@echo '  exclude_dir = ["assets", "tmp", "vendor", "testdata", "build"]' >> .air.toml
	@echo '  exclude_file = []' >> .air.toml
	@echo '  exclude_regex = ["_test.go"]' >> .air.toml
	@echo '  exclude_unchanged = false' >> .air.toml
	@echo '  follow_symlink = false' >> .air.toml
	@echo '  full_bin = ""' >> .air.toml
	@echo '  include_dir = []' >> .air.toml
	@echo '  include_ext = ["go", "tpl", "tmpl", "html"]' >> .air.toml
	@echo '  kill_delay = "0s"' >> .air.toml
	@echo '  log = "build-errors.log"' >> .air.toml
	@echo '  send_interrupt = false' >> .air.toml
	@echo '  stop_on_root = false' >> .air.toml
	@echo '' >> .air.toml
	@echo '[color]' >> .air.toml
	@echo '  app = ""' >> .air.toml
	@echo '  build = "yellow"' >> .air.toml
	@echo '  main = "magenta"' >> .air.toml
	@echo '  runner = "green"' >> .air.toml
	@echo '  watcher = "cyan"' >> .air.toml
	@echo '' >> .air.toml
	@echo '[log]' >> .air.toml
	@echo '  time = false' >> .air.toml
	@echo '' >> .air.toml
	@echo '[misc]' >> .air.toml
	@echo '  clean_on_exit = false' >> .air.toml
	@echo '' >> .air.toml
	@echo '[screen]' >> .air.toml
	@echo '  clear_on_rebuild = false' >> .air.toml
	@echo "$(GREEN)✅ Air configuration created$(NC)"

# Create .air.toml if it doesn't exist
air-config: .air.toml

## swagger: Generate Swagger documentation
swagger:
	@echo "$(BLUE)📚 Generating Swagger documentation...$(NC)"
	@~/go/bin/swag init -g cmd/server/main.go -o docs/ || go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go -o docs/
	@echo "$(GREEN)✅ Swagger docs generated in docs/$(NC)"

# Quick development start
quick-start: docker-up air-config
	@echo "$(GREEN)🚀 Quick start complete!$(NC)"
	@echo "$(YELLOW)Run 'make dev' to start the development server$(NC)" 