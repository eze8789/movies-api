help:  ## Show this help
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

run:   ## Run go program
	go run ./cmd/api/...

prun:  ## Run postgres db
	docker run --name postgresql-movies -p 5432:5432 -e POSTGRES_PASSWORD=test -v ~/go/src/github.com/eze8789/movies-api/postgres-volume:/var/lib/postgresql/data -d postgres
