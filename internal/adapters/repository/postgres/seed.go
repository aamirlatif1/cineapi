package postgres

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SeedDemoUser inserts the pre-activated demo user if it does not already exist.
// Email: demo@cineapi.local  Password: pa55word
func SeedDemoUser(db *gorm.DB) error {
	hash, err := bcrypt.GenerateFromPassword([]byte("pa55word"), 12)
	if err != nil {
		return err
	}

	demo := &UserModel{
		Name:         "Demo User",
		Email:        "demo@cineapi.local",
		PasswordHash: hash,
		Activated:    true,
	}

	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(demo).Error
}
