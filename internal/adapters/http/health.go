package http

import (
	"expvar"
	"net/http"
)

// healthHandler is the liveness probe — returns 200 as long as the process runs.
func (app *Application) healthHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.writeJSON(w, http.StatusOK, envelope{"status": "ok"}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// readyzHandler is the readiness probe — returns 200 only when the DB is reachable.
func (app *Application) readyzHandler(w http.ResponseWriter, r *http.Request) {
	sqlDB, err := app.db.DB()
	if err != nil || sqlDB.PingContext(r.Context()) != nil {
		if writeErr := app.writeJSON(w, http.StatusServiceUnavailable,
			envelope{"status": "unavailable"}, nil); writeErr != nil {
			app.serverErrorResponse(w, r, writeErr)
		}
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"status": "ok"}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// healthcheckHandler returns service status and version information.
func (app *Application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	data := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.Env,
			"version":     app.config.Version,
		},
	}
	if err := app.writeJSON(w, http.StatusOK, data, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// debugVarsHandler exposes the stdlib expvar metrics as JSON.
func (app *Application) debugVarsHandler(w http.ResponseWriter, r *http.Request) {
	expvar.Handler().ServeHTTP(w, r)
}
