<!--
  SYNC IMPACT REPORT
  ==================
  Version change: [TEMPLATE] → 1.0.0
  Modified principles: All placeholders replaced (initial ratification).
  Added sections:
    - Core Principles (I–V)
    - Technology Stack
    - Development Workflow
    - Governance
  Removed sections: None (template placeholders removed).
  Templates reviewed:
    - .specify/templates/plan-template.md ✅ no changes required; Constitution Check
      section already uses runtime-fill language.
    - .specify/templates/spec-template.md ✅ no changes required; structure is
      technology-agnostic.
    - .specify/templates/tasks-template.md ✅ no changes required; path conventions
      are advisory and overridden by plan.md.
  Follow-up TODOs: None — all fields resolved.
-->

# CineAPI Constitution

## Core Principles

### I. Hexagonal Architecture (Ports & Adapters)

The business logic MUST reside exclusively in the **domain** and **application** layers.
Infrastructure concerns (HTTP, database, messaging, external APIs) MUST be implemented
as **adapters** that satisfy **port interfaces** defined by the application layer.

- The `internal/domain` package MUST NOT import any adapter or framework package.
- The `internal/application` package defines port interfaces (e.g., `Repository`,
  `EventPublisher`) and use-case structs that depend only on those interfaces.
- Adapters live under `internal/adapters/` and MUST NOT contain business logic.
- Dependency direction: `adapters` → `application` → `domain`. Never reversed.

**Rationale**: Isolating the core from infrastructure makes business rules independently
testable and the system replaceable at any adapter boundary without touching core logic.

### II. Idiomatic Go

All code MUST follow [Effective Go](https://go.dev/doc/effective_go) and the
[Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

- Interfaces MUST be small (1–3 methods); define them at the consumer, not the producer.
- Errors MUST be returned as values; `panic` is forbidden outside `main` and `init`.
- `context.Context` MUST be the first parameter of every function that performs I/O
  or crosses a layer boundary.
- No use of `init()` outside test helpers.
- Exported names MUST have doc comments; unexported names are commented only when
  the logic is non-obvious.
- `gofmt` and `goimports` MUST be applied; no formatting deviations are accepted.

**Rationale**: Idiomatic Go maximises readability, toolability, and long-term
maintainability for any Go engineer joining the project.

### III. Clean Domain Model

Domain types MUST be pure Go structs with no framework annotations, ORM tags, or
HTTP binding decorators.

- Persistence tags (`db:""`, `json:""`) MUST live on **adapter-layer DTOs**, not on
  domain entities.
- Business invariants MUST be enforced in domain constructors or methods (e.g.,
  `NewMovie(...)` returns `(Movie, error)`).
- Domain packages MUST be importable and testable with zero external dependencies
  beyond the Go standard library.

**Rationale**: A clean domain model ensures the core can be reasoned about, tested,
and evolved independently of any persistence or transport technology.

### IV. Test Discipline

- Domain and application layers MUST have unit tests with ≥ 80 % statement coverage.
- Adapter implementations (HTTP handlers, repository) MUST have integration tests that
  run against real infrastructure (test database, HTTP test server).
- Table-driven tests (`t.Run` + slice of cases) are the PREFERRED pattern for
  unit tests with multiple inputs.
- Tests MUST NOT share mutable global state; each test MUST be independently runnable
  via `go test -run <name>`.
- Test doubles MUST be hand-rolled interfaces or generated via `mockery`; no
  reflection-heavy mocking frameworks.

**Rationale**: A disciplined test suite catches regressions at the correct layer and
keeps the feedback loop fast.

### V. Observability

Every running service instance MUST be observable with zero additional tooling beyond
standard HTTP calls.

- All log output MUST use `log/slog` with JSON format in production (`-env=prod`).
  Human-readable text format is acceptable in local development.
- Every HTTP handler MUST propagate and log a `request_id` (from header or generated).
- The service MUST expose `GET /health` (liveness) and `GET /readyz` (readiness)
  endpoints returning `200 OK` with a JSON body.
- OpenTelemetry trace context (`traceparent` header) MUST be extracted and forwarded
  on all outbound calls.

**Rationale**: Structured, consistent observability is required for debugging production
incidents and is non-negotiable from day one.

## Technology Stack

| Concern | Choice | Notes |
|---|---|---|
| Language | Go 1.25+ | Minimum version; use the latest stable release. |
| HTTP Router | `net/http` + `chi` | No full-stack frameworks (e.g., Gin is acceptable only if `chi` is insufficient). |
| Database | PostgreSQL via `pgx/v5` | Direct driver; no ORM. |
| Migrations | `golang-migrate` | SQL migration files under `migrations/`. |
| Configuration | Environment variables | `godotenv` for local `.env`; never hard-code secrets. |
| Linting | `golangci-lint` | Config in `.golangci.yml`; zero lint errors on CI. |
| Containerisation | Docker + Docker Compose | `Dockerfile` at repo root; `compose.yml` for local stack. |

### Standard Project Layout

```text
cmd/
  api/              # main package — wires adapters and starts server
internal/
  domain/           # entities, value objects, domain services (no external deps)
  application/      # use cases, port interfaces
  adapters/
    http/           # HTTP handlers, middleware, router setup (driving adapter)
    repository/     # PostgreSQL implementations (driven adapter)
migrations/         # SQL migration files (up/down)
pkg/                # Shared utilities with no business logic (optional)
```

The layout MUST follow this structure. Deviations require a justified entry in the
Complexity Tracking table of `plan.md`.

## Development Workflow

- All changes MUST be submitted via a pull request; direct pushes to `main` are
  forbidden.
- `make test` (`go test ./...`) MUST pass with no failures before merge.
- `golangci-lint run` MUST report zero errors before merge.
- API surface MUST be documented in OpenAPI 3.x (`docs/openapi.yaml`); the spec MUST
  be kept in sync with handler implementations.
- Secrets and credentials MUST NOT be committed; `.env` files are git-ignored.
- Each PR MUST include a brief description of which Constitution principles the change
  touches (or confirms it is unaffected).

## Governance

This constitution supersedes all other project guidelines and verbal agreements.

- **Amendment procedure**: Open a PR that modifies this file, increments the version
  per the semantic versioning policy below, and includes a migration note for any
  breaking change.
- **Versioning policy**:
  - MAJOR — backward-incompatible removal or redefinition of a principle.
  - MINOR — new principle or section added, or material expansion of existing guidance.
  - PATCH — clarifications, wording improvements, typo fixes.
- **Compliance review**: Every `plan.md` MUST complete the Constitution Check gate
  before Phase 0 research begins and again after Phase 1 design. Violations MUST be
  justified in the Complexity Tracking table.
- All complexity beyond the standard layout MUST be justified. Favour simplicity;
  YAGNI applies.

**Version**: 1.0.0 | **Ratified**: 2026-03-12 | **Last Amended**: 2026-03-12
