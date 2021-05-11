help:  ## Show this help
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

gorun: ## Run go program, you can pass arguments using make run ARGS="-help"
	go run ./cmd/api/... $(ARGS)

pgrun: ## Run postgres db
	docker stop postgresql-movies && docker rm postgresql-movies && docker run --name postgresql-movies -p 5432:5432 -e POSTGRES_PASSWORD=test -v ~/go/src/github.com/eze8789/movies-api/postgres-volume:/var/lib/postgresql/data -d postgres

lint:  ## Run golangci-lint linter
	golangci-lint run ./...

build: ## Build go binary
	CGO_ENABLED=0 GOOS=linux go build -a -o ./bin/movies-api -ldflags="-s -w" ./cmd/...

run:   ## Run binary, you can pass arguments using make run ARGS="-help"
	./bin/movies-api $(ARGS)