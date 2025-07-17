#!/bin/bash

# User Management System - Setup Script
# This script automates the first-time setup process

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Emojis for better UX
CHECK="✅"
CROSS="❌"
ROCKET="🚀"
GEAR="⚙️"
DATABASE="🗄️"
PACKAGE="📦"
TEST="🧪"
INFO="ℹ️"
WARNING="⚠️"

print_header() {
    echo -e "${PURPLE}"
    echo "╔══════════════════════════════════════════════════════════════════╗"
    echo "║                 User Management System Setup                     ║"
    echo "║                        🚀 Quick Start 🚀                        ║"
    echo "╚══════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

print_step() {
    echo -e "${CYAN}${GEAR} $1${NC}"
}

print_success() {
    echo -e "${GREEN}${CHECK} $1${NC}"
}

print_error() {
    echo -e "${RED}${CROSS} $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}${WARNING} $1${NC}"
}

print_info() {
    echo -e "${BLUE}${INFO} $1${NC}"
}

check_command() {
    if command -v "$1" >/dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

check_prerequisites() {
    print_step "Checking prerequisites..."
    
    local missing=()
    
    # Check Go
    if check_command go; then
        GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
        print_success "Go found: $GO_VERSION"
    else
        missing+=("Go 1.21+")
    fi
    
    # Check Docker
    if check_command docker; then
        DOCKER_VERSION=$(docker --version | cut -d' ' -f3 | sed 's/,//')
        print_success "Docker found: $DOCKER_VERSION"
    else
        missing+=("Docker")
    fi
    
    # Check Docker Compose
    if check_command docker-compose || docker compose version >/dev/null 2>&1; then
        print_success "Docker Compose found"
    else
        missing+=("Docker Compose")
    fi
    
    # Check Git
    if check_command git; then
        GIT_VERSION=$(git --version | cut -d' ' -f3)
        print_success "Git found: $GIT_VERSION"
    else
        missing+=("Git")
    fi
    
    # Check Make
    if check_command make; then
        print_success "Make found"
    else
        missing+=("Make")
    fi
    
    if [ ${#missing[@]} -ne 0 ]; then
        print_error "Missing prerequisites:"
        for item in "${missing[@]}"; do
            echo -e "  ${RED}${CROSS} $item${NC}"
        done
        echo
        print_info "Please install the missing prerequisites and run this script again."
        print_info "Installation guides:"
        echo "  - Go: https://golang.org/dl/"
        echo "  - Docker: https://docs.docker.com/get-docker/"
        echo "  - Git: https://git-scm.com/downloads"
        exit 1
    fi
    
    print_success "All prerequisites found!"
    echo
}

start_databases() {
    print_step "Starting databases (PostgreSQL, MongoDB, Redis)..."
    
    if make docker-up; then
        print_success "Databases started successfully"
        
        # Wait for databases to be ready
        print_step "Waiting for databases to be ready..."
        sleep 10
        
        # Check if databases are healthy
        local retries=30
        local count=0
        while [ $count -lt $retries ]; do
            if docker exec user_mgmt_postgres pg_isready -U postgres >/dev/null 2>&1 && \
               docker exec user_mgmt_mongodb mongosh --eval "db.adminCommand('ping')" >/dev/null 2>&1; then
                print_success "All databases are healthy and ready"
                break
            fi
            count=$((count + 1))
            echo -n "."
            sleep 2
        done
        
        if [ $count -eq $retries ]; then
            print_warning "Databases may not be fully ready yet, but continuing..."
        fi
    else
        print_error "Failed to start databases"
        exit 1
    fi
    echo
}

install_dependencies() {
    print_step "Installing Go dependencies..."
    
    if make deps; then
        print_success "Dependencies installed successfully"
    else
        print_error "Failed to install dependencies"
        exit 1
    fi
    echo
}

build_application() {
    print_step "Building the application..."
    
    if make build; then
        print_success "Application built successfully"
    else
        print_error "Failed to build application"
        exit 1
    fi
    echo
}

run_tests() {
    print_step "Running tests to verify setup..."
    
    if make test >/dev/null 2>&1; then
        print_success "All tests passed"
    else
        print_warning "Some tests failed, but setup can continue"
        print_info "You can run 'make test' later to see detailed test results"
    fi
    echo
}

generate_docs() {
    print_step "Generating API documentation..."
    
    if make swagger >/dev/null 2>&1; then
        print_success "Swagger documentation generated"
    else
        print_warning "Failed to generate documentation, but setup can continue"
    fi
    echo
}

show_completion_info() {
    echo -e "${GREEN}"
    echo "╔══════════════════════════════════════════════════════════════════╗"
    echo "║                    🎉 Setup Complete! 🎉                        ║"
    echo "╚══════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    print_info "Your User Management System is ready to use!"
    echo
    
    echo -e "${YELLOW}📋 Quick Start Commands:${NC}"
    echo "  make run                 # Start the server (port 8080)"
    echo "  PORT=8081 make run       # Start on different port"
    echo "  make test                # Run tests"
    echo "  make docker-down         # Stop databases"
    echo
    
    echo -e "${YELLOW}🌐 Access Points:${NC}"
    echo "  Health Check:    http://localhost:8080/health"
    echo "  Swagger Docs:    http://localhost:8080/swagger/index.html"
    echo "  API Routes:      http://localhost:8080/api/docs/routes"
    echo
    
    echo -e "${YELLOW}🔐 Default Admin Credentials:${NC}"
    echo "  Email:    admin@example.com"
    echo "  Password: admin123"
    echo
    
    echo -e "${YELLOW}📖 Documentation:${NC}"
    echo "  Quick Start:     cat QUICKSTART.md"
    echo "  Development:     cat DEVELOPMENT.md"
    echo "  Testing:         cat TESTING.md"
    echo
}

start_server_prompt() {
    echo -e "${BLUE}${ROCKET} Would you like to start the server now? (y/n): ${NC}"
    read -r response
    
    if [[ "$response" =~ ^[Yy]$ ]]; then
        print_step "Starting the server..."
        echo
        print_info "Server will start on http://localhost:8080"
        print_info "Press Ctrl+C to stop the server"
        echo
        make run
    else
        print_info "You can start the server later with: make run"
    fi
}

main() {
    print_header
    
    # Check if we're in the right directory
    if [ ! -f "go.mod" ] || [ ! -f "Makefile" ]; then
        print_error "This doesn't seem to be the User Management System directory"
        print_info "Please run this script from the project root directory"
        exit 1
    fi
    
    # Run setup steps
    check_prerequisites
    start_databases
    install_dependencies
    build_application
    run_tests
    generate_docs
    
    # Show completion info
    show_completion_info
    
    # Ask if user wants to start the server
    start_server_prompt
}

# Check if script is being run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi 