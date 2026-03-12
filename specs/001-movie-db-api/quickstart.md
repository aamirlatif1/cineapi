# Quickstart: CineAPI

**Branch**: `001-movie-db-api`
**Date**: 2026-03-12

This guide covers: prerequisites, build, run, and smoke-test of the CineAPI service.

---

## Prerequisites

| Tool | Version | Install |
|---|---|---|
| Go | 1.22+ | https://go.dev/dl/ |
| Docker + Docker Compose | any | https://docs.docker.com/get-docker/ |
| golang-migrate CLI | any | `go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest` |
| make | any | pre-installed on macOS/Linux |
| curl | any | pre-installed on macOS/Linux |

---

## 1. Start PostgreSQL

```bash
docker compose up -d
# Starts postgres:16 on localhost:5432 (cineapi / cineapi / cineapi)
```

---

## 2. Run Migrations

```bash
make migrate/up
# equivalent to:
# migrate -path ./migrations -database "$DATABASE_URL" up
```

This creates all tables and seeds the 20 sample movies + demo user.

---

## 3. Build and Run

```bash
go mod tidy
make build     # compiles ./bin/api
make run       # go run ./cmd/api -env=development -port=4000
```

---

## 4. Configuration

The server accepts the following flags (all have defaults):

| Flag | Default | Description |
|---|---|---|
| `-port` | `4000` | TCP port to listen on |
| `-env` | `development` | Environment name (`development` \| `production`) |
| `-version` | `1.0.0` | Reported version in healthcheck |
| `-db-dsn` | env `DATABASE_URL` | PostgreSQL DSN |
| `-db-max-open-conns` | `25` | Max open DB connections |
| `-db-max-idle-conns` | `25` | Max idle DB connections |

Log format:
- `development` → human-readable text
- `production` → structured JSON

**Environment variables** (`.env` file, git-ignored):

```
DATABASE_URL=postgres://cineapi:cineapi@localhost:5432/cineapi?sslmode=disable
```

---

## 5. Verify the Service is Running

```bash
curl -s localhost:4000/v1/healthcheck | jq
```

Expected response:

```json
{
  "status": "available",
  "system_info": {
    "environment": "development",
    "version": "1.0.0"
  }
}
```

---

## 6. Browse Movies (no authentication required)

```bash
# List all movies
curl -s localhost:4000/v1/movies | jq

# Filter by genre
curl -s "localhost:4000/v1/movies?genres=sci-fi" | jq

# Filter by title (partial match)
curl -s "localhost:4000/v1/movies?title=dark" | jq

# Sort by year descending, page 1
curl -s "localhost:4000/v1/movies?sort=-year&page=1&page_size=5" | jq

# Get a specific movie
curl -s localhost:4000/v1/movies/1 | jq
```

---

## 7. Register a New User

```bash
curl -s -X POST localhost:4000/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Smith","email":"alice@example.com","password":"pa55word"}' \
  | jq
```

Expected response (`activated: false`):

```json
{
  "user": {
    "id": 2,
    "name": "Alice Smith",
    "email": "alice@example.com",
    "activated": false,
    "created_at": "2026-03-12T10:00:00Z"
  }
}
```

The activation token is **logged to stdout** (search for `activation_token` in server
output). Copy it for the next step.

---

## 8. Activate the User Account

```bash
# Replace TOKEN with the value from server logs
curl -s -X PUT localhost:4000/v1/users/activated \
  -H "Content-Type: application/json" \
  -d '{"token":"TOKEN"}' \
  | jq
```

Expected: `"activated": true` in the user object.

> **Demo shortcut**: A pre-activated user is seeded at startup:
> - Email: `demo@cineapi.local`
> - Password: `pa55word`

---

## 9. Authenticate (get a Bearer Token)

```bash
curl -s -X POST localhost:4000/v1/tokens/authentication \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@cineapi.local","password":"pa55word"}' \
  | jq
```

Expected:

```json
{
  "authentication_token": {
    "token": "XXXXXXXXXXXXXXXXXXXXXXXXXXX",
    "expiry": "2026-03-13T10:00:00Z"
  }
}
```

Save the token value:

```bash
TOKEN="XXXXXXXXXXXXXXXXXXXXXXXXXXX"
```

---

## 10. Create a Movie (requires authentication)

```bash
curl -s -X POST localhost:4000/v1/movies \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Oppenheimer","year":2023,"runtime":180,"genres":["drama","history"]}' \
  | jq
```

Expected: `HTTP 201` with the created movie and a `Location` header.

---

## 11. Update a Movie (partial patch)

```bash
# Patch only the runtime; must pass current version for conflict detection
curl -s -X PATCH localhost:4000/v1/movies/21 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"runtime":181,"version":1}' \
  | jq
```

---

## 12. Delete a Movie

```bash
curl -s -X DELETE localhost:4000/v1/movies/21 \
  -H "Authorization: Bearer $TOKEN" \
  | jq
```

---

## 13. Password Reset Flow

```bash
# 1. Request reset token (logged to stdout)
curl -s -X POST localhost:4000/v1/tokens/password-reset \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com"}' | jq

# 2. Reset password with the token from server logs
curl -s -X PUT localhost:4000/v1/users/password \
  -H "Content-Type: application/json" \
  -d '{"password":"newpa55word","token":"RESET_TOKEN"}' | jq
```

---

## 14. Runtime Metrics

```bash
curl -s localhost:4000/debug/vars | jq '{
  total_requests_received,
  total_responses_sent,
  total_processing_time_μs
}'
```

---

## Makefile Targets

| Target | Command | Description |
|---|---|---|
| `make build` | `go build -o bin/api ./cmd/api` | Compile binary |
| `make run` | `go run ./cmd/api` | Run in development mode |
| `make test` | `go test ./...` | Run all tests |
| `make lint` | `golangci-lint run` | Run linter |
| `make tidy` | `go mod tidy` | Sync module dependencies |
| `make migrate/up` | `migrate … up` | Apply all pending migrations |
| `make migrate/down` | `migrate … down` | Roll back last migration |
| `make db/up` | `docker compose up -d` | Start PostgreSQL container |
| `make db/down` | `docker compose down` | Stop and remove PostgreSQL container |

---

## Common Errors

| Error | Cause | Fix |
|---|---|---|
| `401 Unauthorized` | Missing or expired token | Re-authenticate via `/v1/tokens/authentication` |
| `403 Forbidden` | Account not activated | Activate account via `/v1/users/activated` |
| `404 Not Found` | ID does not exist | Check the ID; may have been deleted |
| `409 Conflict` | Stale version on PATCH | Re-fetch the movie and retry with current version |
| `422 Unprocessable Entity` | Validation failed | Check the `error` object for field-level messages |
