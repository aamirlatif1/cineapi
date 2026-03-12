package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/aamir-al/cineapi/internal/application"
	"github.com/aamir-al/cineapi/internal/domain"
)

// movieRepo implements application.MovieRepository using GORM.
type movieRepo struct{ db *gorm.DB }

// NewMovieRepo returns a new MovieRepository backed by db.
func NewMovieRepo(db *gorm.DB) application.MovieRepository {
	return &movieRepo{db: db}
}

// Insert persists a new movie and populates m.ID, m.CreatedAt, and m.Version.
func (r *movieRepo) Insert(ctx context.Context, m *domain.Movie) error {
	model := movieFromDomain(m)
	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return result.Error
	}
	m.ID = model.ID
	m.CreatedAt = model.CreatedAt
	m.Version = model.Version
	return nil
}

// Get retrieves a movie by id.
func (r *movieRepo) Get(ctx context.Context, id int64) (*domain.Movie, error) {
	var model MovieModel
	result := r.db.WithContext(ctx).First(&model, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return model.toDomain(), nil
}

// GetAll returns a paginated, filtered, sorted list of movies plus metadata.
func (r *movieRepo) GetAll(ctx context.Context, f application.MovieFilters) ([]*domain.Movie, application.Metadata, error) {
	// Validate and default pagination
	page := f.Page
	if page < 1 {
		page = 1
	}
	pageSize := f.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	// Validate sort field
	sortColumn, sortDir := parseSortField(f.Sort)

	query := r.db.WithContext(ctx).Model(&MovieModel{})

	if f.Title != "" {
		query = query.Where("title ILIKE ?", "%"+f.Title+"%")
	}
	if len(f.Genres) > 0 {
		query = query.Where("genres @> ?", formatGenresArray(f.Genres))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, application.Metadata{}, err
	}

	var models []MovieModel
	result := query.
		Order(fmt.Sprintf("%s %s", sortColumn, sortDir)).
		Limit(pageSize).
		Offset(offset).
		Find(&models)
	if result.Error != nil {
		return nil, application.Metadata{}, result.Error
	}

	movies := make([]*domain.Movie, len(models))
	for i := range models {
		movies[i] = models[i].toDomain()
	}

	lastPage := int((total + int64(pageSize) - 1) / int64(pageSize))
	if lastPage < 1 {
		lastPage = 1
	}
	meta := application.Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     lastPage,
		TotalRecords: int(total),
	}

	return movies, meta, nil
}

// Update performs an optimistic-locking update on a movie.
func (r *movieRepo) Update(ctx context.Context, m *domain.Movie) error {
	result := r.db.WithContext(ctx).
		Model(&MovieModel{}).
		Where("id = ? AND version = ?", m.ID, m.Version).
		Updates(map[string]any{
			"title":   m.Title,
			"year":    m.Year,
			"runtime": m.Runtime,
			"genres":  formatGenresArray(m.Genres),
			"version": gorm.Expr("version + 1"),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrEditConflict
	}
	m.Version++
	return nil
}

// Delete removes a movie by id.
func (r *movieRepo) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&MovieModel{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ----------------------------------------
// helpers
// ----------------------------------------

var allowedSortFields = map[string]string{
	"id": "id", "-id": "id",
	"title": "title", "-title": "title",
	"year": "year", "-year": "year",
	"runtime": "runtime", "-runtime": "runtime",
}

func parseSortField(sort string) (column, direction string) {
	if sort == "" {
		return "id", "ASC"
	}
	if strings.HasPrefix(sort, "-") {
		col, ok := allowedSortFields[sort]
		if !ok {
			return "id", "ASC"
		}
		return col, "DESC"
	}
	col, ok := allowedSortFields[sort]
	if !ok {
		return "id", "ASC"
	}
	return col, "ASC"
}

// formatGenresArray returns a PostgreSQL array literal, e.g. {"drama","crime"}.
func formatGenresArray(genres []string) string {
	quoted := make([]string, len(genres))
	for i, g := range genres {
		quoted[i] = `"` + strings.ReplaceAll(g, `"`, `\"`) + `"`
	}
	return "{" + strings.Join(quoted, ",") + "}"
}
