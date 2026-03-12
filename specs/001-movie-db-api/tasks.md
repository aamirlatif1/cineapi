---
description: "Task list for CineAPI — Movie Database REST API"
---

# Tasks: CineAPI — Movie Database REST API

**Input**: Design documents from `/specs/001-movie-db-api/`
**Prerequisites**: plan.md ✅ spec.md ✅ data-model.md ✅ contracts/openapi.yaml ✅ research.md ✅

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1–US5)
- File paths are relative to repository root

## Path Conventions

Per `plan.md` Hexagonal layout:
- Domain entities: `internal/domain/`
- Port interfaces: `internal/application/`
- HTTP adapter: `internal/adapters/http/`
- Repository adapter: `internal/adapters/repository/postgres/`
- Shared utilities: `pkg/`
- Entry point: `cmd/api/`

---

## Phase 1: Setup

**Purpose**: Repository scaffolding and tooling configuration.

- [x] T001 Initialise Go module (`go mod init github.com/YOUR_USERNAME/cineapi`) and create all source directories: `cmd/api/`, `internal/domain/`, `internal/application/`, `internal/adapters/http/`, `internal/adapters/repository/postgres/`, `pkg/validator/`, `migrations/`, `docs/`
- [x] T002 Add all required dependencies to `go.mod` via `go get`: `github.com/go-chi/chi/v5`, `gorm.io/gorm`, `gorm.io/driver/postgres`, `github.com/golang-migrate/migrate/v4`, `golang.org/x/crypto`, `github.com/joho/godotenv`, `github.com/lib/pq`
- [x] T003 [P] Create `Makefile` with targets: `build` (`go build -o bin/api ./cmd/api`), `run` (`go run ./cmd/api`), `test` (`go test ./...`), `lint` (`golangci-lint run`), `tidy` (`go mod tidy`), `db/up` (`docker compose up -d`), `db/down` (`docker compose down`), `migrate/up` (`migrate -path ./migrations -database $$DATABASE_URL up`), `migrate/down` (`migrate -path ./migrations -database $$DATABASE_URL down 1`)
- [x] T004 [P] Create `compose.yml` with a `postgres:16-alpine` service: database `cineapi`, user `cineapi`, password `cineapi`, port `5432:5432`, named volume for persistence
- [x] T005 [P] Create `.env.example` with all required variables (`DATABASE_URL`, `PORT=4000`, `ENV=development`); create `.gitignore` ignoring `.env`, `bin/`, `*.test`
- [x] T006 [P] Create `.golangci.yml` enabling linters: `govet`, `errcheck`, `staticcheck`, `gofmt`, `goimports`, `gosimple`, `unused`, `misspell`

---

## Phase 2: Foundational

**Purpose**: Core infrastructure shared by every user story. No user story work begins until this phase is complete.

**⚠️ CRITICAL**: All phases 3–7 depend on this phase.

