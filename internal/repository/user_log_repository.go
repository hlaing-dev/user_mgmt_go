package repository

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"user_mgmt_go/internal/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// userLogRepository implements the UserLogRepository interface
type userLogRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
	logChannel chan *models.UserLog
	wg         *sync.WaitGroup
	stopChan   chan struct{}
}

// NewUserLogRepository creates a new user log repository with async logging capability
func NewUserLogRepository(db *mongo.Database) UserLogRepository {
	repo := &userLogRepository{
		db:         db,
		collection: db.Collection(models.UserLog{}.CollectionName()),
		logChannel: make(chan *models.UserLog, 1000), // Buffer size of 1000
		wg:         &sync.WaitGroup{},
		stopChan:   make(chan struct{}),
	}

	// Start async log processor
	repo.startAsyncProcessor()

	return repo
}

// startAsyncProcessor starts the goroutine that processes async logs
func (r *userLogRepository) startAsyncProcessor() {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		
		// Batch processing variables
		batch := make([]*models.UserLog, 0, 10)
		ticker := time.NewTicker(5 * time.Second) // Process batch every 5 seconds
		defer ticker.Stop()

		for {
			select {
			case logEntry := <-r.logChannel:
				batch = append(batch, logEntry)
				
				// Process batch when it reaches size limit
				if len(batch) >= 10 {
					r.processBatch(batch)
					batch = batch[:0] // Reset batch
				}

			case <-ticker.C:
				// Process remaining logs in batch on timer
				if len(batch) > 0 {
					r.processBatch(batch)
					batch = batch[:0]
				}

			case <-r.stopChan:
				// Process remaining logs before shutting down
				if len(batch) > 0 {
					r.processBatch(batch)
				}
				return
			}
		}
	}()
}

// processBatch processes a batch of logs
func (r *userLogRepository) processBatch(logs []*models.UserLog) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := r.BulkCreate(ctx, logs); err != nil {
		log.Printf("Failed to process log batch: %v", err)
		// In production, you might want to implement retry logic or dead letter queue
	}
}

// Create creates a new log entry synchronously
func (r *userLogRepository) Create(ctx context.Context, logEntry *models.UserLog) error {
	if logEntry.Timestamp.IsZero() {
		logEntry.Timestamp = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, logEntry)
	if err != nil {
		return fmt.Errorf("failed to create log entry: %w", err)
	}
	return nil
}

// CreateAsync creates a new log entry asynchronously
func (r *userLogRepository) CreateAsync(logEntry *models.UserLog) error {
	if logEntry.Timestamp.IsZero() {
		logEntry.Timestamp = time.Now()
	}

	select {
	case r.logChannel <- logEntry:
		return nil
	default:
		// Channel is full, log synchronously as fallback
		log.Printf("Async log channel full, falling back to sync logging")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return r.Create(ctx, logEntry)
	}
}

// GetByID retrieves a log entry by ID
func (r *userLogRepository) GetByID(ctx context.Context, id string) (*models.UserLog, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid log ID format: %w", err)
	}

	var logEntry models.UserLog
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&logEntry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("log entry with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get log entry: %w", err)
	}
	return &logEntry, nil
}

// List retrieves logs with advanced filtering
func (r *userLogRepository) List(ctx context.Context, filter models.LogFilterRequest) (*models.UserLogsListResponse, error) {
	// Set defaults
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 10
	}

	// Build MongoDB filter
	mongoFilter := r.buildLogFilter(filter)

	// Count total documents
	total, err := r.collection.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to count logs: %w", err)
	}

	// Build find options
	opts := options.Find().
		SetSkip(int64((filter.Page - 1) * filter.PageSize)).
		SetLimit(int64(filter.PageSize)).
		SetSort(bson.D{{Key: "timestamp", Value: -1}}) // Latest first

	// Execute query
	cursor, err := r.collection.Find(ctx, mongoFilter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find logs: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode results
	var logs []models.UserLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("failed to decode logs: %w", err)
	}

	// Convert to response format
	logResponses := make([]models.UserLogResponse, len(logs))
	for i, logEntry := range logs {
		logResponses[i] = logEntry.ToResponse()
	}

	return &models.UserLogsListResponse{
		Logs:       logResponses,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: CalculateTotalPages(total, filter.PageSize),
	}, nil
}

