package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"

	"github.com/aamir-al/cineapi/internal/application"
	"github.com/aamir-al/cineapi/internal/domain"
)

// ----------------------------------------
// Service interfaces (driving ports consumed by HTTP handlers)
// ----------------------------------------

// MovieService defines the movie use cases required by the HTTP adapter.
type MovieService interface {
	GetMovie(ctx context.Context, id int64) (*domain.Movie, error)
	ListMovies(ctx context.Context, filters application.MovieFilters) ([]*domain.Movie, application.Metadata, error)
	CreateMovie(ctx context.Context, title string, year, runtime int32, genres []string) (*domain.Movie, error)
	UpdateMovie(ctx context.Context, movie *domain.Movie) error
	DeleteMovie(ctx context.Context, id int64) error
}

// UserService defines the user use cases required by the HTTP adapter.
type UserService interface {
	RegisterUser(ctx context.Context, name, email, password string) (*domain.User, *domain.Token, error)
	ActivateUser(ctx context.Context, tokenPlaintext string) (*domain.User, error)
	UpdatePassword(ctx context.Context, tokenPlaintext, newPassword string) error
}

// AuthService defines the authentication use cases required by the HTTP adapter.
type AuthService interface {
	Authenticate(ctx context.Context, email, password string) (*domain.Token, error)
	GeneratePasswordResetToken(ctx context.Context, email string) (*domain.Token, error)
}

// ----------------------------------------
// Config and Application struct
// ----------------------------------------

// Config carries runtime configuration for the application.
type Config struct {
	Port    int
	Env     string
	Version string
}

// Application holds all shared dependencies for the HTTP adapter.
// Wired in cmd/api/main.go.
type Application struct {
	config    Config
	logger    *slog.Logger
	db        *gorm.DB
	tokenRepo application.TokenRepository
	movieSvc  MovieService
	userSvc   UserService
	authSvc   AuthService
	startTime time.Time
}

// NewApplication constructs an Application with the provided dependencies.
func NewApplication(
	cfg Config,
	logger *slog.Logger,
	db *gorm.DB,
	tokenRepo application.TokenRepository,
	movieSvc MovieService,
	userSvc UserService,
	authSvc AuthService,
) *Application {
	return &Application{
		config:    cfg,
		logger:    logger,
		db:        db,
		tokenRepo: tokenRepo,
		movieSvc:  movieSvc,
		userSvc:   userSvc,
		authSvc:   authSvc,
		startTime: time.Now(),
	}
}

// ----------------------------------------
// Router
// ----------------------------------------

// NewRouter builds and returns the chi router with all middleware and routes attached.
func NewRouter(app *Application) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.RealIP)
	r.Use(app.requestID)
	r.Use(app.recoverPanic)
	r.Use(app.logRequest)
	r.Use(app.authenticate)

	r.NotFound(app.notFoundResponse)
	r.MethodNotAllowed(app.methodNotAllowedResponse)

	app.registerRoutes(r)

	return r
}
