# Research: CineAPI — Movie Database REST API

**Feature**: `001-movie-db-api`
**Date**: 2026-03-12
**Status**: Complete — no NEEDS CLARIFICATION items remain

---

## Decision 0: Go Version — 1.25+

**Decision**: Go 1.25+ (use latest stable at implementation time).

**Rationale**: Constitution mandates Go 1.25+. `log/slog` (1.21), `net/http` routing
improvements (1.22), and `slices`/`maps` stdlib packages (1.21+) are all available.

---

## Decision 1: HTTP Router — `chi`

**Decision**: Use `github.com/go-chi/chi/v5` as the HTTP router.

**Rationale**: `chi` is a lightweight, idiomatic Go router built on `net/http`. It
supports URL parameters (`:id`), middleware composition via `chi.Use`, and grouped
routes with shared middleware (e.g., `r.Route("/v1/movies", ...)`). It does not impose
any framework structure on the application and composes naturally with Hexagonal
Architecture — handlers are plain functions that accept `http.ResponseWriter` and
`*http.Request`.

**Alternatives considered**:
- `gorilla/mux` — heavier, maintenance burden reduced; chi is the modern successor.
- `net/http` stdlib multiplexer — lacks URL parameter extraction and middleware chain.
- `gin` — too opinionated; binds to its own context type, conflicting with idiomatic Go.

---

## Decision 2: Storage — PostgreSQL via GORM + golang-migrate

**Decision**: Use `gorm.io/gorm` with `gorm.io/driver/postgres` as the ORM/driver for
all database operations. Schema managed by `golang-migrate/migrate` with plain SQL
migration files under `migrations/`.

**Rationale**: Developer preference for GORM's ergonomic query API over raw SQL with
`pgx/v5`. GORM is confined entirely to the adapter layer (`internal/adapters/repository/postgres/`).
Domain entities are plain Go structs with zero GORM tags; the adapter maintains separate
GORM model structs (DTOs) and maps between them. This preserves Hexagonal Architecture
and Principle III (Clean Domain Model) while accepting the ORM deviation documented in
the Complexity Tracking table.

`golang-migrate` is chosen over GORM's `AutoMigrate` because SQL migration files are
version-controlled, reviewable, reversible, and safe to run in production without
surprises.

**Optimistic locking with GORM**:
```go
result := r.db.WithContext(ctx).
    Model(&MovieModel{}).
    Where("id = ? AND version = ?", movie.ID, movie.Version).
    Updates(map[string]any{"title": movie.Title, ..., "version": gorm.Expr("version + 1")})
if result.RowsAffected == 0 {
    return domain.ErrEditConflict
}
```

**Local dev**: `compose.yml` spins up a PostgreSQL 16 container. `DATABASE_URL` env var
configures the DSN (`postgres://user:pass@localhost:5432/cineapi?sslmode=disable`).

**Alternatives considered**:
- `pgx/v5` direct driver — constitution default; rejected per developer preference.
  Easily restored by replacing the adapter package only.
- `sqlc` — type-safe query generation; adds codegen step not needed for this scope.
- GORM `AutoMigrate` — not used; SQL migration files preferred for auditability.

---

## Decision 3: Authentication — Bearer Token (SHA-256 + bcrypt)

**Decision**:
- **Passwords**: hashed with `bcrypt` (cost 12) via `golang.org/x/crypto/bcrypt`.
- **Tokens** (activation, authentication, password-reset): `crypto/rand` generates
  16 random bytes encoded as base32 (26-char plaintext returned to client once).
  The stored form is the SHA-256 hash of the plaintext.

**Rationale**: bcrypt is the standard for password storage in Go. Token hashing with
SHA-256 means the plaintext is never stored; even a complete database dump yields no
usable tokens. Base32 encoding produces URL-safe tokens without padding issues.

**Token scopes and expiry**:
| Scope | Expiry |
|---|---|
| `activation` | 3 days |
| `authentication` | 24 hours |
| `password-reset` | 30 minutes |

**Alternatives considered**:
- JWT — stateless but requires key management; for a demo API, stateful tokens in the
  repository are simpler and revocable.
- UUID tokens — lower entropy than 16 random bytes; `crypto/rand` is preferable.

---

## Decision 3b: Health Endpoints — `/health`, `/readyz`, `/v1/healthcheck`

**Decision**: Expose three observability endpoints:
- `GET /health` — liveness probe (constitution-required); returns `{"status":"ok"}`.
  Responds `200` as long as the process is running.
- `GET /readyz` — readiness probe (constitution-required); returns `200` only when
  the DB connection pool is healthy; returns `503` if the DB ping fails.
- `GET /v1/healthcheck` — existing spec endpoint; returns service status + version.

**Rationale**: `/health` and `/readyz` are non-negotiable per Constitution Principle V.
`/v1/healthcheck` is retained because it is part of the public API contract (spec
FR-019) and returns richer version information for API consumers.

---

