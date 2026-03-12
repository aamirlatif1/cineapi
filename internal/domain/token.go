package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"
)

// Token scope constants.
const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
	ScopePasswordReset  = "password-reset"
)

// Token is a short-lived, scoped credential issued to a user.
type Token struct {
	Plaintext string // Returned to client once; never stored.
	Hash      []byte // SHA-256(Plaintext); stored in repository.
	UserID    int64
	Expiry    time.Time
	Scope     string
}

// NewToken generates a cryptographically random token for the given user,
// with the given TTL and scope.
func NewToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, err
	}

	plaintext := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(plaintext))

	return &Token{
		Plaintext: plaintext,
		Hash:      hash[:],
		UserID:    userID,
		Expiry:    time.Now().Add(ttl),
		Scope:     scope,
	}, nil
}
