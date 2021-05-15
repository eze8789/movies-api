package main

import (
	"fmt"
	"net/http"
)

func (app *application) logError(r *http.Request, err error) {
	app.logger.LogError(err, map[string]string{"request_method": r.Method,
		"request_url":    r.URL.String(),
		"request_source": r.RemoteAddr,
	})
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

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	msg := "resource not found"
	app.errorResponse(w, r, http.StatusNotFound, msg)
}

func (app *application) notAllowedResponse(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("method not allowed: %s", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, msg)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, err.Error())
}

func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errs map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errs)
}

func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	msg := "unable to update record due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, msg)
}

func (app *application) rateLimitExceedResponse(w http.ResponseWriter, r *http.Request) {
	msg := "too many requests, rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, msg)
}

func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	msg := "invalid credentials"
	app.errorResponse(w, r, http.StatusUnauthorized, msg)
}

func (app *application) invalidAuthTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	msg := "invalid authentication token"
	app.errorResponse(w, r, http.StatusUnauthorized, msg)
}

func (app *application) authReqResponse(w http.ResponseWriter, r *http.Request) {
	msg := "authentication needed to access this resource"
	app.errorResponse(w, r, http.StatusUnauthorized, msg)
}

func (app *application) inactiveUserResponse(w http.ResponseWriter, r *http.Request) {
	msg := "please activate your account to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, msg)
}

func (app *application) unauthorizedResponse(w http.ResponseWriter, r *http.Request) {
	msg := "user account not authorized to perform that operation"
	app.errorResponse(w, r, http.StatusForbidden, msg)
}
