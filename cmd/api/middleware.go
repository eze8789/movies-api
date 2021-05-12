package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

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
