.PHONY: help dev build migrate seed clean test docker-up docker-down

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

dev: ## Run development server
	@echo "Starting development environment..."
	docker-compose up

build: ## Build the Go binary
	@echo "Building backend..."
	go build -o bin/pews ./cmd/pews

migrate: ## Run database migrations
	@echo "Running migrations..."
	go run ./cmd/pews migrate

seed: ## Seed the database with test data
	@echo "Seeding database..."
	@go run scripts/seed.go

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	docker-compose down -v

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

docker-up: ## Start Docker containers
	docker-compose up -d

docker-down: ## Stop Docker containers
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f

install: ## Install Go dependencies
	go mod download
	go mod tidy