// GetByUserID retrieves logs for a specific user
func (r *userLogRepository) GetByUserID(ctx context.Context, userID uuid.UUID, params ListParams) (*models.UserLogsListResponse, error) {
	params.SetDefaults()

	filter := bson.M{"user_id": userID}
	
	// Count total documents
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count user logs: %w", err)
	}

	// Build find options
	opts := options.Find().
		SetSkip(int64(params.GetOffset())).
		SetLimit(int64(params.GetLimit())).
		SetSort(bson.D{{Key: "timestamp", Value: -1}})

	// Execute query
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find user logs: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode results
	var logs []models.UserLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("failed to decode user logs: %w", err)
	}

	// Convert to response format
	logResponses := make([]models.UserLogResponse, len(logs))
	for i, logEntry := range logs {
		logResponses[i] = logEntry.ToResponse()
	}

	return &models.UserLogsListResponse{
		Logs:       logResponses,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(total, params.PageSize),
	}, nil
}

// GetByEvent retrieves logs by event type
func (r *userLogRepository) GetByEvent(ctx context.Context, event models.LogEventType, params ListParams) (*models.UserLogsListResponse, error) {
	params.SetDefaults()

	filter := bson.M{"event": event}
	
	// Count total documents
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count event logs: %w", err)
	}

	// Build find options
	opts := options.Find().
		SetSkip(int64(params.GetOffset())).
		SetLimit(int64(params.GetLimit())).
		SetSort(bson.D{{Key: "timestamp", Value: -1}})

	// Execute query
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find event logs: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode results
	var logs []models.UserLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("failed to decode event logs: %w", err)
	}

	// Convert to response format
	logResponses := make([]models.UserLogResponse, len(logs))
	for i, logEntry := range logs {
		logResponses[i] = logEntry.ToResponse()
	}

	return &models.UserLogsListResponse{
		Logs:       logResponses,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(total, params.PageSize),
	}, nil
}

// Count returns the total number of logs matching the filter
func (r *userLogRepository) Count(ctx context.Context, filter models.LogFilterRequest) (int64, error) {
	mongoFilter := r.buildLogFilter(filter)
	return r.collection.CountDocuments(ctx, mongoFilter)
}

