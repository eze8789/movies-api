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

	rtr.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheck)

	// Movies Endpoints, for this access user activated and authenticated is required
	rtr.HandlerFunc(http.MethodGet, "/v1/movies", app.reqPermission("movies:read", app.listMovie))
	rtr.HandlerFunc(http.MethodPost, "/v1/movies", app.reqPermission("movies:write", app.createMovieHandler))
	rtr.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.reqPermission("movies:read", app.showMovie))
	rtr.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.reqPermission("movies:write", app.updateMovie))
	rtr.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.reqPermission("movies:write", app.deleteMovie))

	// Users Endpoints
	rtr.HandlerFunc(http.MethodPost, "/v1/users", app.registerUser)
	rtr.HandlerFunc(http.MethodPut, "/v1/users/activate", app.activateUser)

	// Tokens Endpoints
	rtr.HandlerFunc(http.MethodPost, "/v1/tokens/activation", app.createActivationToken)
	rtr.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationToken)

	return app.recoverPanic(app.rateLimiter(app.authenticate(rtr)))
}
