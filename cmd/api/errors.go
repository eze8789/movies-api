package main

import (
	"fmt"
	"net/http"
)

func (app *application) logError(r *http.Request, err error) {
	app.logger.Println(err)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	msg := "the server encountered an error and could not process your request"

	app.errorResponse(w, r, http.StatusInternalServerError, msg)
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	msg := "resource not found"
	app.errorResponse(w, r, http.StatusNotFound, msg)
}

func (app *application) notAllowedResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	msg := fmt.Sprintf("method not allowed: %s", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, msg)
}
