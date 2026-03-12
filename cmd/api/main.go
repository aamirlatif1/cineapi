package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	httpAdapter "github.com/aamir-al/cineapi/internal/adapters/http"
	"github.com/aamir-al/cineapi/internal/adapters/repository/postgres"
	"github.com/aamir-al/cineapi/internal/application"
)

const version = "1.0.0"

func main() {
	// Load .env file if present (development convenience; ignored in production)
	_ = godotenv.Load()

	var cfg httpAdapter.Config
	cfg.Version = version

	flag.IntVar(&cfg.Port, "port", getEnvInt("PORT", 4000), "API server port")
	flag.StringVar(&cfg.Env, "env", getEnv("ENV", "development"), "Environment (development|production)")

	dbDSN := flag.String("db-dsn", getEnv("DATABASE_URL", ""), "PostgreSQL DSN")
	flag.Parse()

	// Logger
	var handler slog.Handler
	if cfg.Env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}
	logger := slog.New(handler)

	// Database
	db, err := postgres.Open(*dbDSN)
	if err != nil {
		logger.Error("failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()
	logger.Info("database connection pool established")

	// Register expvar metrics
	httpAdapter.RegisterMetrics()

	// Repositories
	movieRepo := postgres.NewMovieRepo(db)
	userRepo := postgres.NewUserRepo(db)
	tokenRepo := postgres.NewTokenRepo(db)

	// Services
	movieSvc := application.NewMovieService(movieRepo)
	userSvc := application.NewUserService(userRepo, tokenRepo)
	authSvc := application.NewAuthService(userRepo, tokenRepo)

	// Seed demo user
	if err := postgres.SeedDemoUser(db); err != nil {
		logger.Warn("demo user seed failed", slog.Any("error", err))
	}

	// HTTP application
	app := httpAdapter.NewApplication(cfg, logger, db, tokenRepo, movieSvc, userSvc, authSvc)

	if err := serve(app, cfg, logger); err != nil {
		logger.Error("server error", slog.Any("error", err))
		os.Exit(1)
	}
}

func serve(app *httpAdapter.Application, cfg httpAdapter.Config, logger *slog.Logger) error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      httpAdapter.NewRouter(app),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	shutdownErr := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		logger.Info("shutting down server", slog.String("signal", "caught"))
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		shutdownErr <- srv.Shutdown(ctx)
	}()

	logger.Info("starting server",
		slog.String("addr", srv.Addr),
		slog.String("env", cfg.Env),
		slog.String("version", cfg.Version),
	)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return <-shutdownErr
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		var i int
		if _, err := fmt.Sscanf(v, "%d", &i); err == nil {
			return i
		}
	}
	return fallback
}
