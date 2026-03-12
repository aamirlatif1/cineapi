# Implementation Plan: CineAPI — Movie Database REST API

**Branch**: `001-movie-db-api` | **Date**: 2026-03-12 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-movie-db-api/spec.md`

## Summary

CineAPI is a JSON-based RESTful movie database service built in Go. It exposes 12 HTTP
endpoints covering movie CRUD, user registration and activation, token-based
authentication, password recovery, health-check, and runtime metrics. The application
follows Hexagonal Architecture with a PostgreSQL repository adapter (GORM ORM,
golang-migrate for schema migrations) pre-loaded with sample movie data, and uses the
`chi` router for HTTP routing.

## Technical Context

**Language/Version**: Go 1.25+ (use latest stable)
**Primary Dependencies**: `chi` (HTTP router), `gorm.io/gorm` + `gorm.io/driver/postgres`
  (ORM + PostgreSQL driver), `golang-migrate/migrate` (schema migrations),
  `golang.org/x/crypto` (bcrypt), `github.com/joho/godotenv` (local .env loading),
  `log/slog` (stdlib), `expvar` (stdlib)
**Storage**: PostgreSQL (via GORM); Docker Compose provides local DB instance
**Testing**: `go test ./...` with `net/http/httptest` for handler integration tests;
  table-driven unit tests for domain and application layers; test DB via Docker
**Target Platform**: Linux server / local development with Docker Compose
**Project Type**: web-service / REST API
**Performance Goals**: Read endpoints ≤ 500 ms p95; healthcheck ≤ 100 ms p99
**Constraints**: Requires PostgreSQL (via Docker Compose for local dev); data persists
  across restarts; migrations applied manually via `make migrate/up` before first run;
  sample data seeded in migration 000004
**Scale/Scope**: Demo/reference implementation; single instance; ~20 sample movies

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|---|---|---|
| I. Hexagonal Architecture | ✅ PASS | Domain/application layers free of HTTP or storage concerns. PostgreSQL adapter is a driven adapter implementing Repository ports only. |
| II. Idiomatic Go | ✅ PASS | Small interfaces, errors as values, `context.Context` propagation, no panics in library code. |
| III. Clean Domain Model | ✅ PASS | Domain structs carry no ORM/JSON tags. Adapter-layer DTOs (GORM models) live only in `internal/adapters/repository/postgres/`. Business invariants enforced in domain constructors. |
| IV. Test Discipline | ✅ PASS | Domain/application layers unit-tested; HTTP handlers tested with `httptest`; repository integration tests run against real test DB. |
| V. Observability | ✅ PASS | `log/slog` JSON logging; `request_id` middleware; `/health` (liveness), `/readyz` (readiness), `/v1/healthcheck`, and `/debug/vars` endpoints all exposed. |
| Tech Stack — Go version | ✅ PASS | Go 1.25+. |
| Tech Stack — Storage | ✅ PASS | PostgreSQL + `golang-migrate`. |
| Tech Stack — Config | ✅ PASS | `godotenv` + env vars; `.env` git-ignored; `.env.example` committed. |
| Tech Stack — OpenAPI | ✅ PASS | Spec lives at `docs/openapi.yaml` per constitution. |
| Tech Stack — ORM | ⚠️ DEVIATION | Constitution: `pgx/v5` direct driver, no ORM. Plan uses GORM. See Complexity Tracking and **Better Options** below. |

**Post-Phase-1 re-check**: 9 of 10 checks pass. One active deviation (GORM) is
documented with alternatives in the Complexity Tracking table.

## Project Structure

### Documentation (this feature)

```text
specs/001-movie-db-api/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── openapi.yaml     # Phase 1 output — source of truth (copied to docs/openapi.yaml)
└── tasks.md             # Phase 2 output (/speckit.tasks — NOT created here)
```

### Source Code (repository root)

```text
cmd/
  api/
    main.go              # Entry point: parse config, wire adapters, start server

