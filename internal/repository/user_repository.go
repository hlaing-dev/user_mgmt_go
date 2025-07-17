package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"user_mgmt_go/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// userRepository implements the UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create creates a new user in the database
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("user with email %s already exists", user.Email)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

// Update updates a user's fields
func (r *userRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key") || strings.Contains(result.Error.Error(), "unique constraint") {
			return fmt.Errorf("email already exists")
		}
		return fmt.Errorf("failed to update user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user with ID %s not found", id)
	}
	return nil
}

// Delete soft deletes a user
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.User{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user with ID %s not found", id)
	}
	return nil
}

// List retrieves users with pagination and filtering
func (r *userRepository) List(ctx context.Context, params ListParams) (*models.UsersListResponse, error) {
	params.SetDefaults()
	
	if !IsValidUserSortField(params.SortBy) {
		params.SortBy = "created_at"
	}

	var users []models.User
	var total int64

	// Base query
	query := r.db.WithContext(ctx).Model(&models.User{})

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// Apply pagination and sorting
	orderClause := fmt.Sprintf("%s %s", params.SortBy, strings.ToUpper(params.SortDir))
	if err := query.Order(orderClause).
		Offset(params.GetOffset()).
		Limit(params.GetLimit()).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Convert to response format
	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	return &models.UsersListResponse{
		Users:      userResponses,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(total, params.PageSize),
	}, nil
}

// Count returns the total number of users matching the filter
func (r *userRepository) Count(ctx context.Context, filter UserFilter) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.User{})
	
	// Apply filters
	query = r.applyUserFilters(query, filter)
	
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	
	return count, nil
}

// CreateBatch creates multiple users in a single transaction
func (r *userRepository) CreateBatch(ctx context.Context, users []*models.User) error {
	if len(users) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.CreateInBatches(users, 100).Error; err != nil {
			return fmt.Errorf("failed to create users in batch: %w", err)
		}
		return nil
	})
}

// DeleteBatch soft deletes multiple users
func (r *userRepository) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).Delete(&models.User{}, ids)
	if result.Error != nil {
		return fmt.Errorf("failed to delete users in batch: %w", result.Error)
	}
	
	return nil
}

// Search searches users by name or email
func (r *userRepository) Search(ctx context.Context, query string, params ListParams) (*models.UsersListResponse, error) {
	params.SetDefaults()
	
	if !IsValidUserSortField(params.SortBy) {
		params.SortBy = "created_at"
	}

	var users []models.User
	var total int64

	searchTerm := "%" + strings.ToLower(query) + "%"
	
	// Build search query
	dbQuery := r.db.WithContext(ctx).Model(&models.User{}).Where(
		"LOWER(name) LIKE ? OR LOWER(email) LIKE ?", 
		searchTerm, searchTerm,
	)

	// Count total matching records
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	// Apply pagination and sorting
	orderClause := fmt.Sprintf("%s %s", params.SortBy, strings.ToUpper(params.SortDir))
	if err := dbQuery.Order(orderClause).
		Offset(params.GetOffset()).
		Limit(params.GetLimit()).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	// Convert to response format
	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	return &models.UsersListResponse{
		Users:      userResponses,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(total, params.PageSize),
	}, nil
}

// Exists checks if a user with the given email exists
func (r *userRepository) Exists(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return count > 0, nil
}

// GetAllDeleted retrieves all soft-deleted users
func (r *userRepository) GetAllDeleted(ctx context.Context, params ListParams) (*models.UsersListResponse, error) {
	params.SetDefaults()
	
	if !IsValidUserSortField(params.SortBy) {
		params.SortBy = "deleted_at"
	}

	var users []models.User
	var total int64

	// Query only soft-deleted records
	query := r.db.WithContext(ctx).Unscoped().Model(&models.User{}).Where("deleted_at IS NOT NULL")

	// Count total deleted records
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count deleted users: %w", err)
	}

	// Apply pagination and sorting
	orderClause := fmt.Sprintf("%s %s", params.SortBy, strings.ToUpper(params.SortDir))
	if err := query.Order(orderClause).
		Offset(params.GetOffset()).
		Limit(params.GetLimit()).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get deleted users: %w", err)
	}

	// Convert to response format
	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	return &models.UsersListResponse{
		Users:      userResponses,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: CalculateTotalPages(total, params.PageSize),
	}, nil
}

// RestoreDeleted restores a soft-deleted user
func (r *userRepository) RestoreDeleted(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Unscoped().Model(&models.User{}).Where("id = ? AND deleted_at IS NOT NULL", id).Update("deleted_at", nil)
	if result.Error != nil {
		return fmt.Errorf("failed to restore user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("deleted user with ID %s not found", id)
	}
	return nil
}

// PermanentDelete permanently deletes a user from the database
func (r *userRepository) PermanentDelete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Unscoped().Delete(&models.User{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to permanently delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user with ID %s not found", id)
	}
	return nil
}

// applyUserFilters applies filters to the query
func (r *userRepository) applyUserFilters(query *gorm.DB, filter UserFilter) *gorm.DB {
	if filter.Email != "" {
		query = query.Where("LOWER(email) LIKE ?", "%"+strings.ToLower(filter.Email)+"%")
	}
	
	if filter.Name != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(filter.Name)+"%")
	}
	
	if filter.CreatedAt != nil {
		if filter.CreatedAt.From != nil {
			if fromTime, err := time.Parse(time.RFC3339, *filter.CreatedAt.From); err == nil {
				query = query.Where("created_at >= ?", fromTime)
			}
		}
		if filter.CreatedAt.To != nil {
			if toTime, err := time.Parse(time.RFC3339, *filter.CreatedAt.To); err == nil {
				query = query.Where("created_at <= ?", toTime)
			}
		}
	}
	
	if filter.UpdatedAt != nil {
		if filter.UpdatedAt.From != nil {
			if fromTime, err := time.Parse(time.RFC3339, *filter.UpdatedAt.From); err == nil {
				query = query.Where("updated_at >= ?", fromTime)
			}
		}
		if filter.UpdatedAt.To != nil {
			if toTime, err := time.Parse(time.RFC3339, *filter.UpdatedAt.To); err == nil {
				query = query.Where("updated_at <= ?", toTime)
			}
		}
	}
	
	if filter.IsDeleted != nil {
		if *filter.IsDeleted {
			query = query.Unscoped().Where("deleted_at IS NOT NULL")
		} else {
			query = query.Where("deleted_at IS NULL")
		}
	}
	
	return query
} 