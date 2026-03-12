package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/aamir-al/cineapi/internal/domain"
	"github.com/aamir-al/cineapi/pkg/validator"
)

// ----------------------------------------
// US4 — Token endpoints
// ----------------------------------------

func (app *Application) createAuthToken(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(validator.Matches(input.Email, validator.EmailRX), "email", "must be a valid email address")
	v.Check(input.Password != "", "password", "must be provided")
	v.Check(len(input.Password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(input.Password) <= 72, "password", "must not be more than 72 bytes long")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors())
		return
	}

	token, err := app.authSvc.Authenticate(r.Context(), input.Email, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidToken):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if err := app.writeJSON(w, http.StatusCreated, envelope{
		"authentication_token": map[string]any{
			"token":  token.Plaintext,
			"expiry": token.Expiry,
		},
	}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) createPasswordResetToken(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(validator.Matches(input.Email, validator.EmailRX), "email", "must be a valid email address")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors())
		return
	}

	token, err := app.authSvc.GeneratePasswordResetToken(r.Context(), input.Email)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Log token if issued; return 202 regardless to prevent enumeration.
	if token != nil {
		app.logger.Info("password reset token issued",
			slog.String("email", input.Email),
			slog.String("password_reset_token", token.Plaintext),
		)
	}

	if err := app.writeJSON(w, http.StatusAccepted, envelope{
		"message": "an email will be sent to you containing password reset instructions",
	}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
