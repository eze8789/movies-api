package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	rtr := httprouter.New()
	rtr.RedirectTrailingSlash = true
	rtr.NotFound = http.HandlerFunc(app.notFoundResponse)
	rtr.MethodNotAllowed = http.HandlerFunc(app.notAllowedResponse)

	rtr.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	rtr.HandlerFunc(http.MethodGet, "/v1/movies", app.listMovieHandler)
	rtr.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	rtr.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	rtr.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.updateMovieHandler)
	rtr.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)

	return app.recoverPanic(rtr)
}
