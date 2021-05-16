package main

import (
	"errors"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eze8789/movies-api/data"
	"github.com/eze8789/movies-api/validator"
	"github.com/felixge/httpsnoop"
	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "Close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimiter(next http.Handler) http.Handler {
	type client struct {
		lastSeen time.Time
		limiter  *rate.Limiter
	}

	// Protect concurrent writes/reads against the same memory map to limit amount of request
	var mu sync.Mutex
	var clients = make(map[string]*client)

	// clean up memory map if client request not see in 2 minutes
	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()
			for k, v := range clients {
				if time.Since(v.lastSeen) > 2*time.Minute {
					delete(clients, k)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {
			h, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			mu.Lock()

			if _, exist := clients[h]; !exist {
				clients[h] = &client{limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
			}
			clients[h].lastSeen = time.Now()

			if !clients[h].limiter.Allow() {
				app.logError(r, fmt.Errorf("%s - %s: %s Too many requests", r.RemoteAddr, r.Method, r.URL.String()))
				mu.Unlock()
				app.rateLimitExceedResponse(w, r)
				return
			}
			// Do not defer lock, if not block until all the middleware chain completes
			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")
		// Get the authorization header to retrieve the token
		authHeader := r.Header.Get("Authorization")

		// validate token is not empty, if it is return request with anonymous user
		if authHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}
		// validate token is valid
		authTokenParts := strings.Split(authHeader, " ")
		if len(authTokenParts) != 2 || authTokenParts[0] != "Bearer" {
			app.invalidAuthTokenResponse(w, r)
			return
		}
		token := authTokenParts[1]

		v := validator.New()
		if data.ValidateTokenPlain(v, token); !v.Valid() { //nolint:gocritic
			app.invalidAuthTokenResponse(w, r)
			return
		}

		// retrieve user information, early return with invalid token if not exists.
		u, err := app.models.Users.GetForToken(token, data.ScopeAuthentication)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// set user in the context and call next handler in chain
		r = app.contextSetUser(r, u)
		next.ServeHTTP(w, r)
	})
}

// reqAuthenticatedUser validate the user is authenticated
func (app *application) reqAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := app.contextGetUser(r)
		if u.IsAnonym() {
			app.authReqResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// reqActivatedUser validate the user activated the account
func (app *application) reqActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := app.contextGetUser(r)
		if !u.Activated {
			app.inactiveUserResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
	return app.reqAuthenticatedUser(fn)
}

func (app *application) reqPermission(perm string, next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := app.contextGetUser(r)
		userPerms, err := app.models.Permissions.GetAllForUser(u.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		if !userPerms.Include(perm) {
			app.unauthorizedResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
	return app.reqActivatedUser(fn)
}

func (app *application) metrics(next http.Handler) http.Handler {
	totalRequestReceived := expvar.NewInt("request_received")
	totalResponseSent := expvar.NewInt("response_sent")
	totalProcessingTime := expvar.NewInt("total_processing_time")
	totalResponseByCode := expvar.NewMap("total_response_by_status_code")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		totalRequestReceived.Add(1)
		metrics := httpsnoop.CaptureMetrics(next, w, r)
		totalResponseSent.Add(1)

		totalProcessingTime.Add(metrics.Duration.Milliseconds())
		totalResponseByCode.Add(strconv.Itoa(metrics.Code), 1)
	})
}
