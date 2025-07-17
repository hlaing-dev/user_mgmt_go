package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"user_mgmt_go/internal/config"
)

// Test Configuration Validation (Basic)
func TestConfigValidation(t *testing.T) {
	t.Run("Valid Configuration", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				Host:    "localhost",
				Port:    "8080",
				GinMode: "debug",
			},
			Database: config.DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "postgres",
				Password: "password123",
				DBName:   "user_mgmt",
				SSLMode:  "disable",
			},
			JWT: config.JWTConfig{
				Secret: "valid-secret-key-with-sufficient-length",
				Expiry: 0, // Will be set by duration parsing
			},
			Admin: config.AdminConfig{
				Email:    "admin@example.com",
				Password: "secure-admin-password",
			},
		}

		// Basic validation - config exists
		assert.NotNil(t, cfg)
		assert.Equal(t, "localhost", cfg.Server.Host)
		assert.Equal(t, "8080", cfg.Server.Port)
	})
}

// Test Database Connection String Generation
func TestDatabaseConfig(t *testing.T) {
	t.Run("PostgreSQL Connection String Generation", func(t *testing.T) {
		cfg := &config.Config{
			Database: config.DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "postgres",
				Password: "password123",
				DBName:   "user_mgmt",
				SSLMode:  "disable",
			},
		}

		connectionString := cfg.GetDatabaseConnectionString()
		expected := "host=localhost port=5432 user=postgres password=password123 dbname=user_mgmt sslmode=disable"
		assert.Equal(t, expected, connectionString)
	})

	t.Run("Database Config Fields", func(t *testing.T) {
		cfg := &config.Config{
			Database: config.DatabaseConfig{
				Host:     "test-host",
				Port:     "5433",
				User:     "test-user",
				Password: "test-password",
				DBName:   "test-db",
				SSLMode:  "require",
			},
		}

		assert.Equal(t, "test-host", cfg.Database.Host)
		assert.Equal(t, "5433", cfg.Database.Port)
		assert.Equal(t, "test-user", cfg.Database.User)
		assert.Equal(t, "test-password", cfg.Database.Password)
		assert.Equal(t, "test-db", cfg.Database.DBName)
		assert.Equal(t, "require", cfg.Database.SSLMode)
	})
}

// Test MongoDB Configuration
func TestMongoDBConfig(t *testing.T) {
	cfg := &config.Config{
		MongoDB: config.MongoConfig{
			URI:      "mongodb://localhost:27017",
			Database: "user_logs",
		},
	}

	assert.Equal(t, "mongodb://localhost:27017", cfg.MongoDB.URI)
	assert.Equal(t, "user_logs", cfg.MongoDB.Database)
}

// Test JWT Configuration
func TestJWTConfig(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key",
			Expiry: 0, // Duration would be set by parsing
		},
	}

	assert.Equal(t, "test-secret-key", cfg.JWT.Secret)
	assert.NotNil(t, cfg.JWT.Expiry)
}

// Test Admin Configuration
func TestAdminConfig(t *testing.T) {
	cfg := &config.Config{
		Admin: config.AdminConfig{
			Email:    "admin@test.com",
			Password: "admin-password",
		},
	}

	assert.Equal(t, "admin@test.com", cfg.Admin.Email)
	assert.Equal(t, "admin-password", cfg.Admin.Password)
}

// Test CORS Configuration
func TestCORSConfig(t *testing.T) {
	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000", "https://example.com"},
		},
	}

	assert.Len(t, cfg.CORS.AllowedOrigins, 2)
	assert.Contains(t, cfg.CORS.AllowedOrigins, "http://localhost:3000")
	assert.Contains(t, cfg.CORS.AllowedOrigins, "https://example.com")
}

// Test Server Configuration
func TestServerConfig(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:    "0.0.0.0",
			Port:    "9000",
			GinMode: "release",
		},
	}

	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, "9000", cfg.Server.Port)
	assert.Equal(t, "release", cfg.Server.GinMode)
}

// Test Logging Configuration
func TestLoggingConfig(t *testing.T) {
	cfg := &config.Config{
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}

	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "text", cfg.Logging.Format)
}

// Test Config Structure Completeness
func TestConfigStructure(t *testing.T) {
	cfg := &config.Config{}

	// Verify all main sections exist
	assert.NotNil(t, &cfg.Server)
	assert.NotNil(t, &cfg.Database)
	assert.NotNil(t, &cfg.MongoDB)
	assert.NotNil(t, &cfg.JWT)
	assert.NotNil(t, &cfg.Admin)
	assert.NotNil(t, &cfg.CORS)
	assert.NotNil(t, &cfg.Logging)
} 