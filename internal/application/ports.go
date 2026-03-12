// Package application defines use-case interfaces (ports) for the CineAPI service.
// Concrete adapter implementations live in internal/adapters/.
package application

import (
	"context"

	"github.com/aamir-al/cineapi/internal/domain"
)

// ----------------------------------------
// Movie port
// ----------------------------------------

// MovieFilters carries optional filter, sort, and pagination parameters for listing movies.
type MovieFilters struct {
	Title    string
	Genres   []string
	Page     int
	PageSize int
	Sort     string // e.g. "year" or "-year"; prefix "-" means descending
}

// Metadata describes the pagination state of a list response.
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

// MovieRepository is the driven port for movie persistence.
type MovieRepository interface {
	Insert(ctx context.Context, movie *domain.Movie) error
	Get(ctx context.Context, id int64) (*domain.Movie, error)
	GetAll(ctx context.Context, filters MovieFilters) ([]*domain.Movie, Metadata, error)
	Update(ctx context.Context, movie *domain.Movie) error
	Delete(ctx context.Context, id int64) error
}

// ----------------------------------------
// User port
// ----------------------------------------

// UserRepository is the driven port for user persistence.
type UserRepository interface {
	Insert(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}

// ----------------------------------------
// Token port
// ----------------------------------------

// TokenRepository is the driven port for token persistence.
type TokenRepository interface {
	Insert(ctx context.Context, token *domain.Token) error
	// GetForUser looks up the user who owns the given scoped token.
	// Returns ErrInvalidToken when the token does not exist or has expired.
	GetForUser(ctx context.Context, scope string, tokenPlaintext string) (*domain.User, error)
	DeleteAllForUser(ctx context.Context, scope string, userID int64) error
}
