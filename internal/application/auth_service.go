package application

import (
	"context"
	"time"

	"github.com/aamir-al/cineapi/internal/domain"
)

// AuthServiceImpl implements authentication use cases.
type AuthServiceImpl struct {
	userRepo  UserRepository
	tokenRepo TokenRepository
}

// NewAuthService returns a new AuthServiceImpl.
func NewAuthService(userRepo UserRepository, tokenRepo TokenRepository) *AuthServiceImpl {
	return &AuthServiceImpl{userRepo: userRepo, tokenRepo: tokenRepo}
}

// Authenticate verifies credentials and issues a 24-hour authentication token.
func (s *AuthServiceImpl) Authenticate(ctx context.Context, email, password string) (*domain.Token, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	match, err := user.Password.Matches(password)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, domain.ErrInvalidToken
	}

	if !user.Activated {
		return nil, domain.ErrInvalidToken
	}

	token, err := domain.NewToken(user.ID, 24*time.Hour, domain.ScopeAuthentication)
	if err != nil {
		return nil, err
	}

	if err := s.tokenRepo.Insert(ctx, token); err != nil {
		return nil, err
	}

	return token, nil
}

// GeneratePasswordResetToken creates a 30-minute password-reset token for the user
// with the given email. If the email is not registered, it returns nil, nil to
// prevent user enumeration.
func (s *AuthServiceImpl) GeneratePasswordResetToken(ctx context.Context, email string) (*domain.Token, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Do not reveal whether the email exists.
		return nil, nil //nolint:nilerr
	}

	if !user.Activated {
		return nil, nil
	}

	token, err := domain.NewToken(user.ID, 30*time.Minute, domain.ScopePasswordReset)
	if err != nil {
		return nil, err
	}

	if err := s.tokenRepo.Insert(ctx, token); err != nil {
		return nil, err
	}

	return token, nil
}
