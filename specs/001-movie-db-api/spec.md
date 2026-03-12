# Feature Specification: CineAPI — Movie Database REST API

**Feature Branch**: `001-movie-db-api`
**Created**: 2026-03-12
**Status**: Draft
**Input**: User description: CineAPI is a JSON-based RESTful API for retrieving and
managing information about movies, with full support for user authentication, movie
CRUD operations, and application observability.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Browse and Retrieve Movies (Priority: P1)

A consumer of the API (e.g., a mobile app or website) wants to discover movies stored
in the database. They can list all movies — with the ability to filter by genre and
sort by title, year, or runtime — and retrieve the full details of any individual
movie by its unique identifier.

**Why this priority**: This is the primary read path and the core value of the service.
Every other feature depends on there being movie data to interact with, and the read
path can be delivered and tested independently with no authentication.

**Independent Test**: Seed the database with sample movies, issue `GET /v1/movies` and
`GET /v1/movies/:id`, and verify that the correct data is returned with consistent JSON
structure.

**Acceptance Scenarios**:

1. **Given** the movie catalog contains at least one entry, **When** a client requests
   the full list, **Then** the response contains all movies with their key attributes
   (title, year, runtime, genres) and a total record count.
2. **Given** a movie exists with a known identifier, **When** a client requests that
   movie by ID, **Then** the full movie record is returned.
3. **Given** a filter parameter (e.g., genre = "action") is provided, **When** a client
   lists movies, **Then** only movies matching the filter are returned.
4. **Given** a sort parameter (e.g., sort by year descending) is provided, **When** a
   client lists movies, **Then** results are returned in the requested order.
5. **Given** a requested movie ID does not exist, **When** the client retrieves it,
   **Then** a clear "not found" error is returned.

---

### User Story 2 - Manage Movies (Priority: P2)

An authenticated user with an active account wants to add new movies to the catalog,
update existing movie details, and remove movies that are no longer needed.

**Why this priority**: Write operations build on top of the read path (P1) and require
working user authentication (P3), so they are second in delivery order but represent
the other half of the core feature set.

**Independent Test**: Register and activate a test user, obtain an authentication token,
then perform `POST /v1/movies`, `PATCH /v1/movies/:id`, and `DELETE /v1/movies/:id`,
verifying that the catalog reflects each change.

**Acceptance Scenarios**:

1. **Given** an authenticated user provides valid movie details (title, year, runtime,
   genres), **When** they create a new movie, **Then** the movie is persisted and the
   full record is returned with an assigned identifier.
2. **Given** an authenticated user provides partial update data for an existing movie,
   **When** they patch the movie, **Then** only the supplied fields are updated and the
   unchanged fields retain their previous values.
3. **Given** an authenticated user requests deletion of an existing movie, **When** the
   request is processed, **Then** the movie is removed and subsequent retrieval returns
   "not found".
4. **Given** two clients simultaneously submit updates to the same movie, **When** the
   second update is processed, **Then** the system detects the conflict and returns a
   conflict error (optimistic locking via version field).
5. **Given** an unauthenticated client attempts to create, update, or delete a movie,
   **When** the request is processed, **Then** an "unauthorized" error is returned and
   no changes are made.

---

### User Story 3 - User Registration and Activation (Priority: P3)

A new user wants to create an account so they can access protected endpoints. After
submitting their details, they receive an activation token (via email) that they must
use to verify their address and activate their account before gaining full access.

**Why this priority**: User accounts are required for any write operation. Registration
and activation are the gateway to the rest of the authenticated surface.

**Independent Test**: Call `POST /v1/users` with valid details, receive an activation
token, call `PUT /v1/users/activated` with that token, and verify the account becomes
active.

**Acceptance Scenarios**:

1. **Given** a new visitor provides a unique email address, a name, and a valid password,
   **When** they register, **Then** an inactive account is created and an activation
   token is dispatched to their email.
2. **Given** a registered but inactive user has a valid activation token, **When** they
   submit that token, **Then** their account is marked active and they can authenticate.
3. **Given** an email address that is already registered, **When** a second registration
   attempt is made with that address, **Then** a clear "conflict" error is returned and
   no duplicate account is created.
4. **Given** an activation token is expired or invalid, **When** it is submitted,
   **Then** a clear "invalid token" error is returned.

---

### User Story 4 - Authentication and Token Management (Priority: P4)

An active user wants to obtain a short-lived authentication token to authorise requests
to protected endpoints, and to recover access to their account if they forget their
password.

**Why this priority**: Without authentication tokens the write endpoints (US2) are
inaccessible. Password reset is a necessary recovery path for any user account system.

**Independent Test**: Activate a test user, call `POST /v1/tokens/authentication` with
valid credentials, use the returned token on a protected endpoint, and verify access
is granted. Separately call `POST /v1/tokens/password-reset` and reset the password
via `PUT /v1/users/password`.

**Acceptance Scenarios**:

1. **Given** an active user provides valid credentials, **When** they request an
   authentication token, **Then** a short-lived bearer token is returned.
2. **Given** a user provides incorrect credentials, **When** they request a token,
   **Then** an "invalid credentials" error is returned.
3. **Given** an active user requests a password-reset token, **When** the request is
   processed, **Then** a password-reset token is dispatched to their registered email.
4. **Given** a user possesses a valid password-reset token and provides a new password,
   **When** they submit `PUT /v1/users/password`, **Then** their password is updated
   and all existing authentication tokens are invalidated.
5. **Given** a user submits an expired or already-used authentication token, **When**
   they access a protected endpoint, **Then** an "unauthorized" error is returned.

---

### User Story 5 - Service Health and Observability (Priority: P5)

An operator or automated monitoring system wants to check that the service is alive and
ready to accept traffic, and to inspect runtime metrics about its behaviour.