- [x] T007 Create `internal/domain/errors.go` defining sentinel errors: `ErrNotFound`, `ErrDuplicateEmail`, `ErrEditConflict`, `ErrInvalidToken` using `errors.New`
- [x] T008 [P] Create `pkg/validator/validator.go` with a `Validator` struct holding `map[string]string` errors, and methods: `Check(ok bool, key, message string)`, `Valid() bool`, `Errors() map[string]string`
- [x] T009 Create `internal/application/ports.go` defining interfaces: `MovieRepository` (Insert, Get, GetAll, Update, Delete), `UserRepository` (Insert, GetByEmail, Update), `TokenRepository` (Insert, GetForUser, DeleteAllForUser), plus `MovieFilters` and `Metadata` structs
- [x] T010 Create `internal/adapters/repository/postgres/models.go` with GORM adapter DTOs: `MovieModel`, `UserModel`, `TokenModel` structs with GORM struct tags; add `toDomain()` and `fromDomain()` mapping methods on each model
- [x] T011 Create `internal/adapters/repository/postgres/db.go` with `Open(dsn string) (*gorm.DB, error)` that opens a GORM connection with `gorm.io/driver/postgres`, sets connection pool (`SetMaxOpenConns(25)`, `SetMaxIdleConns(25)`, `SetConnMaxIdleTime(15*time.Minute)`), and pings the DB
- [x] T012 [P] Create migration `migrations/000001_create_movies_table.up.sql` (`id bigserial PK`, `title text NOT NULL`, `year integer NOT NULL`, `runtime integer NOT NULL`, `genres text[] NOT NULL`, `created_at timestamptz DEFAULT now()`, `version integer DEFAULT 1`) and matching `.down.sql` (`DROP TABLE movies`)
- [x] T013 [P] Create migration `migrations/000002_create_users_table.up.sql` (`id bigserial PK`, `name text NOT NULL`, `email text UNIQUE NOT NULL`, `password_hash bytea NOT NULL`, `activated boolean DEFAULT false`, `created_at timestamptz DEFAULT now()`, `version integer DEFAULT 1`) and matching `.down.sql`
- [x] T014 [P] Create migration `migrations/000003_create_tokens_table.up.sql` (`hash bytea PK`, `user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE`, `expiry timestamptz NOT NULL`, `scope text NOT NULL`) and matching `.down.sql`
- [x] T015 [P] Create `internal/adapters/http/helpers.go` with: `readJSON(w, r, dst)` (decode body, enforce 1MB max, disallow unknown fields), `writeJSON(w, status, data, headers)` (encode envelope, set Content-Type), `background(fn func())` (goroutine with recover)
- [x] T016 [P] Create `internal/adapters/http/errors.go` with helper functions that call `writeJSON` for each error type: `errorResponse(w, r, status, message)`, `notFoundResponse`, `methodNotAllowedResponse`, `badRequestResponse`, `failedValidationResponse`, `editConflictResponse`, `invalidCredentialsResponse`, `invalidAuthenticationTokenResponse`, `authenticationRequiredResponse`, `inactiveAccountResponse`, `serverErrorResponse`
- [x] T017 Create `internal/adapters/http/middleware.go` with `recoverPanic(next http.Handler)`, `logRequest(next http.Handler)` (log method, URI, protocol, status, duration, request_id using `log/slog`), and `requestID(next http.Handler)` (generate/propagate `X-Request-Id` header and store in context)
- [x] T018 Create `internal/adapters/http/router.go` with `NewRouter(app *Application) http.Handler`: instantiate `chi.NewRouter()`, attach `recoverPanic`, `logRequest`, `requestID` middleware globally; return the router (routes added in story phases)
- [x] T019 Create `cmd/api/main.go` with: `config` struct (port, env, db DSN, db pool settings, version), flag parsing, `godotenv.Load()` call, DB open via `postgres.Open`, `application` struct holding config + db + logger (`slog.New`), `serve()` method starting `http.Server` with 30s timeouts; no handlers wired yet

**Checkpoint**: `go build ./...` MUST compile with no errors before proceeding.

---

## Phase 3: User Story 1 — Browse and Retrieve Movies (Priority: P1) 🎯 MVP

**Goal**: Any client can list all movies (filtered, sorted, paginated) and retrieve a single movie by ID.

**Independent Test**: Run `make db/up && make migrate/up && make run`, then:
```
curl localhost:4000/v1/movies          # returns paginated list with metadata
curl localhost:4000/v1/movies/1        # returns single movie
curl "localhost:4000/v1/movies?genres=drama&sort=-year"  # filtered + sorted
curl localhost:4000/v1/movies/9999     # returns 404
```