// GetEventStats returns statistics about events for a user within a time period
func (r *userLogRepository) GetEventStats(ctx context.Context, userID *uuid.UUID, days int) (map[models.LogEventType]int64, error) {
	matchStage := bson.M{
		"timestamp": bson.M{
			"$gte": time.Now().AddDate(0, 0, -days),
		},
	}
	
	if userID != nil {
		matchStage["user_id"] = *userID
	}

	pipeline := []bson.M{
		{"$match": matchStage},
		{
			"$group": bson.M{
				"_id":   "$event",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get event stats: %w", err)
	}
	defer cursor.Close(ctx)

	stats := make(map[models.LogEventType]int64)
	for cursor.Next(ctx) {
		var result struct {
			ID    models.LogEventType `bson:"_id"`
			Count int64               `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		stats[result.ID] = result.Count
	}

	return stats, nil
}

// GetUserActivity returns recent activity for a user
func (r *userLogRepository) GetUserActivity(ctx context.Context, userID uuid.UUID, days int) ([]models.UserLogResponse, error) {
	filter := bson.M{
		"user_id": userID,
		"timestamp": bson.M{
			"$gte": time.Now().AddDate(0, 0, -days),
		},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(50) // Limit to 50 recent activities

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get user activity: %w", err)
	}
	defer cursor.Close(ctx)

	var logs []models.UserLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("failed to decode user activity: %w", err)
	}

	// Convert to response format
	activities := make([]models.UserLogResponse, len(logs))
	for i, logEntry := range logs {
		activities[i] = logEntry.ToResponse()
	}

	return activities, nil
}

// DeleteOldLogs deletes logs older than specified days
func (r *userLogRepository) DeleteOldLogs(ctx context.Context, olderThanDays int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -olderThanDays)
	filter := bson.M{
		"timestamp": bson.M{
			"$lt": cutoffDate,
		},
	}

	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old logs: %w", err)
	}

	return result.DeletedCount, nil
}

// BulkCreate creates multiple log entries in a single operation
func (r *userLogRepository) BulkCreate(ctx context.Context, logs []*models.UserLog) error {
	if len(logs) == 0 {
		return nil
	}

	// Convert to interface slice for MongoDB
	documents := make([]interface{}, len(logs))
	for i, logEntry := range logs {
		if logEntry.Timestamp.IsZero() {
			logEntry.Timestamp = time.Now()
		}
		documents[i] = logEntry
	}

	_, err := r.collection.InsertMany(ctx, documents)
	if err != nil {
		return fmt.Errorf("failed to bulk create logs: %w", err)
	}

	return nil
}

// SearchLogs searches logs based on a search term
func (r *userLogRepository) SearchLogs(ctx context.Context, searchTerm string, filter models.LogFilterRequest) (*models.UserLogsListResponse, error) {
	// Set defaults
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 10
	}

	// Build search filter
	searchFilter := bson.M{
		"$or": []bson.M{
			{"event": bson.M{"$regex": searchTerm, "$options": "i"}},
			{"data.action": bson.M{"$regex": searchTerm, "$options": "i"}},
			{"data.error": bson.M{"$regex": searchTerm, "$options": "i"}},
			{"ip_address": bson.M{"$regex": searchTerm, "$options": "i"}},
		},
	}

	// Combine with existing filter
	mongoFilter := r.buildLogFilter(filter)
	combinedFilter := bson.M{
		"$and": []bson.M{mongoFilter, searchFilter},
	}

	// Count total documents
	total, err := r.collection.CountDocuments(ctx, combinedFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	// Build find options
	opts := options.Find().
		SetSkip(int64((filter.Page - 1) * filter.PageSize)).
		SetLimit(int64(filter.PageSize)).
		SetSort(bson.D{{Key: "timestamp", Value: -1}})

	// Execute query
	cursor, err := r.collection.Find(ctx, combinedFilter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to search logs: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode results
	var logs []models.UserLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("failed to decode search results: %w", err)
	}

	// Convert to response format
	logResponses := make([]models.UserLogResponse, len(logs))
	for i, logEntry := range logs {
		logResponses[i] = logEntry.ToResponse()
	}

	return &models.UserLogsListResponse{
		Logs:       logResponses,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: CalculateTotalPages(total, filter.PageSize),
	}, nil
}

// buildLogFilter builds MongoDB filter from LogFilterRequest
func (r *userLogRepository) buildLogFilter(filter models.LogFilterRequest) bson.M {
	mongoFilter := bson.M{}

	if filter.UserID != nil {
		mongoFilter["user_id"] = *filter.UserID
	}

	if filter.Event != nil {
		mongoFilter["event"] = *filter.Event
	}

	if filter.IPAddress != "" {
		mongoFilter["ip_address"] = bson.M{"$regex": filter.IPAddress, "$options": "i"}
	}

	if filter.Action != "" {
		mongoFilter["data.action"] = bson.M{"$regex": filter.Action, "$options": "i"}
	}

	// Date range filter
	if filter.StartDate != nil || filter.EndDate != nil {
		timeFilter := bson.M{}
		if filter.StartDate != nil {
			timeFilter["$gte"] = *filter.StartDate
		}
		if filter.EndDate != nil {
			timeFilter["$lte"] = *filter.EndDate
		}
		mongoFilter["timestamp"] = timeFilter
	}

	return mongoFilter
}

// Close gracefully shuts down the async processor
func (r *userLogRepository) Close() {
	close(r.stopChan)
	r.wg.Wait()
	close(r.logChannel)
} 