internal/
  domain/
    movie.go             # Movie entity, NewMovie constructor, Validate()
    user.go              # User entity, password type (bcrypt), NewUser constructor
    token.go             # Token entity, NewToken constructor, scope constants
    errors.go            # Sentinel domain errors (ErrNotFound, ErrDuplicateEmail, …)

  application/
    ports.go             # MovieRepository, UserRepository, TokenRepository interfaces
    movie_service.go     # Use cases: ListMovies, GetMovie, CreateMovie, UpdateMovie, DeleteMovie
    user_service.go      # Use cases: RegisterUser, ActivateUser, UpdatePassword
    auth_service.go      # Use cases: Authenticate, GeneratePasswordResetToken

  adapters/
    http/
      router.go          # chi router, route registration, middleware chain
      middleware.go      # requestID, authenticate, requireActivatedUser, recoverPanic, logRequest
      movies.go          # movie handlers (list, show, create, update, delete)
      users.go           # user handlers (register, activate, updatePassword)
      tokens.go          # token handlers (createAuthToken, createPasswordResetToken)
      health.go          # /health (liveness), /readyz (readiness), /v1/healthcheck, /debug/vars handlers
      helpers.go         # readJSON, writeJSON, errorResponse helpers
      errors.go          # HTTP-layer error constructors (notFoundResponse, etc.)

    repository/
      postgres/
        db.go            # GORM DB setup (Open, connection pool, ping)
        models.go        # GORM model structs (MovieModel, UserModel, TokenModel) — adapter DTOs only
        movie_repo.go    # MovieRepository implementation via GORM
        user_repo.go     # UserRepository implementation via GORM
        token_repo.go    # TokenRepository implementation via GORM
        seed.go          # Seed function: inserts ~20 sample movies + 1 demo user

pkg/
  validator/
    validator.go         # Generic input validator (field-level error accumulation)

migrations/              # golang-migrate SQL files (up/down)
  000001_create_movies_table.up.sql
  000001_create_movies_table.down.sql
  000002_create_users_table.up.sql
  000002_create_users_table.down.sql
  000003_create_tokens_table.up.sql
  000003_create_tokens_table.down.sql
  000004_seed_sample_data.up.sql
  000004_seed_sample_data.down.sql

compose.yml              # PostgreSQL service for local development
.env.example             # Example env vars (committed); .env is git-ignored
docs/
  openapi.yaml           # OpenAPI 3.0 spec (constitution-required location)
Makefile                 # build, run, test, lint, migrate targets
```

**Structure Decision**: Single Go module at repository root following the constitution's
standard Hexagonal layout. The `internal/` boundary enforces that only `cmd/api` wires
adapters; domain and application layers are package-private to the module.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|---|---|---|
| GORM instead of `pgx/v5` | Developer preference; ergonomic CRUD API reduces adapter boilerplate. GORM tags are 100% confined to adapter DTOs in `internal/adapters/repository/postgres/models.go`. Domain entities are tag-free. Swapping to pgx later touches only the adapter package. | See **Better Options** below — two fully constitution-compliant alternatives exist and are recommended over GORM. |

## Better Options

### Option A — `pgx/v5` + hand-written SQL (recommended, fully compliant)

Use `github.com/jackc/pgx/v5` directly with `pgxpool` for connection pooling.
Write SQL queries by hand in the repository adapter. Map `pgx.Rows` → domain entities
in the adapter.

**Pros**: Zero ORM overhead; SQL is explicit and reviewable; fully constitution-compliant;
no extra abstraction layer; `pgx/v5` is the fastest Go PostgreSQL driver.

**Cons**: More boilerplate per query (~5–10 lines of scan code per entity).

```go
// internal/adapters/repository/postgres/movie_repo.go
func (r *movieRepo) Get(ctx context.Context, id int64) (*domain.Movie, error) {
    row := r.pool.QueryRow(ctx,
        `SELECT id, title, year, runtime, genres, created_at, version
         FROM movies WHERE id = $1`, id)
    var m domain.Movie
    err := row.Scan(&m.ID, &m.Title, &m.Year, &m.Runtime,
                    &m.Genres, &m.CreatedAt, &m.Version)
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, domain.ErrNotFound
    }
    return &m, err
}
```

---

### Option B — `sqlc` + `pgx/v5` (recommended for larger teams)

Define SQL queries in `.sql` files; `sqlc` generates type-safe Go functions backed by
`pgx/v5`. No ORM, no hand-written scan code.

```sql
-- queries/movies.sql
-- name: GetMovie :one
SELECT * FROM movies WHERE id = $1;
```

Generates: `func (q *Queries) GetMovie(ctx context.Context, id int64) (Movie, error)`

**Pros**: SQL is the source of truth; generated code is type-safe; no reflection; fully
constitution-compliant; great for teams who want SQL control without scan boilerplate.

**Cons**: Adds a codegen step (`sqlc generate`) to the build pipeline; generated types
are adapter-layer structs (not domain entities) — mapping still needed.

---

### Current Choice — GORM (active deviation)

Retained per developer preference. Acceptable only while GORM structs are confined
entirely to the adapter layer. Any GORM tag appearing on a `domain.*` type is an
immediate Principle III violation and MUST be fixed before merge.
