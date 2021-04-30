package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	webserverTimeout = 30
	version          = "0.1"
)

type config struct {
	port int
	env  string
}

type application struct {
	config config
	logger *log.Logger
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 8000, "HTTP server port")
	flag.StringVar(&cfg.env, "env", "dev", "Running environment")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	app := application{
		config: cfg,
		logger: logger,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	// gracefully shutdown
	go func() {
		<-quit
		logger.Println("shutting down webserver")
		ctx, cancel := context.WithTimeout(context.Background(), webserverTimeout*time.Second)
		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			logger.Fatal("could not gracefully shutdown server")
		}
		close(done)
	}()

	// start webserver
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		logger.Printf("starting server in env %s, port %d", cfg.env, cfg.port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Fatalf("could not start webserver on: %s. err: ", srv.Addr, err)
		}
		wg.Done()
	}()
	wg.Wait()

	<-done
	logger.Print("webserver stopped")
}
