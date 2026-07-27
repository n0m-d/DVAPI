.PHONY: help run dev build test lint sqlc migrate-up migrate-down migrate-status migrate-create seed seed-status tools tidy docker-up docker-down

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

run: build ## Run the API server
	./bin/server

dev: ## Run with live reload (air)
	air

build: ## Build binaries
	@mkdir -p bin
	go build -o bin/server ./cmd/server
	go build -o bin/migrate ./cmd/migrate
	go build -o bin/seed ./cmd/seed

seed: build ## Seed development fake data (ARGS="users courses ..." to limit)
	./bin/seed $(if $(ARGS),$(ARGS),users courses enrollments assignments announcements submissions)

seed-status: build ## Show current seed status
	./bin/seed status

test: ## Run tests
	go test ./... -race -count=1

lint: ## Run golangci-lint
	golangci-lint run ./...

sqlc: ## Generate type-safe SQL code
	sqlc generate

migrate-up: build ## Apply all pending migrations
	./bin/migrate up

migrate-down: build ## Roll back the most recent migration
	./bin/migrate down

migrate-status: build ## Show migration status
	./bin/migrate status

migrate-create: ## Create a new migration (usage: make migrate-create NAME=add_foo)
	@test -n "$(NAME)" || (echo "NAME is required, e.g. make migrate-create NAME=add_foo" && exit 1)
	goose -dir migrations create $(NAME) sql

tidy: ## Tidy go modules
	go mod tidy

tools: ## Install dev tools
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

swag:
	swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal