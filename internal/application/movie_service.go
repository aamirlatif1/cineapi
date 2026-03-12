package application

import (
	"context"

	"github.com/aamir-al/cineapi/internal/domain"
)

// MovieServiceImpl implements the movie use cases.
type MovieServiceImpl struct {
	repo MovieRepository
}

// NewMovieService returns a new MovieServiceImpl.
func NewMovieService(repo MovieRepository) *MovieServiceImpl {
	return &MovieServiceImpl{repo: repo}
}

// GetMovie retrieves a single movie by id.
func (s *MovieServiceImpl) GetMovie(ctx context.Context, id int64) (*domain.Movie, error) {
	return s.repo.Get(ctx, id)
}

// ListMovies returns a paginated, filtered list of movies.
func (s *MovieServiceImpl) ListMovies(ctx context.Context, filters MovieFilters) ([]*domain.Movie, Metadata, error) {
	return s.repo.GetAll(ctx, filters)
}

// CreateMovie validates and persists a new movie.
func (s *MovieServiceImpl) CreateMovie(ctx context.Context, title string, year, runtime int32, genres []string) (*domain.Movie, error) {
	movie, validationErrors := domain.NewMovie(title, year, runtime, genres)
	if validationErrors != nil {
		return nil, &ValidationError{Fields: validationErrors}
	}
	if err := s.repo.Insert(ctx, movie); err != nil {
		return nil, err
	}
	return movie, nil
}

// UpdateMovie persists changes to an existing movie.
func (s *MovieServiceImpl) UpdateMovie(ctx context.Context, movie *domain.Movie) error {
	return s.repo.Update(ctx, movie)
}

// DeleteMovie removes a movie by id.
func (s *MovieServiceImpl) DeleteMovie(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// ----------------------------------------
// ValidationError
// ----------------------------------------

// ValidationError carries per-field validation messages.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return "validation failed"
}
