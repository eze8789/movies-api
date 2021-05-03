package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
	rtr := httprouter.New()
	rtr.RedirectTrailingSlash = true
	rtr.NotFound = http.HandlerFunc(app.notFoundResponse)
	rtr.MethodNotAllowed = http.HandlerFunc(app.notAllowedResponse)

	rtr.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	rtr.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	rtr.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)

	return rtr
}