**Why this priority**: Observability is a cross-cutting concern needed from the first
deployment. It can be validated independently and adds no dependencies on other stories.

**Independent Test**: Issue `GET /v1/healthcheck` without credentials and verify a
status response is returned. Issue `GET /debug/vars` and verify runtime metrics are
accessible.

**Acceptance Scenarios**:

1. **Given** the service is running, **When** any client requests `GET /v1/healthcheck`,
   **Then** a response is returned with the current service status and version — no
   authentication required.
2. **Given** the service is running, **When** an operator requests `GET /debug/vars`,
   **Then** a response containing runtime metrics (e.g., request counts, goroutine
   count, memory stats) is returned.

---

### Edge Cases

- A request references a movie or user that does not exist → clear "not found" error.
- A client submits a request body with invalid or missing required fields → a "validation
  error" is returned listing every failing field.
- An authentication token has expired → "unauthorized" error; no data is leaked.
- A password-reset token is submitted more than once → second use returns an error;
  the token is single-use.
- Concurrent `PATCH` requests target the same movie with stale version numbers →
  conflict error is returned to the second writer; no silent data loss occurs.
- The database is unreachable at startup → the readiness endpoint returns a non-200
  status and the service does not accept traffic.
- A client sends a `PATCH` with an empty JSON body → no fields are changed; the current
  record is returned unchanged (idempotent no-op).
- Pagination parameters exceed total record count → an empty result set is returned
  without error.

## Requirements *(mandatory)*

### Functional Requirements

**Movie Catalogue**

- **FR-001**: System MUST allow any client (unauthenticated or authenticated) to
  retrieve a paginated list of all movies.
- **FR-002**: System MUST support filtering the movie list by title (partial match)
  and by one or more genres.
- **FR-003**: System MUST support sorting the movie list by title, year, or runtime
  in ascending or descending order.
- **FR-004**: System MUST allow any client to retrieve a single movie by its unique
  identifier.
- **FR-005**: System MUST allow authenticated users to create a new movie record by
  supplying title, release year, runtime, and at least one genre.
- **FR-006**: System MUST allow authenticated users to update one or more fields of an
  existing movie using a partial (PATCH) update.
- **FR-007**: System MUST allow authenticated users to delete a movie by identifier.
- **FR-008**: System MUST implement optimistic concurrency control on movie updates to
  prevent silent overwrites when two clients update the same record concurrently.

**User Accounts**

- **FR-009**: System MUST allow new visitors to register an account by providing a
  unique email address, display name, and password.
- **FR-010**: System MUST enforce password complexity (minimum 8 characters, maximum
  72 characters).
- **FR-011**: System MUST dispatch an account-activation token to the user's email
  address upon registration.
- **FR-012**: System MUST allow users to activate their account by submitting a valid
  activation token.
- **FR-013**: System MUST prevent duplicate accounts: registering with an already-used
  email MUST return an error.

**Authentication and Password Recovery**

- **FR-014**: System MUST allow active users to obtain a time-limited authentication
  bearer token by supplying valid credentials.
- **FR-015**: System MUST reject authentication requests from inactive or non-existent
  accounts.
- **FR-016**: System MUST allow active users to request a password-reset token sent to
  their registered email.
- **FR-017**: System MUST allow users to update their password by submitting a valid,
  unexpired password-reset token alongside a new password.
- **FR-018**: All existing authentication tokens for a user MUST be invalidated when
  that user's password is changed.

**Observability**

- **FR-019**: System MUST expose a health-check endpoint that returns service status
  and version information without requiring authentication.
- **FR-020**: System MUST expose a metrics endpoint that provides runtime statistics
  (request counts, memory, goroutines, etc.).
- **FR-021**: System MUST emit a structured log entry for every request, including
  status code, latency, request identifier, and HTTP method.

### Key Entities

- **Movie**: Represents a film entry. Key attributes: unique identifier, title (string),
  release year (integer), runtime in minutes (integer), genres (list of strings),
  creation timestamp, version counter (for optimistic locking).

- **User**: Represents an account holder. Key attributes: unique identifier, display
  name, email address (unique), hashed password, activation status (boolean), creation
  timestamp, version counter.

- **Token**: A short-lived credential scoped to a specific purpose (account activation,
  authentication, or password reset). Key attributes: plaintext value (returned once),
  stored hash, owning user, expiry timestamp, scope label.

### Assumptions

- Any authenticated, active user may perform all movie write operations; no separate
  admin role is required for this version.
- Email dispatch is fire-and-forget for this version; delivery confirmation and retry
  are out of scope.
- The `/debug/vars` metrics endpoint is accessible without authentication; it is
  assumed to be network-restricted at the infrastructure level in production.
- Pagination defaults: page size 20 records, first page is page 1.
- Authentication tokens expire after 24 hours; activation tokens expire after 3 days;
  password-reset tokens expire after 30 minutes.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All movie read operations return results within 500 ms under a load of 50
  concurrent users.
- **SC-002**: A new user can complete registration, account activation, and first
  authenticated request within 5 minutes using only the API documentation.
- **SC-003**: All API responses for both success and error cases conform to a consistent
  JSON envelope structure, validated by an automated contract test suite.
- **SC-004**: The service health endpoint responds within 100 ms regardless of database
  state, ensuring monitoring systems always receive a timely answer.
- **SC-005**: Concurrent update conflicts are detected 100 % of the time; no silent
  data-loss occurs under concurrent write load.
- **SC-006**: All validation errors return a response body identifying each failing
  field by name, enabling clients to surface actionable feedback to users.
- **SC-007**: Expired or invalid tokens are rejected 100 % of the time; no
  authenticated resource is accessible via an expired credential.
