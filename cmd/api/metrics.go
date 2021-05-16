package main

import (
	"database/sql"
	"expvar"
	"runtime"
	"time"
)

func exposeMetrics(db *sql.DB) {
	expvar.NewString("version").Set(version)

	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("db_status", expvar.Func(func() interface{} {
		return db.Stats()
	}))

	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))
}
