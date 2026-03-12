# Data Model: CineAPI — Movie Database REST API

**Feature**: `001-movie-db-api`
**Date**: 2026-03-12

---

## Entities

### Movie

Represents a film entry in the catalog.

```go
// internal/domain/movie.go
type Movie struct {
    ID        int64     // Auto-assigned; read-only after creation
    Title     string    // Required; 1–500 characters
    Year      int32     // Required; 1888 – current year + 1
    Runtime   int32     // Required; minutes; > 0
    Genres    []string  // Required; 1–5 unique string values
    CreatedAt time.Time // Set at creation; read-only
    Version   int32     // Starts at 1; incremented on every update (optimistic locking)
}
```

**Constructor**: `NewMovie(title string, year, runtime int32, genres []string) (Movie, error)`
Returns a validation error if any field fails its rule. The caller never constructs
a Movie literal directly.

**Validation rules**:
| Field | Rule |
|---|---|
| Title | Required; length 1–500 |
| Year | 1888 ≤ year ≤ current year + 1 |
| Runtime | > 0 |
| Genres | Length 1–5; all values unique |

**Version / optimistic locking**: On `PATCH`, the client MUST supply the current
`version` value. If the version in the store has advanced since the client read the
record, an `ErrEditConflict` is returned and the update is rejected.

---

### User

Represents an account holder.

```go
// internal/domain/user.go
type User struct {
    ID        int64     // Auto-assigned
    Name      string    // Required; 1–500 characters
    Email     string    // Required; unique; valid RFC 5322; ≤500 characters
    Password  Password  // Custom type — see below
    Activated bool      // false until activation token is consumed
    CreatedAt time.Time // Set at creation; read-only
    Version   int32     // Optimistic locking (used on activation and password update)
}

// Password wraps hashing so plaintext never leaks into callers.
type Password struct {
    plaintext *string // nil after hashing; only set during construction
    Hash      []byte  // bcrypt hash (cost 12); always stored
}
```

**Constructor**: `NewUser(name, email, plaintext string) (User, error)`

**Methods on Password**:
- `Set(plaintext string) error` — hashes and stores; clears plaintext.
- `Matches(plaintext string) (bool, error)` — bcrypt comparison.

**Validation rules**:
| Field | Rule |
|---|---|
| Name | Required; length 1–500 |
| Email | Required; valid email format; ≤500 chars |
| Password (plaintext) | Length 8–72 (72 = bcrypt max input) |

---

### Token

A short-lived, scoped credential issued to a user.

```go
// internal/domain/token.go
type Token struct {
    Plaintext string    // Returned to client once; never stored
    Hash      []byte    // SHA-256(plaintext); stored in repository
    UserID    int64     // Owner
    Expiry    time.Time // Absolute expiry timestamp
    Scope     string    // "activation" | "authentication" | "password-reset"
}

const (
    ScopeActivation    = "activation"     // 3-day expiry
    ScopeAuthentication = "authentication" // 24-hour expiry
    ScopePasswordReset = "password-reset"  // 30-minute expiry
)
```

**Constructor**: `NewToken(userID int64, ttl time.Duration, scope string) (Token, error)`
Generates 16 cryptographically random bytes → base32-encodes to a 26-char plaintext
string → stores SHA-256 hash. Returns the Token with plaintext populated.

**Lookup**: Tokens are always looked up by hashing the submitted plaintext and comparing
to stored hashes. The plaintext is never stored.

---

## Domain Errors

```go
// internal/domain/errors.go
var (
    ErrNotFound       = errors.New("record not found")
    ErrDuplicateEmail = errors.New("duplicate email address")
    ErrEditConflict   = errors.New("edit conflict")
    ErrInvalidToken   = errors.New("invalid or expired token")
)
```

These sentinel errors bubble from repository adapters through the application layer to
the HTTP adapter, which maps them to appropriate HTTP status codes.

---

## Port Interfaces

Defined in `internal/application/ports.go`. The application layer depends only on
these interfaces; concrete implementations live in adapters.

