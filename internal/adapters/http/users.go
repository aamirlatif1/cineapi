package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/aamir-al/cineapi/internal/application"
	"github.com/aamir-al/cineapi/internal/domain"
	"github.com/aamir-al/cineapi/pkg/validator"
)

// ----------------------------------------
// US3 — User registration and activation
// ----------------------------------------

func (app *Application) registerUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, token, err := app.userSvc.RegisterUser(r.Context(), input.Name, input.Email, input.Password)
	if err != nil {
		var ve *application.ValidationError
		switch {
		case errors.As(err, &ve):
			app.failedValidationResponse(w, r, ve.Fields)
		case errors.Is(err, domain.ErrDuplicateEmail):
			v := validator.New()
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors())
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Log activation token to stdout (fire-and-forget; no email infra in this demo)
	app.logger.Info("activation token issued",
		slog.String("email", user.Email),
		slog.String("activation_token", token.Plaintext),
	)

	if err := app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) activateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(input.TokenPlaintext != "", "token", "must be provided")
	v.Check(len(input.TokenPlaintext) == 26, "token", "must be 26 bytes long")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors())
		return
	}

	user, err := app.userSvc.ActivateUser(r.Context(), input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidToken):
			v := validator.New()
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors())
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// ----------------------------------------
// US4 — Password update
// ----------------------------------------

func (app *Application) updateUserPassword(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Password       string `json:"password"`
		TokenPlaintext string `json:"token"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(input.Password != "", "password", "must be provided")
	v.Check(len(input.Password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(input.Password) <= 72, "password", "must not be more than 72 bytes long")
	v.Check(input.TokenPlaintext != "", "token", "must be provided")
	v.Check(len(input.TokenPlaintext) == 26, "token", "must be 26 bytes long")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors())
		return
	}

	if err := app.userSvc.UpdatePassword(r.Context(), input.TokenPlaintext, input.Password); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidToken):
			v := validator.New()
			v.AddError("token", "invalid or expired password reset token")
			app.failedValidationResponse(w, r, v.Errors())
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"message": "your password was successfully reset"}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
