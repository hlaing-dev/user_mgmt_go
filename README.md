# User Management System - Backend

## Overview
A comprehensive user management system built with Go, featuring admin authentication, CRUD operations, and asynchronous logging. Designed for senior-level assessment with clean architecture and best practices.

## 🚀 Quick Start

### First Time Setup (New Machine)
```bash
# 1. Clone the repository
git clone <your-repo-url>
cd user_mgmt_go

# 2. Run the automated setup script
chmod +x setup.sh
./setup.sh
```

### Daily Development
```bash
# Quick start for daily use
./start.sh

# Or manually:
make docker-up    # Start databases
make run          # Start server on port 8080
# PORT=8081 make run  # Different port
```

### Access Points
- **Health Check**: http://localhost:8080/health
- **Swagger Docs**: http://localhost:8080/swagger/index.html
- **Default Admin**: `admin@example.com` / `admin123`

### 📖 Documentation
- **[Architecture & Technical Guide](ARCHITECTURE.md)** - Comprehensive technical documentation
- **[Quick Start Guide](QUICKSTART.md)** - Complete setup and usage guide
- **[Development Guide](DEVELOPMENT.md)** - Advanced development topics  
- **[Testing Guide](TESTING.md)** - Comprehensive testing information

## Tech Stack

### Backend Framework
- **Go 1.21+**: Primary programming language
- **Gin**: High-performance HTTP web framework
  - Fast routing and middleware support
  - JSON binding and validation
  - Excellent performance for REST APIs

### Databases
- **PostgreSQL**: Relational database for user data
  - ACID compliance for user operations
  - Strong consistency for authentication
  - Excellent Go ecosystem support

- **MongoDB**: NoSQL database for logging
  - Flexible schema for various log events
  - High-write performance for async logging
  - Easy horizontal scaling

### ORM & Database Drivers
- **GORM**: Object-Relational Mapping for PostgreSQL
  - Type-safe database operations
  - Automatic migrations
  - Relationships and associations

- **MongoDB Go Driver**: Official MongoDB driver
  - Native Go integration
  - Connection pooling
  - Aggregation pipeline support

### Authentication & Security
- **JWT (JSON Web Tokens)**: Stateless authentication
  - Secure token-based auth
  - Cross-service compatibility
  - Easy frontend integration

- **bcrypt**: Password hashing
  - Industry-standard password security
  - Salt generation and verification
  - Resistance to timing attacks

### Configuration & Environment
- **Viper**: Configuration management
  - Multiple config formats support
  - Environment variable binding
  - Live config reloading

### API Documentation
- **Swagger/OpenAPI**: API documentation
  - Interactive API explorer
  - Automatic documentation generation
  - Frontend integration support

### Testing
- **Go Testing**: Built-in testing framework
- **Testify**: Enhanced testing assertions
  - Rich assertion library
  - Mocking capabilities
  - Suite-based testing

### CORS Support
- **Gin-CORS**: Cross-Origin Resource Sharing
  - React frontend integration
  - Configurable CORS policies
  - Preflight request handling

## Architecture

### Clean Architecture Pattern
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│    Handlers     │────│    Services     │────│   Repository    │
│   (HTTP Layer)  │    │ (Business Logic)│    │  (Data Layer)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
    ┌─────────┐            ┌─────────┐              ┌─────────┐
    │   Gin   │            │  Models │              │Database │
    │Framework│            │ & DTOs  │              │ Layers  │
    └─────────┘            └─────────┘              └─────────┘
```

### Project Structure
```
user_mgmt_go/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Configuration management
│   ├── handlers/                # HTTP request handlers
│   ├── middleware/              # Custom middleware
│   ├── models/                  # Data models and DTOs
│   ├── services/                # Business logic layer
│   ├── repository/              # Data access layer
│   └── utils/                   # Utility functions
├── pkg/                         # Public packages
├── migrations/                  # Database migrations
├── tests/                       # Test files
├── docs/                        # API documentation
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
├── .env.example                 # Environment variables template
├── docker-compose.yml           # Local development setup
└── README.md                    # This file
```

## Database Schemas

### Users Table (PostgreSQL)
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### User Logs Collection (MongoDB)
```json
{
  "_id": "ObjectId",
  "user_id": "UUID",
  "event": "string",
  "data": {
    "action": "CREATE|UPDATE|DELETE|LOGIN",
    "details": "object"
  },
  "timestamp": "ISODate"
}
```

## Features

### Core Functionality
- ✅ Admin authentication with JWT tokens
- ✅ User CRUD operations (Create, Read, Update, Delete)
- ✅ Data table format for user listing
- ✅ Asynchronous logging mechanism
- ✅ Input validation and error handling
- ✅ CORS support for React frontend

### Security Features
- ✅ Password hashing with bcrypt
- ✅ JWT token authentication
- ✅ Input sanitization
- ✅ SQL injection prevention (via GORM)
- ✅ Rate limiting middleware

### Logging System
- ✅ Asynchronous event logging
- ✅ Go channels for non-blocking operations
- ✅ Structured logging with context
- ✅ Event categorization and filtering

## API Endpoints

### Authentication
- `POST /api/auth/login` - Admin login
- `POST /api/auth/refresh` - Token refresh

### User Management
- `GET /api/users` - List all users (paginated)
- `GET /api/users/:id` - Get user by ID
- `POST /api/users` - Create new user
- `PUT /api/users/:id` - Update user
- `DELETE /api/users/:id` - Delete user

### Logging
- `GET /api/logs` - Get user logs (admin only)
- `GET /api/logs/:userId` - Get logs for specific user

## Development Setup

### Prerequisites
- Go 1.21 or higher
- PostgreSQL 14+
- MongoDB 6.0+
- Docker & Docker Compose (optional)

### Installation Steps
1. Clone the repository
2. Copy `.env.example` to `.env` and configure
3. Install dependencies: `go mod tidy`
4. Run migrations: `go run cmd/migrate/main.go`
5. Start the server: `go run cmd/server/main.go`

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test package
go test ./internal/services/...
```

## Environment Variables
```env
# Server Configuration
PORT=8080
GIN_MODE=release

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=user
DB_PASSWORD=password
DB_NAME=user_mgmt

# MongoDB Configuration
MONGO_URI=mongodb://localhost:27017
MONGO_DB=user_logs

# JWT Configuration
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h

# Admin Credentials
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=admin123
```

## Docker Development
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## Performance Considerations
- Connection pooling for both PostgreSQL and MongoDB
- Asynchronous logging to prevent blocking main operations
- Pagination for large datasets
- Database indexing for optimal query performance
- JWT token caching for reduced database calls

## Monitoring & Observability
- Structured logging with different levels
- Request/response logging middleware
- Performance metrics collection
- Health check endpoints

## Next Steps
1. ✅ Backend API development
2. 🔄 React frontend development
3. 🔄 Integration testing
4. 🔄 Deployment configuration
5. 🔄 Production optimizations

---

This backend is designed to seamlessly integrate with a React frontend, providing a robust foundation for user management operations.