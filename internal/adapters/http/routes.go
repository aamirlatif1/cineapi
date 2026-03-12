package http

import "github.com/go-chi/chi/v5"

// registerRoutes wires all HTTP routes to their handlers.
// Routes are added in story-phase order: US5, US1, US3, US4, US2.
func (app *Application) registerRoutes(r *chi.Mux) {
	// US5 — Observability (no auth required)
	r.Get("/health", app.healthHandler)
	r.Get("/readyz", app.readyzHandler)
	r.Get("/debug/vars", app.debugVarsHandler)

	r.Route("/v1", func(r chi.Router) {
		// US5 — Healthcheck
		r.Get("/healthcheck", app.healthcheckHandler)

		// US1 — Browse movies (no auth required)
		r.Get("/movies", app.listMovies)
		r.Get("/movies/{id}", app.showMovie)

		// US2 — Manage movies (requires activated user)
		r.Group(func(r chi.Router) {
			r.Use(app.requireActivatedUser)
			r.Post("/movies", app.createMovie)
			r.Patch("/movies/{id}", app.updateMovie)
			r.Delete("/movies/{id}", app.deleteMovie)
		})

		// US3 — User registration and activation
		r.Post("/users", app.registerUser)
		r.Put("/users/activated", app.activateUser)

		// US4 — Auth tokens and password
		r.Post("/tokens/authentication", app.createAuthToken)
		r.Post("/tokens/password-reset", app.createPasswordResetToken)
		r.Put("/users/password", app.updateUserPassword)
	})
}
