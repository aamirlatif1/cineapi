# CineAPI — Requirements Document

## 1. Project Overview

CineAPI is a JSON-based RESTful API for retrieving and managing information about movies. It serves as a comprehensive movie database service — similar in concept to the Open Movie Database API — with full support for user authentication, movie CRUD operations, and application observability.

## 2. Supported APIs

| Method | URL Pattern | Action |
|--------|-------------|--------|
| `GET` | `/v1/healthcheck` | Show application health and version information |
| `GET` | `/v1/movies` | Show the details of all movies |
| `POST` | `/v1/movies` | Create a new movie |
| `GET` | `/v1/movies/:id` | Show the details of a specific movie |
| `PATCH` | `/v1/movies/:id` | Update the details of a specific movie |
| `DELETE` | `/v1/movies/:id` | Delete a specific movie |
| `POST` | `/v1/users` | Register a new user |
| `PUT` | `/v1/users/activated` | Activate a specific user |
| `PUT` | `/v1/users/password` | Update the password for a specific user |
| `POST` | `/v1/tokens/authentication` | Generate a new authentication token |
| `POST` | `/v1/tokens/password-reset` | Generate a new password-reset token |
| `GET` | `/debug/vars` | Display application metrics |

## 3. Resource Definitions

### 3.1 Movie

A movie resource represents a single film entry in the database.

| Field | Type | Description |
|-------|------|-------------|
| `id` | `integer` | Unique identifier (auto-generated) |
| `title` | `string` | Title of the movie |
| `year` | `integer` | Release year |
| `runtime` | `string` | Runtime duration (e.g. `"102 mins"`) |
| `genres` | `[]string` | List of genres |
| `version` | `integer` | Record version number (used for optimistic concurrency control) |
| `created_at` | `datetime` | Timestamp of record creation |

### 3.2 User

A user resource represents a registered consumer of the API.

| Field | Type | Description |
|-------|------|-------------|
| `id` | `integer` | Unique identifier (auto-generated) |
| `name` | `string` | Full name of the user |
| `email` | `string` | Email address (unique) |
| `password` | `string` | Hashed password (never returned in responses) |
| `activated` | `boolean` | Whether the user account has been activated |
| `created_at` | `datetime` | Timestamp of registration |

### 3.3 Token

A token resource is used for authentication and password-reset flows.

| Field | Type | Description |
|-------|------|-------------|
| `token` | `string` | Plaintext token value (returned once on creation) |
| `user_id` | `integer` | Associated user ID |
| `expiry` | `datetime` | Token expiration timestamp |
| `scope` | `string` | Token scope (`authentication` or `password-reset`) |

## 4. Functional Requirements

### 4.1 Healthcheck

- The `/v1/healthcheck` endpoint must return the current application status, environment name, and version number.
- No authentication is required.

### 4.2 Movies

- **List Movies** — Support filtering by title, genres, and pagination via `page`, `page_size`, and `sort` query parameters.
- **Create Movie** — Accept a JSON body with `title`, `year`, `runtime`, and `genres`. Validate all input fields. Requires authentication.
- **Read Movie** — Return a single movie by its ID. Return `404 Not Found` if the movie does not exist.
- **Update Movie** — Accept a partial JSON body (via `PATCH`) and apply changes. Implement optimistic concurrency control using the `version` field to prevent edit conflicts. Requires authentication.
- **Delete Movie** — Remove a movie by its ID. Requires authentication with appropriate permissions.

### 4.3 Users

- **Register** — Accept `name`, `email`, and `password`. Hash the password before storage. Send an activation email with a token upon successful registration.
- **Activate** — Accept an activation token and set the user's `activated` field to `true`.
- **Password Reset** — Accept a new password along with a valid password-reset token and update the user's password.

### 4.4 Tokens

- **Authentication Token** — Accept `email` and `password`. On successful credential verification, return a bearer token with a defined expiry.
- **Password-Reset Token** — Accept an `email`, generate a password-reset token, and send it to the user's email.

### 4.5 Debug Metrics

- The `/debug/vars` endpoint must expose runtime application metrics (e.g. goroutine count, memory usage, request counts).
- Access should be restricted in production environments.

## 5. Non-Functional Requirements

### 5.1 Data Format

- All request and response bodies must use `application/json`.
- Envelope all successful responses in a top-level JSON object keyed by the resource name (e.g. `{"movie": {...}}`).

### 5.2 Error Handling

- Return consistent, structured error responses in the format: `{"error": "message"}` or `{"errors": {"field": "message"}}` for validation errors.
- Use appropriate HTTP status codes (400, 401, 403, 404, 409, 422, 429, 500).

### 5.3 Validation

- All input data must be validated before processing. Invalid requests should return `422 Unprocessable Entity` with descriptive error messages.

### 5.4 Authentication & Authorization

- Use stateful bearer token authentication.
- Tokens must be stored as SHA-256 hashes in the database; plaintext tokens are returned to the user only once at creation.
- Protect endpoints based on user activation status and permission scope.

### 5.5 Rate Limiting

- Implement IP-based rate limiting to protect against abuse.
- Return `429 Too Many Requests` when the limit is exceeded.

### 5.6 CORS

- Support configurable Cross-Origin Resource Sharing (CORS) with trusted origins.

### 5.7 Logging & Metrics

- Log all incoming requests with method, URL, response status, and duration.
- Expose internal metrics via the `/debug/vars` endpoint.

### 5.8 Graceful Shutdown

- The application must handle interrupt and termination signals and complete in-flight requests before shutting down.

### 5.9 Database

- Use PostgreSQL as the primary data store.
- Manage schema changes using versioned SQL migrations.

## 6. Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go |
| Database | PostgreSQL |
| Migrations | `golang-migrate` |
| Router | `chi` or standard library (Go 1.25+) |
