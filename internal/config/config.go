package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for our application
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	MongoDB  MongoConfig    `mapstructure:"mongodb"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Admin    AdminConfig    `mapstructure:"admin"`
	CORS     CORSConfig     `mapstructure:"cors"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port    string `mapstructure:"port"`
	Host    string `mapstructure:"host"`
	GinMode string `mapstructure:"gin_mode"`
}

// DatabaseConfig holds PostgreSQL database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

// MongoConfig holds MongoDB configuration
type MongoConfig struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string        `mapstructure:"secret"`
	Expiry time.Duration `mapstructure:"expiry"`
}

// AdminConfig holds admin user configuration
type AdminConfig struct {
	Email    string `mapstructure:"email"`
	Password string `mapstructure:"password"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// LoadConfig loads configuration from environment variables and config files
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Set default values
	setDefaults()

	// Enable reading from environment variables
	viper.AutomaticEnv()

	// Bind environment variables with custom names
	bindEnvVars()

	err = viper.ReadInConfig()
	if err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return
	}

	return
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.gin_mode", "debug")

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "password123")
	viper.SetDefault("database.db_name", "user_mgmt")
	viper.SetDefault("database.ssl_mode", "disable")

	// MongoDB defaults
	viper.SetDefault("mongodb.uri", "mongodb://admin:password123@localhost:27017")
	viper.SetDefault("mongodb.database", "user_logs")

	// JWT defaults
	viper.SetDefault("jwt.secret", "your-super-secret-jwt-key")
	viper.SetDefault("jwt.expiry", "24h")

	// Admin defaults
	viper.SetDefault("admin.email", "admin@example.com")
	viper.SetDefault("admin.password", "admin123")

	// CORS defaults
	viper.SetDefault("cors.allowed_origins", []string{"http://localhost:3000", "http://localhost:3001"})

	// Logging defaults
	viper.SetDefault("logging.level", "debug")
	viper.SetDefault("logging.format", "json")
}

// bindEnvVars binds environment variables to configuration keys
func bindEnvVars() {
	// Server
	viper.BindEnv("server.port", "PORT")
	viper.BindEnv("server.host", "HOST")
	viper.BindEnv("server.gin_mode", "GIN_MODE")

	// Database
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.db_name", "DB_NAME")
	viper.BindEnv("database.ssl_mode", "DB_SSLMODE")

	// MongoDB
	viper.BindEnv("mongodb.uri", "MONGO_URI")
	viper.BindEnv("mongodb.database", "MONGO_DATABASE")

	// JWT
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("jwt.expiry", "JWT_EXPIRY")

	// Admin
	viper.BindEnv("admin.email", "ADMIN_EMAIL")
	viper.BindEnv("admin.password", "ADMIN_PASSWORD")

	// CORS
	viper.BindEnv("cors.allowed_origins", "ALLOWED_ORIGINS")

	// Logging
	viper.BindEnv("logging.level", "LOG_LEVEL")
	viper.BindEnv("logging.format", "LOG_FORMAT")
}

// GetDatabaseConnectionString returns the database connection string
func (c *Config) GetDatabaseConnectionString() string {
	return "host=" + c.Database.Host +
		" port=" + c.Database.Port +
		" user=" + c.Database.User +
		" password=" + c.Database.Password +
		" dbname=" + c.Database.DBName +
		" sslmode=" + c.Database.SSLMode
} 