- [x] T020 [P] [US1] Create `internal/domain/movie.go`: define `Movie` struct (ID int64, Title string, Year int32, Runtime int32, Genres []string, CreatedAt time.Time, Version int32) and `NewMovie(title string, year, runtime int32, genres []string) (*Movie, error)` constructor that validates all fields using `pkg/validator` and returns validation errors
- [x] T021 [US1] Create `internal/adapters/repository/postgres/movie_repo.go`: implement `MovieRepository` with `Get(ctx, id)` (GORM First, map ErrRecordNotFound → domain.ErrNotFound) and `GetAll(ctx, filters)` (GORM Where with ILIKE title filter, `@>` array genres filter, ORDER BY mapped sort field, LIMIT/OFFSET pagination, COUNT for metadata)
- [x] T022 [US1] Create `internal/application/movie_service.go`: implement `GetMovie(ctx, id) (*domain.Movie, error)` and `ListMovies(ctx, filters MovieFilters) ([]*domain.Movie, Metadata, error)` use cases that call `MovieRepository` port methods
- [x] T023 [P] [US1] Create migration `migrations/000004_seed_sample_data.up.sql` inserting 20 sample movies (The Godfather 1972, The Shawshank Redemption 1994, Pulp Fiction 1994, Schindler's List 1993, The Dark Knight 2008, Forrest Gump 1994, Inception 2010, Interstellar 2014, The Matrix 1999, Goodfellas 1990, Fight Club 1999, The Silence of the Lambs 1991, Parasite 2019, Whiplash 2014, The Grand Budapest Hotel 2014, La La Land 2016, Get Out 2017, 1917 2019, Dune 2021, Everything Everywhere All at Once 2022) and matching `.down.sql` (`DELETE FROM movies`)
- [x] T024 [P] [US1] Add `listMovies(w http.ResponseWriter, r *http.Request)` handler to `internal/adapters/http/movies.go`: parse query params (title, genres, page, page_size, sort) with validation, call `MovieService.ListMovies`, write JSON envelope `{"movies":[...],"metadata":{...}}`
- [x] T025 [P] [US1] Add `showMovie(w http.ResponseWriter, r *http.Request)` handler to `internal/adapters/http/movies.go`: extract `:id` via `chi.URLParam`, validate positive integer, call `MovieService.GetMovie`, write JSON envelope `{"movie":{...}}`, map `ErrNotFound` → `notFoundResponse`
- [x] T026 [US1] Register `GET /v1/movies` → `listMovies` and `GET /v1/movies/{id}` → `showMovie` in `internal/adapters/http/router.go`
- [x] T027 [US1] Wire `MovieRepository` and `MovieService` into `application` struct in `cmd/api/main.go`; pass `application` to `NewRouter`

**Checkpoint**: US1 independently functional — movie list and detail endpoints return seeded data.

---

## Phase 4: User Story 5 — Service Health and Observability (Priority: P5)

**Goal**: Liveness, readiness, version, and metrics endpoints available to operators.

**Independent Test**:
```
curl localhost:4000/health             # {"status":"ok"}
curl localhost:4000/readyz             # {"status":"ok"} or 503 if DB down
curl localhost:4000/v1/healthcheck     # {"status":"available","system_info":{...}}
curl localhost:4000/debug/vars         # expvar JSON with counters
```

- [x] T028 [P] [US5] Create `internal/adapters/http/health.go`: implement `healthHandler` returning `{"status":"ok"}` for `GET /health` (no DB check; liveness only)
- [x] T029 [P] [US5] Add `readyzHandler` to `internal/adapters/http/health.go`: ping the DB via `db.Raw("SELECT 1").Error`; return `{"status":"ok"}` on success, `{"status":"unavailable"}` with HTTP 503 on failure
- [x] T030 [P] [US5] Add `healthcheckHandler` to `internal/adapters/http/health.go`: return `{"status":"available","system_info":{"environment":cfg.env,"version":cfg.version}}`
- [x] T031 [P] [US5] Add `debugVarsHandler` to `internal/adapters/http/health.go`: serve `expvar.Handler()` at `GET /debug/vars`
- [x] T032 [US5] Register `expvar.NewInt` counters (`total_requests_received`, `total_responses_sent`, `total_processing_time_μs`) in `cmd/api/main.go`; update `logRequest` middleware in `internal/adapters/http/middleware.go` to increment all three on each request
- [x] T033 [US5] Register `GET /health`, `GET /readyz`, `GET /v1/healthcheck`, `GET /debug/vars` routes in `internal/adapters/http/router.go`

**Checkpoint**: All 4 observability endpoints reachable without authentication.

---

## Phase 5: User Story 3 — User Registration and Activation (Priority: P3)

**Goal**: New users can register and activate their accounts via a token.

**Independent Test**:
```
curl -X POST localhost:4000/v1/users \
  -d '{"name":"Alice","email":"alice@example.com","password":"pa55word"}'
# → 202, activated:false; activation token logged to stdout

curl -X PUT localhost:4000/v1/users/activated \
  -d '{"token":"<TOKEN_FROM_LOGS>"}'
# → 200, activated:true
```

- [x] T034 [P] [US3] Create `internal/domain/user.go`: define `User` struct and `Password` struct (unexported `plaintext *string`, exported `Hash []byte`); implement `Password.Set(plaintext string) error` (bcrypt cost 12), `Password.Matches(plaintext string) (bool, error)`; implement `NewUser(name, email, plaintext string) (*User, error)` with validation
- [x] T035 [P] [US3] Create `internal/domain/token.go`: define `Token` struct (Plaintext string, Hash []byte, UserID int64, Expiry time.Time, Scope string); define scope constants (`ScopeActivation`, `ScopeAuthentication`, `ScopePasswordReset`); implement `NewToken(userID int64, ttl time.Duration, scope string) (*Token, error)` generating 16 `crypto/rand` bytes, base32-encoding to plaintext, storing SHA-256 hash
- [x] T036 [US3] Create `internal/adapters/repository/postgres/user_repo.go`: implement `UserRepository` with `Insert(ctx, user)` (GORM Create, map unique-constraint error → `ErrDuplicateEmail`), `GetByEmail(ctx, email)` (GORM Where, map not-found → `ErrNotFound`), `Update(ctx, user)` (GORM Where id AND version, RowsAffected=0 → `ErrEditConflict`, increment version)
- [x] T037 [US3] Create `internal/adapters/repository/postgres/token_repo.go`: implement `TokenRepository` with `Insert(ctx, token)` (GORM Create), `GetForUser(ctx, scope, plaintext)` (SHA-256 hash plaintext, GORM join tokens→users WHERE hash=$1 AND scope=$2 AND expiry>now(), return *domain.User or ErrInvalidToken), `DeleteAllForUser(ctx, scope, userID)` (GORM Delete WHERE user_id AND scope)
- [x] T038 [US3] Create `internal/application/user_service.go`: implement `RegisterUser(ctx, name, email, password string) (*domain.User, *domain.Token, error)` (create User via domain constructor, insert user, create activation token, insert token, return both); implement `ActivateUser(ctx, tokenPlaintext string) (*domain.User, error)` (get user for token, set Activated=true, update user, delete activation tokens for user)
- [x] T039 [P] [US3] Add `registerUser(w, r)` handler to `internal/adapters/http/users.go`: decode `{"name","email","password"}`, call `UserService.RegisterUser`, log activation token plaintext to stdout (fire-and-forget), write 202 with `{"user":{...}}`
- [x] T040 [P] [US3] Add `activateUser(w, r)` handler to `internal/adapters/http/users.go`: decode `{"token"}`, validate 26-char length, call `UserService.ActivateUser`, write 200 with `{"user":{...}}`; map `ErrInvalidToken` → `failedValidationResponse`
- [x] T041 [US3] Register `POST /v1/users` → `registerUser` and `PUT /v1/users/activated` → `activateUser` in `internal/adapters/http/router.go`
- [x] T042 [US3] Wire `UserRepository`, `TokenRepository`, and `UserService` into `application` struct in `cmd/api/main.go`

**Checkpoint**: Registration and activation flow fully functional end-to-end.

---

## Phase 6: User Story 4 — Authentication and Token Management (Priority: P4)

**Goal**: Active users can obtain bearer tokens; expired tokens are rejected; passwords can be reset.

**Independent Test**:
```
curl -X POST localhost:4000/v1/tokens/authentication \
  -d '{"email":"alice@example.com","password":"pa55word"}'
# → 201 with authentication_token

curl -H "Authorization: Bearer <TOKEN>" localhost:4000/v1/movies
# → 200 (token accepted by authenticate middleware)

curl -X POST localhost:4000/v1/tokens/password-reset \
  -d '{"email":"alice@example.com"}'
# → 202; reset token logged to stdout

curl -X PUT localhost:4000/v1/users/password \
  -d '{"password":"newpa55word","token":"<RESET_TOKEN>"}'
# → 200; old auth tokens invalidated
```

- [x] T043 [US4] Create `internal/application/auth_service.go`: implement `Authenticate(ctx, email, password string) (*domain.Token, error)` (get user by email, verify password with `Matches`, reject inactive users with `ErrInvalidToken`, create + insert auth token with 24h TTL, return token); implement `GeneratePasswordResetToken(ctx, email string) (*domain.Token, error)` (get user by email, create + insert password-reset token with 30m TTL, return token — no error if email not found to prevent enumeration)
- [x] T044 [US4] Extend `internal/application/user_service.go` with `UpdatePassword(ctx, tokenPlaintext, newPassword string) error`: get user for password-reset token, set new password hash, update user, delete ALL auth tokens for user, delete password-reset token
- [x] T045 [P] [US4] Create `internal/adapters/http/tokens.go`: implement `createAuthToken(w, r)` (decode email+password, call `AuthService.Authenticate`, write 201 `{"authentication_token":{"token":"...","expiry":"..."}}`, map `ErrInvalidToken` → `invalidCredentialsResponse`)
- [x] T046 [P] [US4] Add `createPasswordResetToken(w, r)` to `internal/adapters/http/tokens.go`: decode email, validate format, call `AuthService.GeneratePasswordResetToken`, log token to stdout, write 202 `{"message":"an email will be sent..."}` regardless of whether email is registered
- [x] T047 [US4] Add `updateUserPassword(w, r)` handler to `internal/adapters/http/users.go`: decode `{"password","token"}`, validate both fields, call `UserService.UpdatePassword`, write 200 `{"message":"your password was successfully reset"}`, map `ErrInvalidToken` → `failedValidationResponse`
- [x] T048 [US4] Add `authenticate(next http.Handler)` middleware to `internal/adapters/http/middleware.go`: extract `Authorization: Bearer <token>` header (return `invalidAuthenticationTokenResponse` if malformed), hash token, call `TokenRepository.GetForUser` for authentication scope, store `*domain.User` in request context; set `Vary: Authorization` response header
- [x] T049 [US4] Add `requireActivatedUser(next http.Handler)` middleware to `internal/adapters/http/middleware.go`: retrieve user from context (set by `authenticate`); return `authenticationRequiredResponse` if anonymous, `inactiveAccountResponse` if `Activated == false`
- [x] T050 [US4] Register in `internal/adapters/http/router.go`: `POST /v1/tokens/authentication` → `createAuthToken`, `POST /v1/tokens/password-reset` → `createPasswordResetToken`, `PUT /v1/users/password` → `updateUserPassword`; apply `authenticate` middleware globally (sets user context, does NOT require auth)
- [x] T051 [US4] Wire `AuthService` into `application` struct and pass `TokenRepository` to `authenticate` middleware in `cmd/api/main.go`

**Checkpoint**: Authentication token issued and accepted on protected endpoints; password reset invalidates old tokens.

---

## Phase 7: User Story 2 — Manage Movies (Priority: P2)

**Goal**: Authenticated, activated users can create, update, and delete movies; optimistic locking prevents silent overwrites.

**Depends on**: US1 (movie domain + read repo), US3 (user accounts), US4 (auth middleware)

**Independent Test**:
```
TOKEN=$(curl -s -X POST localhost:4000/v1/tokens/authentication \
  -d '{"email":"demo@cineapi.local","password":"pa55word"}' | jq -r .authentication_token.token)

curl -X POST localhost:4000/v1/movies \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Oppenheimer","year":2023,"runtime":180,"genres":["drama","history"]}'
# → 201 with Location header

curl -X PATCH localhost:4000/v1/movies/21 \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"runtime":181,"version":1}'
# → 200 with updated movie

curl -X DELETE localhost:4000/v1/movies/21 -H "Authorization: Bearer $TOKEN"
# → 200 {"message":"movie successfully deleted"}

curl localhost:4000/v1/movies/21   # → 404
```

- [x] T052 [US2] Extend `internal/application/movie_service.go` with: `CreateMovie(ctx, title string, year, runtime int32, genres []string) (*domain.Movie, error)` (construct via `domain.NewMovie`, call `MovieRepository.Insert`); `UpdateMovie(ctx, movie *domain.Movie) error` (call `MovieRepository.Update`, propagate `ErrEditConflict`); `DeleteMovie(ctx, id int64) error` (call `MovieRepository.Delete`, propagate `ErrNotFound`)
- [x] T053 [US2] Extend `internal/adapters/repository/postgres/movie_repo.go` with: `Insert(ctx, movie)` (GORM Create, map result back to movie.ID and movie.Version), `Update(ctx, movie)` (GORM Where `id = ? AND version = ?`, update all fields + `version = version + 1`, check `RowsAffected == 0` → `ErrEditConflict`), `Delete(ctx, id)` (GORM Delete Where id, check `RowsAffected == 0` → `ErrNotFound`)
- [x] T054 [P] [US2] Add `createMovie(w, r)` handler to `internal/adapters/http/movies.go`: decode `{"title","year","runtime","genres"}`, call `MovieService.CreateMovie`, set `Location: /v1/movies/{id}` header, write 201 `{"movie":{...}}`
- [x] T055 [P] [US2] Add `updateMovie(w, r)` handler to `internal/adapters/http/movies.go`: extract `:id`, fetch existing movie, decode partial JSON body using pointer fields (`*string`, `*int32`, `*[]string`) to detect supplied-vs-absent fields, apply only non-nil values to the movie, call `MovieService.UpdateMovie`; map `ErrEditConflict` → `editConflictResponse`, `ErrNotFound` → `notFoundResponse`
- [x] T056 [P] [US2] Add `deleteMovie(w, r)` handler to `internal/adapters/http/movies.go`: extract `:id`, call `MovieService.DeleteMovie`; map `ErrNotFound` → `notFoundResponse`; write 200 `{"message":"movie successfully deleted"}`
- [x] T057 [US2] Register in `internal/adapters/http/router.go` under `requireActivatedUser` middleware: `POST /v1/movies` → `createMovie`, `PATCH /v1/movies/{id}` → `updateMovie`, `DELETE /v1/movies/{id}` → `deleteMovie`

**Checkpoint**: All 5 user stories independently functional. Full API surface operational.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Seed demo user, final wiring, lint, doc sync.

- [x] T058 Create `internal/adapters/repository/postgres/seed.go` with `SeedDemoUser(db *gorm.DB) error`: insert a pre-activated user (`demo@cineapi.local`, password `pa55word`) if not already present; call from `cmd/api/main.go` after migrations
- [x] T059 [P] Copy `specs/001-movie-db-api/contracts/openapi.yaml` to `docs/openapi.yaml` (constitution-required location); verify the spec reflects all 12 implemented endpoints
- [x] T060 [P] Create `migrations/000005_seed_demo_user.up.sql` inserting the demo user with bcrypt hash of `pa55word` and `activated = true`; matching `.down.sql` (`DELETE FROM users WHERE email = 'demo@cineapi.local'`)
- [x] T061 Verify `cmd/api/main.go` wires all adapters: `movieRepo`, `userRepo`, `tokenRepo` passed to services; `AuthService.TokenRepository` reference used in `authenticate` middleware; all routes registered; server shutdown goroutine handles `SIGINT`/`SIGTERM`
- [x] T062 Run `golangci-lint run ./...` and fix all reported errors until zero remain
- [x] T063 Validate `quickstart.md` end-to-end: `make db/up` → `make migrate/up` → `make run` → execute every curl example in the quickstart; confirm all expected responses

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 ⚠️ BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2
- **US5 (Phase 4)**: Depends on Phase 2; can run in parallel with Phase 3
- **US3 (Phase 5)**: Depends on Phase 2; can start after Phase 2 (no dependency on US1 or US5)
- **US4 (Phase 6)**: Depends on US3 (Phase 5) — users must exist before auth tokens
- **US2 (Phase 7)**: Depends on US1 (Phase 3) + US3 (Phase 5) + US4 (Phase 6)
- **Polish (Phase 8)**: Depends on all phases complete

### User Story Dependencies

```
Phase 1 (Setup)
    └── Phase 2 (Foundational)
            ├── Phase 3 (US1: Browse Movies)  ─────────────────────┐
            ├── Phase 4 (US5: Observability)  ── independent        │
            └── Phase 5 (US3: Registration)                         │
                    └── Phase 6 (US4: Auth)                         │
                            └── Phase 7 (US2: Manage Movies) ←──────┘
                                    └── Phase 8 (Polish)
```

### Within Each Phase

- Models/entities → repositories → services → handlers → route registration → main.go wiring
- T053 (movie repo write methods) depends on T021 (movie repo read methods — same file)
- T044 (UpdatePassword) depends on T038 (UserService exists in same file)

### Parallel Opportunities

**Phase 1** — T003, T004, T005, T006 all run in parallel after T001+T002

**Phase 2** — after T007 and T008:
- T012, T013, T014 (migration files) in parallel
- T015, T016 (HTTP helpers, errors) in parallel
- T010 (ports) → T011 (DB setup) → T018 (main.go skeleton)

**Phase 3** — T020 and T023 in parallel; T024 and T025 in parallel after T022

**Phase 4** — T028, T029, T030, T031 all in parallel

**Phase 5** — T034 and T035 in parallel; T039 and T040 in parallel after T038

**Phase 6** — T045 and T046 in parallel after T043; T048 and T049 in parallel

**Phase 7** — T054, T055, T056 in parallel after T052 and T053

---

## Implementation Strategy

### MVP (User Story 1 only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (**CRITICAL** — blocks all stories)
3. Complete Phase 3: US1 — Browse and Retrieve Movies
4. **STOP and VALIDATE**: `curl localhost:4000/v1/movies` returns seeded movie data
5. Demo/deploy if sufficient

### Incremental Delivery

1. Phase 1+2 → Foundation ready
2. Phase 3 → Browse movies ✅ deploy
3. Phase 4 → Health endpoints ✅ deploy
4. Phase 5 → User registration ✅ deploy
5. Phase 6 → Authentication ✅ deploy
6. Phase 7 → Movie management ✅ deploy
7. Phase 8 → Polish → full release

### Parallel Team Strategy (3 developers)

After Phase 2 completes:
- **Dev A**: Phase 3 (US1) then Phase 7 (US2)
- **Dev B**: Phase 4 (US5) then Phase 5 (US3)
- **Dev C**: Phase 5 (US3) then Phase 6 (US4)

---

## Notes

- `[P]` tasks touch different files with no incomplete dependencies — safe to parallelise
- `[USx]` label maps each task to its user story for traceability
- Each checkpoint must pass before the next phase begins
- Partial PATCH (T055) MUST use pointer fields to distinguish absent-vs-zero values
- Token lookup always re-hashes the plaintext; plaintext is never stored or logged except the one-time dispatch log
- GORM tags MUST NOT appear on any `internal/domain/` struct — Principle III hard boundary
