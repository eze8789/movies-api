package main

import (
	"context"
	"net/http"

	"github.com/eze8789/movies-api/data"
)

type contextKey string

const userContextKey = contextKey("user")

// contextSetUser return a new copy of the request using our own custom key to add the User struct for authentication
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// contextGetUser extract the User struct from the request context
// Intentionally panic since it's an unexpected error, we only should see it if we call the method at the wrong time (bug)
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("no user value in request context")
	}
	return user
}
