package http

import (
	"context"
	"expvar"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/aamir-al/cineapi/internal/domain"
)

// contextKey is an unexported type for context keys in this package.
type contextKey string

const userContextKey = contextKey("user")

// ----------------------------------------
// recoverPanic
// ----------------------------------------

// recoverPanic recovers from any panic in the next handler, closes the connection,
// and returns a 500 response.
func (app *Application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ----------------------------------------
// requestID
// ----------------------------------------

// requestID generates or propagates an X-Request-Id header and stores it in the context.
func (app *Application) requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = generateRequestID()
		}
		ctx := context.WithValue(r.Context(), contextKey("request_id"), id)
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// ----------------------------------------
// logRequest
// ----------------------------------------

// logRequest logs each request's method, URI, status code, duration, and request_id,
// and increments the expvar request/response counters.
func (app *Application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID, _ := r.Context().Value(contextKey("request_id")).(string)

		totalRequestsReceived.Add(1)

		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		totalResponsesSent.Add(1)
		totalProcessingTimeMicroseconds.Add(duration.Microseconds())

		app.logger.Info("request",
			slog.String("method", r.Method),
			slog.String("uri", r.URL.RequestURI()),
			slog.String("proto", r.Proto),
			slog.Int("status", wrapped.status),
			slog.String("duration", duration.String()),
			slog.String("request_id", reqID),
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// ----------------------------------------
// expvar counters (registered in main.go)
// ----------------------------------------

var (
	totalRequestsReceived           *expvar.Int
	totalResponsesSent              *expvar.Int
	totalProcessingTimeMicroseconds *expvar.Int
)

// RegisterMetrics must be called once from main before serving requests.
func RegisterMetrics() {
	totalRequestsReceived = expvar.NewInt("total_requests_received")
	totalResponsesSent = expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds = expvar.NewInt("total_processing_time_μs")
}

// ----------------------------------------
// authenticate
// ----------------------------------------

// authenticate extracts the Bearer token from the Authorization header, looks up the
// owning user, and stores the user in the request context.
// Anonymous requests (no header) are allowed through with a nil user in context.
func (app *Application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := parts[1]
		if len(token) != 26 {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := app.tokenRepo.GetForUser(r.Context(), domain.ScopeAuthentication, token)
		if err != nil {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ----------------------------------------
// requireActivatedUser
// ----------------------------------------

// requireActivatedUser ensures the request carries a valid, activated user.
func (app *Application) requireActivatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(userContextKey).(*domain.User)
		if !ok || user == nil {
			app.authenticationRequiredResponse(w, r)
			return
		}
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
