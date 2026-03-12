package application

import (
	"context"
	"time"

	"github.com/aamir-al/cineapi/internal/domain"
)

// UserServiceImpl implements user management use cases.
type UserServiceImpl struct {
	userRepo  UserRepository
	tokenRepo TokenRepository
}

// NewUserService returns a new UserServiceImpl.
func NewUserService(userRepo UserRepository, tokenRepo TokenRepository) *UserServiceImpl {
	return &UserServiceImpl{userRepo: userRepo, tokenRepo: tokenRepo}
}

// RegisterUser creates a new inactive user and an activation token.
func (s *UserServiceImpl) RegisterUser(ctx context.Context, name, email, password string) (*domain.User, *domain.Token, error) {
	user, valErrs := domain.NewUser(name, email, password)
	if valErrs != nil {
		return nil, nil, &ValidationError{Fields: valErrs}
	}

	if err := s.userRepo.Insert(ctx, user); err != nil {
		return nil, nil, err
	}

	token, err := domain.NewToken(user.ID, 3*24*time.Hour, domain.ScopeActivation)
	if err != nil {
		return nil, nil, err
	}

	if err := s.tokenRepo.Insert(ctx, token); err != nil {
		return nil, nil, err
	}

	return user, token, nil
}

// ActivateUser activates a user account using the provided activation token.
func (s *UserServiceImpl) ActivateUser(ctx context.Context, tokenPlaintext string) (*domain.User, error) {
	user, err := s.tokenRepo.GetForUser(ctx, domain.ScopeActivation, tokenPlaintext)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	user.Activated = true
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	if err := s.tokenRepo.DeleteAllForUser(ctx, domain.ScopeActivation, user.ID); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdatePassword resets a user's password using a password-reset token, then
// invalidates all existing authentication tokens for that user.
func (s *UserServiceImpl) UpdatePassword(ctx context.Context, tokenPlaintext, newPassword string) error {
	user, err := s.tokenRepo.GetForUser(ctx, domain.ScopePasswordReset, tokenPlaintext)
	if err != nil {
		return domain.ErrInvalidToken
	}

	if err := user.Password.Set(newPassword); err != nil {
		return err
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Invalidate all authentication tokens
	if err := s.tokenRepo.DeleteAllForUser(ctx, domain.ScopeAuthentication, user.ID); err != nil {
		return err
	}

	// Consume the password-reset token
	return s.tokenRepo.DeleteAllForUser(ctx, domain.ScopePasswordReset, user.ID)
}
