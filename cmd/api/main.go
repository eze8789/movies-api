package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

const (
	webserverTimeout = 30
	version          = "0.1"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
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

	pgUser := os.Getenv("POSTGRES_USER")
	pgPWD := os.Getenv("POSTGRES_PWD")
	pgDB := os.Getenv("POSTGRES_DB")
	cfg.db.dsn = fmt.Sprintf("postgres://%s:%s@localhost/%s?sslmode=disable", pgUser, pgPWD, pgDB)

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := openDB(cfg.db.dsn)
	if err != nil {
		logger.Fatalf("unable to connect to database: %s\n", err)
	}
	defer db.Close()

	logger.Printf("database connection established")

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
			logger.Fatalf("could not start webserver on: %s. err: %s", srv.Addr, err)
		}
		wg.Done()
	}()
	wg.Wait()

	<-done
	logger.Print("webserver stopped")
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
