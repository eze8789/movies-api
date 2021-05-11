# Movies API

movies-api is a project to implement an API to access movie informations.

The project integrates a GO webserver with a Postgres DB as a data store.

### How to run it

#### Build
```make build```

#### Run
```make build run```

#### Run GO in tmp
```make gorun```

#### Makefile support args
```$ make gorun ARGS="-help"
go run ./cmd/api/... -help
Usage of /tmp/go-build236066498/b001/exe/api:
  -env string
        Running environment (default "dev")
  -port int
        HTTP server port (default 8000)
```

##### Example changing environment and port
```make gorun ARGS="-env test","-port 4000"```

#### Help
```
$ make help
help:   Show this help
gorun:  Run go program, you can pass arguments using make run ARGS="-help"
pgrun:  Run postgres db
lint:   Run golangci-lint linter
build:  Build go binary
run:    Run binary, you can pass arguments using make run ARGS="-help"
```