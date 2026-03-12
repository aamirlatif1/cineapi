# ==================================================================================== #
# HELPERS
# ==================================================================================== #

.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run: run the cmd/api application
.PHONY: run
run:
	go run ./cmd/api

## build: build the cmd/api application
.PHONY: build
build:
	go build -o bin/api ./cmd/api

## tidy: tidy modfile and format .go files
.PHONY: tidy
tidy:
	go mod tidy
	go fmt ./...

## test: run all tests
.PHONY: test
test:
	go test ./...

## lint: run golangci-lint
.PHONY: lint
lint:
	golangci-lint run ./...

# ==================================================================================== #
# DATABASE
# ==================================================================================== #

## db/up: start the PostgreSQL container
.PHONY: db/up
db/up:
	docker compose up -d

## db/down: stop and remove the PostgreSQL container
.PHONY: db/down
db/down:
	docker compose down

## migrate/up: apply all pending migrations
.PHONY: migrate/up
migrate/up:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

## migrate/down: roll back the last migration
.PHONY: migrate/down
migrate/down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down 1
