package http

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/aamir-al/cineapi/internal/application"
	"github.com/aamir-al/cineapi/internal/domain"
	"github.com/aamir-al/cineapi/pkg/validator"
)

// ----------------------------------------
// US1 — Browse and retrieve movies
// ----------------------------------------

func (app *Application) listMovies(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title    string
		Genres   []string
		Page     int
		PageSize int
		Sort     string
	}

	v := validator.New()
	qs := r.URL.Query()

	input.Title = readString(qs, "title", "")
	input.Genres = readCSV(qs, "genres", []string{})
	input.Page = readInt(qs, "page", 1, v)
	input.PageSize = readInt(qs, "page_size", 20, v)
	input.Sort = readString(qs, "sort", "id")

	v.Check(input.Page > 0, "page", "must be greater than zero")
	v.Check(input.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(input.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(input.PageSize <= 100, "page_size", "must be a maximum of 100")
	v.Check(validator.In(input.Sort, "id", "-id", "title", "-title", "year", "-year", "runtime", "-runtime"),
		"sort", "invalid sort value")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors())
		return
	}

	filters := application.MovieFilters{
		Title:    input.Title,
		Genres:   input.Genres,
		Page:     input.Page,
		PageSize: input.PageSize,
		Sort:     input.Sort,
	}

	movies, meta, err := app.movieSvc.ListMovies(r.Context(), filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"movies": movies, "metadata": meta}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) showMovie(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.movieSvc.GetMovie(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// ----------------------------------------
// US2 — Manage movies
// ----------------------------------------

func (app *Application) createMovie(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	movie, err := app.movieSvc.CreateMovie(r.Context(), input.Title, input.Year, input.Runtime, input.Genres)
	if err != nil {
		var ve *application.ValidationError
		if errors.As(err, &ve) {
			app.failedValidationResponse(w, r, ve.Fields)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	if err := app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) updateMovie(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.movieSvc.GetMovie(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Pointer fields detect absent-vs-zero in partial PATCH body
	var input struct {
		Title   *string  `json:"title"`
		Year    *int32   `json:"year"`
		Runtime *int32   `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		movie.Title = *input.Title
	}
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	// Re-validate
	_, valErrs := domain.NewMovie(movie.Title, movie.Year, movie.Runtime, movie.Genres)
	if valErrs != nil {
		app.failedValidationResponse(w, r, valErrs)
		return
	}

	if err := app.movieSvc.UpdateMovie(r.Context(), movie); err != nil {
		switch {
		case errors.Is(err, domain.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) deleteMovie(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	if err := app.movieSvc.DeleteMovie(r.Context(), id); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// ----------------------------------------
// Query-string helpers
// ----------------------------------------

func readIDParam(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("invalid id parameter")
	}
	return id, nil
}

func readString(qs map[string][]string, key, defaultValue string) string {
	if values, ok := qs[key]; ok && len(values) > 0 && values[0] != "" {
		return values[0]
	}
	return defaultValue
}

func readCSV(qs map[string][]string, key string, defaultValue []string) []string {
	if values, ok := qs[key]; ok && len(values) > 0 && values[0] != "" {
		return strings.Split(values[0], ",")
	}
	return defaultValue
}

func readInt(qs map[string][]string, key string, defaultValue int, v *validator.Validator) int {
	s := readString(qs, key, "")
	if s == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}
	return i
}
