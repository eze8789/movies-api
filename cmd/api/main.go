package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/eze8789/movies-api/data"
	"github.com/eze8789/movies-api/jsonlog"
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
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 8000, "HTTP server port")
	flag.StringVar(&cfg.env, "env", "dev", "Running environment")
	flag.Parse()

	logLevel, err := GetInt("MOVIES_API_LOG_LEVEL")
	if err != nil {
		log.Fatal("please set a valid log level")
	}

	// Configure Rate Limiting
	cfg.limiter.rps, err = GetFloat("RATE_LIMIT_RPS")
	if err != nil {
		log.Fatal("please set a valid RPS rate limit value")
	}
	cfg.limiter.burst, err = GetInt("RATE_LIMIT_BURST")
	if err != nil {
		log.Fatal("please set a valid RPS rate limit value")
	}
	cfg.limiter.enabled = GetBool("RATE_LIMIT_ENABLED")

	// Configure Postgres DB
	pgUser := os.Getenv("POSTGRES_USER")
	pgPWD := os.Getenv("POSTGRES_PWD")
	pgDB := os.Getenv("POSTGRES_DB")
	cfg.db.maxOpenConns, err = GetInt("POSTGRES_MAX_OPEN_CONNS")
	if err != nil {
		log.Fatal("please set a valid MaxOpenConnections")
	}
	cfg.db.maxIdleConns, err = GetInt("POSTGRES_MAX_IDLE_CONNS")
	if err != nil {
		log.Fatal("please set a valid MaxIdleConnections")
	}
	cfg.db.maxIdleTime = os.Getenv("POSTGRES_MAX_IDLE_TIME")

	cfg.db.dsn = fmt.Sprintf("postgres://%s:%s@localhost/%s?sslmode=disable", pgUser, pgPWD, pgDB)

	logger := jsonlog.New(os.Stdout, jsonlog.Level(logLevel))

	db, err := openDB(&cfg)
	if err != nil {
		logger.LogFatal(err, nil)
	}
	defer db.Close()

	logger.LogInfo("database connection established", nil)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	app.server()
}

func openDB(cfg *config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	t, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //nolint:gomnd
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