```go
type MovieRepository interface {
    Insert(ctx context.Context, movie *domain.Movie) error
    Get(ctx context.Context, id int64) (*domain.Movie, error)
    GetAll(ctx context.Context, filters MovieFilters) ([]*domain.Movie, Metadata, error)
    Update(ctx context.Context, movie *domain.Movie) error
    Delete(ctx context.Context, id int64) error
}

type UserRepository interface {
    Insert(ctx context.Context, user *domain.User) error
    GetByEmail(ctx context.Context, email string) (*domain.User, error)
    Update(ctx context.Context, user *domain.User) error
}

type TokenRepository interface {
    Insert(ctx context.Context, token *domain.Token) error
    GetForUser(ctx context.Context, scope string, tokenPlaintext string) (*domain.User, error)
    DeleteAllForUser(ctx context.Context, scope string, userID int64) error
}
```

### Supporting Types

```go
// MovieFilters — input to GetAll
type MovieFilters struct {
    Title    string
    Genres   []string
    Page     int
    PageSize int
    Sort     string   // e.g., "year" or "-year"
}

// Metadata — returned alongside paginated results
type Metadata struct {
    CurrentPage  int
    PageSize     int
    FirstPage    int
    LastPage     int
    TotalRecords int
}
```

---

## PostgreSQL Repository Layout (GORM Adapter)

GORM model structs live exclusively in `internal/adapters/repository/postgres/models.go`.
Domain entities are **never** passed directly to GORM — the adapter maps between them.

```go
// internal/adapters/repository/postgres/models.go

type MovieModel struct {
    ID        int64          `gorm:"primaryKey;autoIncrement"`
    Title     string         `gorm:"not null"`
    Year      int32          `gorm:"not null"`
    Runtime   int32          `gorm:"not null"`
    Genres    pq.StringArray `gorm:"type:text[];not null"`
    CreatedAt time.Time      `gorm:"autoCreateTime"`
    Version   int32          `gorm:"default:1;not null"`
}

type UserModel struct {
    ID           int64     `gorm:"primaryKey;autoIncrement"`
    Name         string    `gorm:"not null"`
    Email        string    `gorm:"uniqueIndex;not null"`
    PasswordHash []byte    `gorm:"not null"`
    Activated    bool      `gorm:"default:false;not null"`
    CreatedAt    time.Time `gorm:"autoCreateTime"`
    Version      int32     `gorm:"default:1;not null"`
}

type TokenModel struct {
    Hash      []byte    `gorm:"primaryKey"`
    UserID    int64     `gorm:"not null;index"`
    Expiry    time.Time `gorm:"not null"`
    Scope     string    `gorm:"not null"`
}
```

Each repository struct holds a `*gorm.DB` reference and implements the corresponding
port interface via GORM query methods.

```go
// internal/adapters/repository/postgres/movie_repo.go
type movieRepo struct{ db *gorm.DB }

func (r *movieRepo) Get(ctx context.Context, id int64) (*domain.Movie, error) {
    var m MovieModel
    result := r.db.WithContext(ctx).First(&m, id)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        return nil, domain.ErrNotFound
    }
    // ... map MovieModel → domain.Movie
}
```

### Database Schema (managed by golang-migrate)

```sql
-- migrations/000001_create_movies_table.up.sql
CREATE TABLE movies (
    id         bigserial PRIMARY KEY,
    title      text        NOT NULL,
    year       integer     NOT NULL,
    runtime    integer     NOT NULL,
    genres     text[]      NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    version    integer     NOT NULL DEFAULT 1
);

-- migrations/000002_create_users_table.up.sql
CREATE TABLE users (
    id            bigserial   PRIMARY KEY,
    name          text        NOT NULL,
    email         text UNIQUE NOT NULL,
    password_hash bytea       NOT NULL,
    activated     boolean     NOT NULL DEFAULT false,
    created_at    timestamptz NOT NULL DEFAULT now(),
    version       integer     NOT NULL DEFAULT 1
);

-- migrations/000003_create_tokens_table.up.sql
CREATE TABLE tokens (
    hash    bytea       PRIMARY KEY,
    user_id bigint      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expiry  timestamptz NOT NULL,
    scope   text        NOT NULL
);
```

---

## State Transitions

### User Activation Flow

```
[Registered] --submit activation token--> [Activated]
```
- `Activated = false` on `Insert`
- Set `Activated = true` on `Update` after token validated

### Password Reset Flow

```
[Authenticated] --request reset token-->
  [Token issued] --submit token + new password-->
    [Password updated, all auth tokens invalidated]
```

### Optimistic Locking (Movie + User Update)

```
client reads version=N
client PATCHes with version=N
  if store.version == N → update, increment version to N+1
  if store.version != N → return ErrEditConflict (HTTP 409)
```
