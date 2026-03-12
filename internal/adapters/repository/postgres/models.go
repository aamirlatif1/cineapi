package postgres

import (
	"time"

	"github.com/lib/pq"

	"github.com/aamir-al/cineapi/internal/domain"
)

// ----------------------------------------
// MovieModel — GORM adapter DTO for movies
// ----------------------------------------

// MovieModel is the GORM representation of a movie row. It MUST NOT be used
// outside of this package. Domain entities carry no GORM tags.
type MovieModel struct {
	ID        int64          `gorm:"primaryKey;autoIncrement;column:id"`
	Title     string         `gorm:"not null;column:title"`
	Year      int32          `gorm:"not null;column:year"`
	Runtime   int32          `gorm:"not null;column:runtime"`
	Genres    pq.StringArray `gorm:"type:text[];not null;column:genres"`
	CreatedAt time.Time      `gorm:"autoCreateTime;column:created_at"`
	Version   int32          `gorm:"default:1;not null;column:version"`
}

func (MovieModel) TableName() string { return "movies" }

func (m *MovieModel) toDomain() *domain.Movie {
	return &domain.Movie{
		ID:        m.ID,
		Title:     m.Title,
		Year:      m.Year,
		Runtime:   m.Runtime,
		Genres:    []string(m.Genres),
		CreatedAt: m.CreatedAt,
		Version:   m.Version,
	}
}

func movieFromDomain(m *domain.Movie) *MovieModel {
	return &MovieModel{
		ID:        m.ID,
		Title:     m.Title,
		Year:      m.Year,
		Runtime:   m.Runtime,
		Genres:    pq.StringArray(m.Genres),
		CreatedAt: m.CreatedAt,
		Version:   m.Version,
	}
}

// ----------------------------------------
// UserModel — GORM adapter DTO for users
// ----------------------------------------

// UserModel is the GORM representation of a user row.
type UserModel struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;column:id"`
	Name         string    `gorm:"not null;column:name"`
	Email        string    `gorm:"uniqueIndex;not null;column:email"`
	PasswordHash []byte    `gorm:"not null;column:password_hash"`
	Activated    bool      `gorm:"default:false;not null;column:activated"`
	CreatedAt    time.Time `gorm:"autoCreateTime;column:created_at"`
	Version      int32     `gorm:"default:1;not null;column:version"`
}

func (UserModel) TableName() string { return "users" }

func (u *UserModel) toDomain() *domain.User {
	user := &domain.User{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Activated: u.Activated,
		CreatedAt: u.CreatedAt,
		Version:   u.Version,
	}
	user.Password.Hash = u.PasswordHash
	return user
}

func userFromDomain(u *domain.User) *UserModel {
	return &UserModel{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.Password.Hash,
		Activated:    u.Activated,
		CreatedAt:    u.CreatedAt,
		Version:      u.Version,
	}
}

// ----------------------------------------
// TokenModel — GORM adapter DTO for tokens
// ----------------------------------------

// TokenModel is the GORM representation of a token row.
type TokenModel struct {
	Hash   []byte    `gorm:"primaryKey;column:hash"`
	UserID int64     `gorm:"not null;index;column:user_id"`
	Expiry time.Time `gorm:"not null;column:expiry"`
	Scope  string    `gorm:"not null;column:scope"`
}

func (TokenModel) TableName() string { return "tokens" }

func tokenFromDomain(t *domain.Token) *TokenModel {
	return &TokenModel{
		Hash:   t.Hash,
		UserID: t.UserID,
		Expiry: t.Expiry,
		Scope:  t.Scope,
	}
}
