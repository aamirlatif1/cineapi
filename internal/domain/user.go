package domain

import (
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/aamir-al/cineapi/pkg/validator"
)

// User represents an account holder.
type User struct {
	ID        int64
	Name      string
	Email     string
	Password  Password
	Activated bool
	CreatedAt time.Time
	Version   int32
}

// Password wraps bcrypt hashing so plaintext never leaks out of this type.
type Password struct {
	plaintext *string
	Hash      []byte
}

// Set hashes plaintext with bcrypt cost 12 and stores the hash.
func (p *Password) Set(plaintext string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintext
	p.Hash = hash
	return nil
}

// Matches reports whether plaintext matches the stored bcrypt hash.
func (p *Password) Matches(plaintext string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.Hash, []byte(plaintext))
	if err != nil {
		switch {
		case err == bcrypt.ErrMismatchedHashAndPassword:
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

// NewUser creates and validates a User, hashing the password immediately.
func NewUser(name, email, plaintext string) (*User, map[string]string) {
	v := validator.New()

	v.Check(name != "", "name", "must be provided")
	v.Check(len(name) <= 500, "name", "must not be more than 500 bytes long")

	v.Check(email != "", "email", "must be provided")
	v.Check(len(email) <= 500, "email", "must not be more than 500 bytes long")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")

	v.Check(plaintext != "", "password", "must be provided")
	v.Check(len(plaintext) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(plaintext) <= 72, "password", "must not be more than 72 bytes long")

	if !v.Valid() {
		return nil, v.Errors()
	}

	u := &User{
		Name:      name,
		Email:     email,
		Activated: false,
		Version:   1,
	}
	if err := u.Password.Set(plaintext); err != nil {
		return nil, map[string]string{"password": "failed to hash password"}
	}
	return u, nil
}