## Decision 3c: Configuration — `godotenv`

**Decision**: Load `github.com/joho/godotenv` in `cmd/api/main.go` before flag parsing.
Committed `.env.example` documents all required variables; `.env` is git-ignored.

```
DATABASE_URL=postgres://cineapi:cineapi@localhost:5432/cineapi?sslmode=disable
PORT=4000
ENV=development
```

**Rationale**: Constitution mandates `godotenv` for local `.env` loading. Keeps secrets
out of source control while making local setup a single `cp .env.example .env` step.

---

## Decision 4: Structured Logging — `log/slog`

**Decision**: Use the stdlib `log/slog` package (available since Go 1.21). JSON handler
in production, text handler in development, controlled by `LOG_FORMAT` env var.

**Rationale**: `slog` is idiomatic, zero-dependency, and supports structured key-value
logging. Every request log entry includes: `method`, `path`, `status`, `duration_ms`,
`request_id`.

**Alternatives considered**:
- `zerolog` / `zap` — high-performance but external dependencies; unnecessary overhead
  for a demo service.
- `log` (old stdlib) — unstructured; not acceptable per constitution Principle V.

---

## Decision 5: Runtime Metrics — `expvar`

**Decision**: Expose `GET /debug/vars` using the stdlib `expvar` package. Register
custom counters for: `total_requests_received`, `total_responses_sent`,
`total_processing_time_μs`.

**Rationale**: `expvar` is zero-dependency, thread-safe, and JSON-formatted. It
integrates directly with `net/http` via `expvar.Handler()`. Satisfies spec FR-020
without any external observability infrastructure.

**Alternatives considered**:
- Prometheus metrics — production-grade but adds a dependency; out of scope for demo.

---

## Decision 6: Input Validation

**Decision**: Hand-rolled validator in `pkg/validator` accumulating per-field errors
into a `map[string]string`. All validation runs in the domain constructor or in the
HTTP layer before calling the application service.

**Rationale**: Go's idiomatic approach. Avoids reflection-heavy validation libraries.
Keeps validation logic close to the entity definition.

**Validation rules summary**:

| Field | Rules |
|---|---|
| Movie.Title | required, 1–500 chars |
| Movie.Year | 1888 – current year + 1 |
| Movie.Runtime | > 0 |
| Movie.Genres | 1–5 entries, unique values |
| User.Name | required, 1–500 chars |
| User.Email | required, valid RFC 5322, ≤500 chars |
| User.Password | 8–72 chars (72 = bcrypt limit) |

---

## Decision 7: Pagination and Filtering

**Decision**: Query parameters for list endpoints:
- `?page=1&page_size=20` (defaults; page_size max 100)
- `?title=godfather` (case-insensitive substring match)
- `?genres=drama,crime` (all specified genres must be present)
- `?sort=year` / `?sort=-year` (prefix `-` = descending; allowed: `id`, `title`, `year`, `runtime`)

Pagination metadata returned in response envelope:
```json
{
  "metadata": {
    "current_page": 1,
    "page_size": 20,
    "first_page": 1,
    "last_page": 3,
    "total_records": 47
  }
}
```

**Rationale**: Offset-based pagination is simple for an in-memory store. The filter/sort
logic is implemented inside the in-memory repository's `GetAll` method.

---

## Decision 8: Sample Data

**Decision**: Seed 20 movies covering diverse genres (action, drama, comedy, sci-fi,
thriller, animation) and decades (1970s–2020s). One demo user (`demo@cineapi.local`,
password `pa55word`) is seeded in activated state for immediate API exploration.

**Sample movies** (representative):
1. The Godfather (1972, Drama, Crime) — 175 min
2. The Shawshank Redemption (1994, Drama) — 142 min
3. Pulp Fiction (1994, Crime, Drama) — 154 min
4. Schindler's List (1993, Drama, History) — 195 min
5. The Dark Knight (2008, Action, Crime) — 152 min
6. Forrest Gump (1994, Drama, Romance) — 142 min
7. Inception (2010, Action, Sci-Fi) — 148 min
8. Interstellar (2014, Adventure, Sci-Fi) — 169 min
9. The Matrix (1999, Action, Sci-Fi) — 136 min
10. Goodfellas (1990, Crime, Drama) — 146 min
11. Fight Club (1999, Drama, Thriller) — 139 min
12. The Silence of the Lambs (1991, Crime, Thriller) — 118 min
13. Parasite (2019, Drama, Thriller) — 132 min
14. Whiplash (2014, Drama, Music) — 107 min
15. The Grand Budapest Hotel (2014, Adventure, Comedy) — 99 min
16. La La Land (2016, Drama, Music, Romance) — 128 min
17. Get Out (2017, Horror, Thriller) — 104 min
18. 1917 (2019, Drama, War) — 119 min
19. Dune (2021, Adventure, Sci-Fi) — 155 min
20. Everything Everywhere All at Once (2022, Action, Comedy, Sci-Fi) — 139 min
