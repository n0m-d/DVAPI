.PHONY: help up down stop restart build rebuild logs logs-backend logs-frontend ps \
	backend-shell frontend-shell db-shell clean prune \
	dev run  seed library-shell \
	migrate-up migrate-down migrate-status

COMPOSE := docker compose
BACKEND  := $(COMPOSE) exec backend
FRONTEND := $(COMPOSE) exec frontend
LIBRARY  := $(COMPOSE) exec library
SEED_TARGETS := $(if $(ARGS),$(ARGS),users courses enrollments assignments submissions announcements)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

## --- Stack lifecycle -------------------------------------------------------

up: ## Build (if needed), start stack, migrate, and seed
	$(COMPOSE) up -d --build
	$(BACKEND) go run ./cmd/migrate up
	$(BACKEND) go run ./cmd/seed $(SEED_TARGETS)
	cp backend/.env.example backend/.env
	cp frontend/.env.example frontend/.env

dev: up ## Alias for `up`

down: ## Stop and remove containers
	$(COMPOSE) down

stop: ## Stop containers without removing them
	$(COMPOSE) stop

restart: ## Restart all services
	$(COMPOSE) restart

build: ## Build all service images (does not start containers)
	$(COMPOSE) build

rebuild: ## Rebuild all service images without cache
	$(COMPOSE) build --no-cache

logs: ## Follow logs for all services
	$(COMPOSE) logs -f

logs-backend: ## Follow backend (air) logs
	$(COMPOSE) logs -f backend

logs-frontend: ## Follow frontend logs
	$(COMPOSE) logs -f frontend

ps: ## Show service status
	$(COMPOSE) ps

clean: ## Stop stack and remove named volumes (drops DB/Redis data)
	$(COMPOSE) down -v

prune: ## Remove stack, volumes, and locally built images
	$(COMPOSE) down -v --rmi local

## --- Shells ----------------------------------------------------------------

backend-shell: ## Open a shell in the backend container
	$(BACKEND) bash

frontend-shell: ## Open a shell in the frontend container
	$(FRONTEND) sh

db-shell: ## Open a psql shell in the postgres container
	$(COMPOSE) exec postgres psql -U postgres -d dvapi

library-shell: ## Open a shell in the library container
	$(LIBRARY) bash

## --- Backend commands (run inside the backend container) -------------------

seed: ## Seed development fake data (ARGS="users courses ..." to limit)
	$(BACKEND) go run ./cmd/seed $(SEED_TARGETS)

seed-status: ## Show current seed status
	$(BACKEND) go run ./cmd/seed status

migrate-up: ## Apply all pending migrations
	$(BACKEND) go run ./cmd/migrate up

migrate-down: ## Roll back the most recent migration
	$(BACKEND) go run ./cmd/migrate down

migrate-status: ## Show migration status
	$(BACKEND) go run ./cmd/migrate status

fe-reset: ## Rebuild and restart the frontend image
	$(COMPOSE) up -d --build --force-recreate frontend