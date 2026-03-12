package postgres

import (
	"context"
	"crypto/sha256"
	"time"

	"gorm.io/gorm"

	"github.com/aamir-al/cineapi/internal/application"
	"github.com/aamir-al/cineapi/internal/domain"
)

// tokenRepo implements application.TokenRepository using GORM.
type tokenRepo struct{ db *gorm.DB }

// NewTokenRepo returns a new TokenRepository backed by db.
func NewTokenRepo(db *gorm.DB) application.TokenRepository {
	return &tokenRepo{db: db}
}

// Insert persists a token (its hash, not plaintext).
func (r *tokenRepo) Insert(ctx context.Context, t *domain.Token) error {
	model := tokenFromDomain(t)
	return r.db.WithContext(ctx).Create(model).Error
}

// GetForUser looks up the user who owns the given plaintext token for the given scope.
// Returns ErrInvalidToken when the token is missing, expired, or invalid.
func (r *tokenRepo) GetForUser(ctx context.Context, scope, plaintext string) (*domain.User, error) {
	hash := sha256.Sum256([]byte(plaintext))

	var userModel UserModel
	result := r.db.WithContext(ctx).
		Joins("JOIN tokens ON tokens.user_id = users.id").
		Where("tokens.hash = ? AND tokens.scope = ? AND tokens.expiry > ?", hash[:], scope, time.Now()).
		First(&userModel)

	if result.Error != nil {
		return nil, domain.ErrInvalidToken
	}

	return userModel.toDomain(), nil
}

// DeleteAllForUser removes all tokens for a user with the given scope.
func (r *tokenRepo) DeleteAllForUser(ctx context.Context, scope string, userID int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND scope = ?", userID, scope).
		Delete(&TokenModel{}).Error
}
