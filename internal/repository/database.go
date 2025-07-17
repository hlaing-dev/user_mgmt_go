package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"user_mgmt_go/internal/config"
	"user_mgmt_go/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database holds both PostgreSQL and MongoDB connections
type Database struct {
	PostgreSQL *gorm.DB
	MongoDB    *mongo.Database
	Config     *config.Config
}

// NewDatabase creates a new database instance with both connections
func NewDatabase(cfg *config.Config) (*Database, error) {
	db := &Database{
		Config: cfg,
	}

	// Initialize PostgreSQL connection
	if err := db.connectPostgreSQL(); err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Initialize MongoDB connection
	if err := db.connectMongoDB(); err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Run migrations for PostgreSQL
	if err := db.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create MongoDB indexes
	if err := db.createMongoIndexes(); err != nil {
		return nil, fmt.Errorf("failed to create MongoDB indexes: %w", err)
	}

	log.Println("✅ Database connections established successfully")
	return db, nil
}

// connectPostgreSQL establishes connection to PostgreSQL using GORM
func (d *Database) connectPostgreSQL() error {
	dsn := d.Config.GetDatabaseConnectionString()

	// Configure GORM logger based on environment
	var gormLogger logger.Interface
	if d.Config.Server.GinMode == "debug" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	// GORM configuration
	gormConfig := &gorm.Config{
		Logger:                 gormLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	var err error
	d.PostgreSQL, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := d.PostgreSQL.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)           // Maximum idle connections
	sqlDB.SetMaxOpenConns(100)          // Maximum open connections
	sqlDB.SetConnMaxLifetime(time.Hour) // Connection maximum lifetime

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	log.Printf("✅ PostgreSQL connected to: %s:%s/%s", 
		d.Config.Database.Host, 
		d.Config.Database.Port, 
		d.Config.Database.DBName)

	return nil
}

// connectMongoDB establishes connection to MongoDB
func (d *Database) connectMongoDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// MongoDB client options
	clientOptions := options.Client().
		ApplyURI(d.Config.MongoDB.URI).
		SetMaxPoolSize(100).                    // Maximum connection pool size
		SetMinPoolSize(10).                     // Minimum connection pool size
		SetMaxConnIdleTime(30 * time.Minute).  // Maximum connection idle time
		SetServerSelectionTimeout(5 * time.Second). // Server selection timeout
		SetSocketTimeout(10 * time.Second)     // Socket timeout

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Get database instance
	d.MongoDB = client.Database(d.Config.MongoDB.Database)

	log.Printf("✅ MongoDB connected to: %s/%s", 
		d.Config.MongoDB.URI, 
		d.Config.MongoDB.Database)

	return nil
}

// runMigrations runs PostgreSQL migrations
func (d *Database) runMigrations() error {
	log.Println("🔄 Running PostgreSQL migrations...")

	// Auto-migrate models
	if err := d.PostgreSQL.AutoMigrate(
		&models.User{},
	); err != nil {
		return fmt.Errorf("failed to run auto-migration: %w", err)
	}

	// Create custom indexes if needed
	if err := d.createPostgreSQLIndexes(); err != nil {
		return fmt.Errorf("failed to create PostgreSQL indexes: %w", err)
	}

	log.Println("✅ PostgreSQL migrations completed")
	return nil
}

// createPostgreSQLIndexes creates custom indexes for PostgreSQL
func (d *Database) createPostgreSQLIndexes() error {
	// Create index on email for faster lookups
	if err := d.PostgreSQL.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)").Error; err != nil {
		return fmt.Errorf("failed to create email index: %w", err)
	}

	// Create index on created_at for faster date-based queries
	if err := d.PostgreSQL.Exec("CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at)").Error; err != nil {
		return fmt.Errorf("failed to create created_at index: %w", err)
	}

	return nil
}

// createMongoIndexes creates indexes for MongoDB collections
func (d *Database) createMongoIndexes() error {
	log.Println("🔄 Creating MongoDB indexes...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := d.MongoDB.Collection(models.UserLog{}.CollectionName())

	// Create indexes
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
			},
			Options: options.Index().SetName("idx_user_id"),
		},
		{
			Keys: bson.D{
				{Key: "event", Value: 1},
			},
			Options: options.Index().SetName("idx_event"),
		},
		{
			Keys: bson.D{
				{Key: "timestamp", Value: -1}, // Descending for latest first
			},
			Options: options.Index().SetName("idx_timestamp"),
		},
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "timestamp", Value: -1},
			},
			Options: options.Index().SetName("idx_user_timestamp"),
		},
		{
			Keys: bson.D{
				{Key: "ip_address", Value: 1},
			},
			Options: options.Index().SetName("idx_ip_address").SetSparse(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create MongoDB indexes: %w", err)
	}

	log.Println("✅ MongoDB indexes created")
	return nil
}

// Close closes all database connections
func (d *Database) Close() error {
	// Close PostgreSQL connection
	if d.PostgreSQL != nil {
		if sqlDB, err := d.PostgreSQL.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				log.Printf("Error closing PostgreSQL connection: %v", err)
			}
		}
	}

	// Close MongoDB connection
	if d.MongoDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := d.MongoDB.Client().Disconnect(ctx); err != nil {
			log.Printf("Error closing MongoDB connection: %v", err)
			return err
		}
	}

	log.Println("✅ Database connections closed")
	return nil
}

// HealthCheck checks the health of both database connections
func (d *Database) HealthCheck() (bool, bool) {
	var pgHealthy, mongoHealthy bool

	// Check PostgreSQL
	if d.PostgreSQL != nil {
		if sqlDB, err := d.PostgreSQL.DB(); err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := sqlDB.PingContext(ctx); err == nil {
				pgHealthy = true
			}
		}
	}

	// Check MongoDB
	if d.MongoDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := d.MongoDB.Client().Ping(ctx, nil); err == nil {
			mongoHealthy = true
		}
	}

	return pgHealthy, mongoHealthy
} 