package domain

import "errors"

var (
	// ErrNotFound is returned when a requested record does not exist.
	ErrNotFound = errors.New("record not found")

	// ErrDuplicateEmail is returned when a user registration uses an already-registered email.
	ErrDuplicateEmail = errors.New("duplicate email address")

	// ErrEditConflict is returned when an optimistic-locking version mismatch is detected.
	ErrEditConflict = errors.New("edit conflict")

	// ErrInvalidToken is returned when a submitted token is invalid, expired, or already used.
	ErrInvalidToken = errors.New("invalid or expired token")
)
