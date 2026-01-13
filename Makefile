# Variables
DOCKER_COMPOSE = docker compose
DOCKER = docker 
PROJECT_NAME = orders
ENV_FILE = .env
linter = golangci-lint

# Commands
.PHONY: help
help: ## Show available commands
	@echo "[LOCAL]: Available commands:"
	@echo "  make lint      - Start $(linter) at main-service and producer-service"
	@echo "  make lint-fix  - Start $(linter) with fix at main-service and producer-service"
	@echo "[DOCKER COMPOSE]: Available commands:"
	@echo "  make up        - Start all services in detached mode"
	@echo "  make down      - Stop and remove all services"
	@echo "  make restart   - Restart all services"
	@echo "  make logs      - Show all logs (follow mode)"
	@echo "  make build     - Rebuild all images (no cache)"
	@echo "  make rebuild   - Down, rebuild and up all services"
	@echo "  make clean     - Clear all (images, volumes, cache)"
	@echo "  make ps        - Show status of containers"
	@echo "  make status    - Alias for ps"
	@echo "  make help      - Show this help"

.PHONY: lint
lint: ## Check golangci lint errors
	@echo "Start $(linter) at main-service..."
	cd main-service && $(linter) run ./...
	@echo "Start $(linter) at producer-service..."
	cd producer-service && $(linter) run ./...


.PHONY: lint-fix
lint-fix: ## Check and fix golangci lint errors
	@echo "Start $(linter) with fix at main-service..."
	cd main-service && $(linter) --fix run ./...
	@echo "Start $(linter) with fix at producer-service..."
	cd producer-service && $(linter) --fix run ./...


.PHONY: up
up: ## Detached run all services
	$(DOCKER_COMPOSE) up -d

.PHONY: down
down: ## Stop and delete all services
	$(DOCKER_COMPOSE) down

.PHONY: restart
restart: ## Restart all services
	$(DOCKER_COMPOSE) down
	$(DOCKER_COMPOSE) up -d

.PHONY: logs
logs: ## Show all logs
	$(DOCKER_COMPOSE) logs -f


# Development

.PHONY: build
build:  ## Rebuild all images
	$(DOCKER_COMPOSE) build --no-cache

.PHONY: rebuild
rebuild: down build up  ## Rebuild and start all images


# Cleaning

.PHONY: clean
clean: ## Clear all (images, volumes, cache)
	$(DOCKER) system prune -af
	$(DOCKER) volume prune -f


# Status

.PHONY: ps
ps:  ## Show status of containers
	$(DOCKER_COMPOSE) ps

.PHONY: status
status: ps