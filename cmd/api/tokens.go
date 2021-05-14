package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/eze8789/movies-api/data"
	"github.com/eze8789/movies-api/validator"
)

// createActivationToken send an email with a new activation token in case the original expired or was lost by user
func (app *application) createActivationToken(w http.ResponseWriter, r *http.Request) {
	app.logger.LogInfo(fmt.Sprintf("%s - %s: %s", r.RemoteAddr, r.Method, r.URL.String()), nil)
	var input struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// validate email is a valid email address
	v := validator.New()
	if data.ValidateEmail(v, input.Email); !v.Valid() { //nolint:gocritic
		app.logError(r, errors.New("email validation error"))
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	u, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("email", "no matching email address found")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if u.Activated {
		v.AddError("email", "user already activated")
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	token, err := app.models.Tokens.New(u.ID, data.ActivationTokenDuration, data.ScopeActivation)
	if err != nil {
		app.logger.LogError(err, nil)
		app.serverErrorResponse(w, r, err)
		return
	}

	// Run send mail process on background to ensure UX is not affected
	app.runBackground(func() {
		tmplData := map[string]interface{}{
			"activationToken": token.PlainToken,
			"userName":        u.Name,
		}
		err = app.mailer.Send(u.Email, "token_activation.tmpl", tmplData)
		if err != nil {
			app.logger.LogError(err, nil)
		}
	})

	// Send response with status accepted in case the previous goroutine fails
	msg := envelope{"message": "an email will be sent to the registered email with activation instructions"}
	err = app.writeJSON(w, http.StatusAccepted, msg, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// createAuthenticationToken return a token to an authenticated user to use the API
func (app *application) createAuthenticationToken(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlain(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// get user by email, keep error ambiguous to not show information in case of an attack against a user
	u, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	b, err := u.Password.Match(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !b {
		app.invalidCredentialsResponse(w, r)
		return
	}

	token, err := app.models.Tokens.New(u.ID, data.AuthenticationTokenDuration, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
