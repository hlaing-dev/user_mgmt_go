# User Management System - Development Guide 🚀

## 🎯 What We've Built

A **production-ready User Management System** with:
- **🔐 JWT Authentication** with role-based access control
- **👥 Complete User CRUD** with validation and security
- **📊 Async MongoDB Logging** for all user activities
- **🛡️ Enterprise Security** (rate limiting, CORS, input validation)
- **🏗️ Clean Architecture** with dependency injection
- **🚀 Professional Deployment** setup

## 📁 Project Structure

```
user_mgmt_go/
├── cmd/server/main.go          # 🚀 Application entry point
├── internal/                   # 🏗️ Core application code
│   ├── config/                 # ⚙️ Configuration management
│   ├── handlers/               # 🌐 HTTP handlers (REST API)
│   ├── middleware/             # 🛡️ Security & authentication middleware
│   ├── models/                 # 📋 Data models
│   ├── repository/             # 🗃️ Database access layer
│   └── utils/                  # 🔧 Utilities (JWT, auth, etc.)
├── docker-compose.yml          # 🐳 Development databases
├── Dockerfile                  # 📦 Application containerization
├── Makefile                    # 🛠️ Development automation
├── config.yaml                 # ⚙️ Application configuration
└── scripts/                    # 📝 Database initialization
    ├── init-postgres.sql
    └── init-mongo.js
```

## 🚀 Quick Start Guide

### 1. **Install Dependencies**
```bash
make setup
```
This installs Go development tools and downloads dependencies.

### 2. **Start Databases** (Requires Docker)
```bash
# Start Docker Desktop first, then:
make docker-up
```
This starts PostgreSQL and MongoDB containers.

### 3. **Start Development Server**
```bash
make dev
```
This starts the server with hot reload on `http://localhost:8080`

## 🎮 Available Commands

```bash
# Development
make setup          # Initial project setup
make dev            # Start with hot reload
make build          # Build application
make run            # Build and run
make test           # Run tests with coverage

# Docker
make docker-up      # Start databases
make docker-down    # Stop databases
make docker-reset   # Reset database data

# Code Quality
make fmt            # Format code
make lint           # Run linter
make vet            # Run go vet
make check          # Run all checks

# Utilities
make clean          # Clean build artifacts
make deps           # Update dependencies
```

## 📚 **Swagger API Documentation**

Interactive API documentation available at: **http://localhost:8080/swagger/index.html**

```bash
# Generate/update Swagger docs
make swagger

# View documentation
make dev
# Then open: http://localhost:8080/swagger/index.html
```

## 🧪 **Testing**

Comprehensive test suite with coverage reporting:

```bash
# Run all tests
make test

# View coverage report
open coverage.html
```

**Test Coverage Includes:**
- ✅ Password hashing & verification
- ✅ JWT token management
- ✅ Configuration validation  
- ✅ HTTP request/response handling
- ✅ Model serialization
- ✅ Performance benchmarks

## 🔑 API Endpoints

Once running, the API provides:

### 🔓 **Public Endpoints**
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - User login
- `POST /api/auth/admin/login` - Admin login
- `GET /health` - Health check

### 🔐 **User Endpoints** (Requires JWT)
- `GET /api/users/profile` - Get user profile
- `PUT /api/users/profile` - Update user profile

### 👑 **Admin Endpoints** (Requires Admin JWT)
- `GET /api/admin/users` - List all users
- `POST /api/admin/users` - Create user
- `PUT /api/admin/users/:id` - Update user
- `DELETE /api/admin/users/:id` - Delete user
- `GET /api/admin/users/:id/logs` - Get user activity logs

### 📊 **Documentation & Monitoring**
- `GET /api/docs/routes` - API documentation
- `GET /metrics` - Prometheus metrics

## 🔧 Configuration

Edit `config.yaml` for your environment:

```yaml
# Key settings to customize:
server:
  port: "8080"
  gin_mode: "debug"

database:
  host: "localhost"
  password: "your-secure-password"

jwt:
  secret: "your-super-secret-jwt-key"
  expiry: "24h"

admin:
  email: "admin@yourcompany.com"
  password: "your-admin-password"
```

## 🗃️ Database Setup

### PostgreSQL (User Data)
- **Host:** localhost:5432
- **Database:** user_mgmt
- **User:** postgres
- **Password:** password123

### MongoDB (Activity Logs)
- **Host:** localhost:27017
- **Database:** user_logs
- **Auth:** admin/password123

## 🔐 Default Credentials

**Admin User:**
- Email: `admin@example.com`
- Password: `admin123`

⚠️ **Change these in production!**

## 🧪 Testing the API

### Register a User
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe", 
    "email": "john@example.com",
    "password": "securepass123"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepass123"
  }'
```

### Admin Login
```bash
curl -X POST http://localhost:8080/api/auth/admin/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123"
  }'
```

## 🚨 Troubleshooting

### Build Issues
```bash
make clean
make deps
make build
```

### Database Connection Issues
```bash
make docker-down
make docker-up
make db-status
```

### View Logs
```bash
make docker-logs
```

## 🏗️ Architecture Highlights

### **Clean Architecture** 
- Clear separation of concerns
- Dependency injection pattern
- Interface-based design

### **Security Features**
- JWT authentication with role-based access
- BCrypt password hashing
- Rate limiting and CORS protection
- Input validation and sanitization

### **Monitoring & Logging**
- Structured JSON logging
- Async MongoDB activity logs
- Prometheus metrics
- Health check endpoints

### **Development Experience**
- Hot reload with Air
- Comprehensive Makefile
- Docker development environment
- Automated testing and linting

## 🚀 Production Deployment

1. **Update Configuration**
   ```bash
   cp config.sample.yaml config.production.yaml
   # Edit production values
   ```

2. **Build Docker Image**
   ```bash
   docker build -t user-mgmt-system .
   ```

3. **Deploy with Docker Compose**
   ```bash
   docker-compose --profile full up -d
   ```

## 📈 Next Steps

You can extend this system with:
- Email verification
- Password reset functionality  
- OAuth integration
- Advanced user roles
- File upload capabilities
- API versioning
- Caching with Redis

---

## 🎉 Success!

Your **User Management System** is ready for development! 

Start with `make docker-up` and `make dev` to begin coding. 🚀 