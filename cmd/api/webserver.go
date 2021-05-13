package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) server() {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		ErrorLog:     log.New(app.logger, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	// gracefully shutdown
	go func() {
		sig := <-quit
		app.logger.LogInfo("shutting down webserver, waiting for background tasks", map[string]string{"signal": sig.String()})
		ctx, cancel := context.WithTimeout(context.Background(), webserverTimeout*time.Second)
		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			app.logger.LogFatal(err, nil)
		}
		close(done)
	}()

	// start webserver
	app.wg.Add(1)
	go func() {
		app.logger.LogInfo("starting webserver", map[string]string{"environment": app.config.env, "port": srv.Addr})
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			app.logger.LogFatal(err, nil)
		}
		app.wg.Done()
	}()
	app.wg.Wait()

	<-done
	app.logger.LogInfo("webserver stopped", map[string]string{"environment": app.config.env, "port": srv.Addr})
}
