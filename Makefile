.PHONY: help
help:  					## Show this help
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

.PHONY: gorun
gorun: 					## Run go program, you can pass arguments using make run ARGS="-help"
	@echo 'Running movies-api'
	@go run ./cmd/api/... $(ARGS)

.PHONY: pgrun
pgrun: 					## Run postgres db
	@docker stop postgresql-movies && docker rm postgresql-movies && docker run --name postgresql-movies -p 5432:5432 -e POSTGRES_PASSWORD=test -v ~/go/src/github.com/eze8789/movies-api/postgres-volume:/var/lib/postgresql/data -d postgres

.PHONY: lint
lint:  					## Run golangci-lint linter
	@echo 'Running linting...'
	golangci-lint run ./...

current_time = $(shell date --iso-8601=seconds)
git_description = $(shell git describe --always --dirty)
linker_flags = '-s -w -X main.buildTime=${current_time} -X main.version=${git_description}'

.PHONY: build
build: 					## Build go binary
	@echo 'Building movies-api...'
	@CGO_ENABLED=0 GOOS=linux go build -a -o ./bin/movies-api -ldflags=${linker_flags} ./cmd/...

.PHONY: run
run:   					## Run binary, you can pass arguments using make run ARGS="-help"
	@echo 'Starting movies-api'
	./bin/movies-api $(ARGS)

.PHONY: create_migration
create_migration: 			## Create a migration file, pass name using ARGS="<file_name>"
	@echo 'Creating migration files for ${ARGS}...'
	migrate create -seq -ext=.sql -dir=./migrations ${ARGS}

.PHONY: confirm_migration
confirm_migration:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: migration_up
migration_up: confirm_migration    	## Run database migrations to local postgres
	@echo 'Running up migrations...'
	migrate -path ./migrations -database=postgres://movies_api:123Change@localhost/movies?sslmode=disable up

.PHONY: migration_goto
migration_goto: confirm_migration 	## Go to migration version, pass name using ARGS=<number>
	@echo 'Migrating to version ${ARGS}...'
	migrate -path ./migrations -database=postgres://movies_api:123Change@localhost/movies?sslmode=disable goto ${ARGS}


.PHONY: audit
audit:					## Run go mod commands and staticcheck tools
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...
