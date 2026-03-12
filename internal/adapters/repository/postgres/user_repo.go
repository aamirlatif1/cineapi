package postgres

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/aamir-al/cineapi/internal/application"
	"github.com/aamir-al/cineapi/internal/domain"
)

// userRepo implements application.UserRepository using GORM.
type userRepo struct{ db *gorm.DB }

// NewUserRepo returns a new UserRepository backed by db.
func NewUserRepo(db *gorm.DB) application.UserRepository {
	return &userRepo{db: db}
}

// Insert persists a new user and populates u.ID and u.CreatedAt.
func (r *userRepo) Insert(ctx context.Context, u *domain.User) error {
	model := userFromDomain(u)
	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		if isDuplicateKeyError(result.Error) {
			return domain.ErrDuplicateEmail
		}
		return result.Error
	}
	u.ID = model.ID
	u.CreatedAt = model.CreatedAt
	return nil
}

// GetByEmail retrieves a user by email address.
func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model UserModel
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&model)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return model.toDomain(), nil
}

// Update performs an optimistic-locking update on a user.
func (r *userRepo) Update(ctx context.Context, u *domain.User) error {
	result := r.db.WithContext(ctx).
		Model(&UserModel{}).
		Where("id = ? AND version = ?", u.ID, u.Version).
		Updates(map[string]any{
			"name":          u.Name,
			"email":         u.Email,
			"password_hash": u.Password.Hash,
			"activated":     u.Activated,
			"version":       gorm.Expr("version + 1"),
		})
	if result.Error != nil {
		if isDuplicateKeyError(result.Error) {
			return domain.ErrDuplicateEmail
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrEditConflict
	}
	u.Version++
	return nil
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}
