package domain

import (
	"time"

	"github.com/aamir-al/cineapi/pkg/validator"
)

// Movie represents a film entry in the catalog.
type Movie struct {
	ID        int64
	Title     string
	Year      int32
	Runtime   int32
	Genres    []string
	CreatedAt time.Time
	Version   int32
}

// NewMovie creates a validated Movie. It returns a non-nil validation error map
// when any field fails its rule.
func NewMovie(title string, year, runtime int32, genres []string) (*Movie, map[string]string) {
	v := validator.New()

	v.Check(title != "", "title", "must be provided")
	v.Check(len(title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(year != 0, "year", "must be provided")
	v.Check(year >= 1888, "year", "must be greater than 1888")
	v.Check(year <= int32(time.Now().Year()+1), "year", "must not be in the future")

	v.Check(runtime != 0, "runtime", "must be provided")
	v.Check(runtime > 0, "runtime", "must be a positive integer")

	v.Check(genres != nil, "genres", "must be provided")
	v.Check(len(genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(genres), "genres", "must not contain duplicate values")

	if !v.Valid() {
		return nil, v.Errors()
	}

	return &Movie{
		Title:   title,
		Year:    year,
		Runtime: runtime,
		Genres:  genres,
		Version: 1,
	}, nil
}
