# cineapi Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-03-12

## Active Technologies

- Go 1.22+ (use latest stable) + `chi` (HTTP router), `gorm.io/gorm` + `gorm.io/driver/postgres` (ORM), `golang-migrate` (migrations), `golang.org/x/crypto` (bcrypt) (001-movie-db-api)

## Project Structure

```text
cmd/api/              # main package — wires adapters and starts server
internal/domain/      # entities, value objects, domain errors (no external deps)
internal/application/ # use cases and port interfaces
internal/adapters/http/          # chi HTTP handlers and middleware
internal/adapters/repository/postgres/   # GORM PostgreSQL repository implementations
pkg/validator/        # generic input validator
migrations/           # SQL migration files (future PostgreSQL adapter)
```

## Commands

```bash
make build        # go build -o bin/api ./cmd/api
make run          # go run ./cmd/api -env=development -port=4000
make test         # go test ./...
make lint         # golangci-lint run
make tidy         # go mod tidy
make db/up        # docker compose up -d (start PostgreSQL)
make migrate/up   # apply all pending golang-migrate migrations
make migrate/down # roll back last migration
```

## Code Style

- Follow Effective Go and Go Code Review Comments
- Hexagonal Architecture: dependency direction adapters → application → domain
- Errors as values; no panic outside main/init
- context.Context as first param on all I/O-crossing functions
- gofmt + goimports applied; zero golangci-lint errors
- Table-driven tests preferred; no shared mutable global state in tests

## Recent Changes

- 001-movie-db-api: Go + chi + GORM + PostgreSQL + golang-migrate + bcrypt

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
