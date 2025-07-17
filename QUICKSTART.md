# 🚀 Quick Start Guide - User Management System

This guide will help you get the User Management System up and running quickly on any machine.

## 📋 Prerequisites

Before you start, make sure you have:

- **Go 1.21+** - [Download here](https://golang.org/dl/)
- **Docker & Docker Compose** - [Download here](https://docs.docker.com/get-docker/)
- **Git** - [Download here](https://git-scm.com/downloads)

### Verify Prerequisites
```bash
go version        # Should show Go 1.21+
docker --version  # Should show Docker 20.10+
git --version     # Should show Git 2.0+
```

## 🏃‍♂️ First Time Setup (New Machine)

### Option 1: Use the Setup Script (Recommended)
```bash
# Clone the repository
git clone <your-repo-url>
cd user_mgmt_go

# Run the setup script
chmod +x setup.sh
./setup.sh
```

### Option 2: Manual Setup
```bash
# 1. Clone the repository
git clone <your-repo-url>
cd user_mgmt_go

# 2. Start databases
make docker-up

# 3. Install Go dependencies
make deps

# 4. Build the application
make build

# 5. Start the server
make run
# OR for different port:
# PORT=8081 make run
```

## 🔄 Daily Development Workflow

### Start Development Session
```bash
# 1. Make sure you're in the project directory
cd user_mgmt_go

# 2. Start databases (if not running)
make docker-up

# 3. Start the server
make run
```

### Quick Commands Reference
```bash
# Database Management
make docker-up          # Start all databases (PostgreSQL, MongoDB, Redis)
make docker-down        # Stop all databases
make docker-restart     # Restart all databases
make docker-logs        # View database logs

# Development
make deps               # Install/update Go dependencies
make build              # Build the application
make run                # Start the server (port 8080)
PORT=8081 make run      # Start on different port
make dev                # Start with hot reload (if air is installed)

# Testing & Quality
make test               # Run all tests
make test-coverage      # Run tests with coverage report
make lint               # Run linter
make fmt                # Format code

# Documentation
make swagger            # Generate Swagger documentation
make docs               # Open documentation in browser

# Cleanup
make clean              # Clean build artifacts
make docker-clean       # Clean Docker volumes (CAUTION: removes data)
```

## 🌐 Access Points

Once the server is running, you can access:

| Service | URL | Description |
|---------|-----|-------------|
| **API Health** | http://localhost:8080/health | Quick health check |
| **Swagger Docs** | http://localhost:8080/swagger/index.html | Interactive API documentation |
| **API Routes** | http://localhost:8080/api/docs/routes | List of all available routes |

## 🔐 Default Credentials

The system creates a default admin user:
- **Email**: `admin@example.com`
- **Password**: `admin123`

### First Login Test
```bash
# Test login with curl
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}'
```

## 🛠️ Common Development Tasks

### Creating a New User
```bash
# First, login to get JWT token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}' | \
  jq -r '.token')

# Create a new user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"John Doe","email":"john@example.com","password":"password123"}'
```

### Viewing Logs
```bash
# View application logs
docker logs user_mgmt_go_app

# View database logs
make docker-logs

# View MongoDB logs specifically
docker logs user_mgmt_mongodb
```

## 🐛 Troubleshooting

### Port Already in Use
```bash
# If port 8080 is busy, use a different port
PORT=8081 make run

# Or find what's using the port
lsof -i :8080
# Kill the process if needed
sudo kill -9 <PID>
```

### Database Connection Issues
```bash
# Check if databases are running
docker ps

# Restart databases
make docker-restart

# Check database health
curl http://localhost:8080/health
```

### MongoDB Authentication Errors
```bash
# Reset MongoDB data
make docker-down
docker volume rm user_mgmt_go_mongodb_data
make docker-up
```

### Missing Dependencies
```bash
# Reinstall Go dependencies
go mod tidy
make deps

# Clean module cache if needed
go clean -modcache
make deps
```

### Configuration Issues
```bash
# Check if config.yaml exists
ls -la config.yaml

# Copy from sample if missing
cp config.sample.yaml config.yaml
```

### Environment Configuration (.env)

The application supports **dual configuration**:
- **`config.yaml`**: Base configuration
- **`.env`**: Environment-specific overrides ✅ (Already created)

```bash
# Check .env file (should exist)
ls -la .env

# If missing, copy from example:
cp .env.example .env

# Edit .env for your environment:
nano .env
```

**Priority Order (highest to lowest):**
1. OS Environment Variables
2. `.env` file
3. `config.yaml` defaults

**Production Security:**
```bash
# Always change these in production:
JWT_SECRET=your-256-bit-secret
ADMIN_PASSWORD=secure-password
DB_PASSWORD=secure-db-password
```

## 🧪 Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Open coverage report in browser
open coverage.html
```

## 📊 Monitoring & Health Checks

### Quick Health Check
```bash
curl http://localhost:8080/health
```

### Detailed System Status
```bash
# Get system statistics (requires admin token)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/admin/stats
```

### Database Status
```bash
# Check PostgreSQL
docker exec user_mgmt_postgres pg_isready -U postgres

# Check MongoDB
docker exec user_mgmt_mongodb mongosh --eval "db.adminCommand('ping')"

# Check Redis
docker exec user_mgmt_redis redis-cli ping
```

## 🔧 Environment Variables

You can customize the application using environment variables:

```bash
# Server Configuration
export PORT=8080
export HOST=localhost
export GIN_MODE=debug

# Database Configuration
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password123
export DB_NAME=user_mgmt

# MongoDB Configuration
export MONGO_URI="mongodb://admin:password123@localhost:27017"
export MONGO_DATABASE=user_logs

# JWT Configuration
export JWT_SECRET=your-super-secret-key
export JWT_EXPIRY=24h

# Admin Configuration
export ADMIN_EMAIL=admin@example.com
export ADMIN_PASSWORD=admin123
```

## 📝 Notes

- The application uses PostgreSQL for user data and MongoDB for activity logs
- All databases run in Docker containers for easy development
- The application automatically creates database schemas and indexes
- Swagger documentation is automatically generated and updated
- All passwords are securely hashed using bcrypt
- JWT tokens are used for authentication

## 🆘 Getting Help

If you encounter issues:

1. Check the troubleshooting section above
2. Review the logs: `docker logs user_mgmt_go_app`
3. Ensure all prerequisites are installed
4. Try the setup script: `./setup.sh`
5. Check the detailed documentation in `DEVELOPMENT.md`

## 🎯 Next Steps

After successful setup:

1. Explore the Swagger documentation at `/swagger/index.html`
2. Read the API documentation for endpoint details
3. Check out `TESTING.md` for comprehensive testing guide
4. Review `DEVELOPMENT.md` for advanced development topics

Happy coding! 🚀 