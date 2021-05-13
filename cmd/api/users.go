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

	// Run send mail process on background to ensure UX is not affected
	app.runBackground(func() {
		err = app.mailer.Send(user.Email, "registered_user.tmpl", user)
		if err != nil {
			app.logError(r, err)
		}
	})

	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.logError(r, err)
		app.serverErrorResponse(w, r, err)
	}
}
