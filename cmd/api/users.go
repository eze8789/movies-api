package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/eze8789/movies-api/data"
	"github.com/eze8789/movies-api/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.LogInfo(fmt.Sprintf("%s - %s: %s", r.RemoteAddr, r.Method, r.URL.String()), nil)
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.logger.LogError(err, nil)
		app.badRequestResponse(w, r, err)
		return
	}

	// copy json to struct to validate input before create the user in the database
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() { //nolint:gocritic
		app.logError(r, errors.New("user validation error"))
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		app.logError(r, err)
		switch {
		case errors.Is(err, data.ErrDuplicatedEmail):
			// This message makes us suceptible to user enumeration, if privacy matters
			// remove the message or make it ambiguous
			v.AddError("email", "email already registered")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// generate token for email activation
	token, err := app.models.Tokens.New(user.ID, data.ActivationTokenDuration, data.ScopeActivation)
	if err != nil {
		app.logger.LogError(err, nil)
		app.serverErrorResponse(w, r, err)
		return
	}

	// Run send mail process on background to ensure UX is not affected
	app.runBackground(func() {
		tmplData := map[string]interface{}{
			"activationToken": token.PlainToken,
			"userName":        user.Name,
		}
		err = app.mailer.Send(user.Email, "registered_user.tmpl", tmplData)
		if err != nil {
			app.logger.LogError(err, nil)
		}
	})

	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.logger.LogError(err, nil)
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.LogInfo(fmt.Sprintf("%s - %s: %s", r.RemoteAddr, r.Method, r.URL.String()), nil)

	var input struct {
		TokenPlain string `json:"token"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.logger.LogError(err, nil)
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateTokenPlain(v, input.TokenPlain); !v.Valid() { //nolint:gocritic
		app.logError(r, errors.New("token validation error"))
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// retrieve user for the given token
	u, err := app.models.Users.GetForToken(input.TokenPlain, data.ScopeActivation)
	if err != nil {
		app.logger.LogError(err, nil)
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	u.Activated = true
	err = app.models.Users.Update(u)

	if err != nil {
		app.logger.LogError(err, nil)
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.Tokens.DeleteAllByUser(u.ID, data.ScopeActivation)
	if err != nil {
		app.logger.LogError(err, nil)
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"user": u}, nil)
	if err != nil {
		app.logger.LogError(err, nil)
		app.serverErrorResponse(w, r, err)
	}
}
