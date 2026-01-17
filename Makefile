.PHONY: test test-unit test-integration test-coverage lint build clean docker-build docker-run help

# Variables
BINARY_NAME=halerbackend
DOCKER_IMAGE=halerbackend:latest
COVERAGE_FILE=coverage.out

# Colors for terminal output
GREEN=\033[0;32m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## Show this help message
	@echo "$(GREEN)HalerBackend - Available commands:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

install: ## Install dependencies
	@echo "$(GREEN)Installing dependencies...$(NC)"
	go mod download
	go mod tidy

test: test-unit ## Run all tests (unit only, no database required)

test-unit: ## Run unit tests
	@echo "$(GREEN)Running unit tests...$(NC)"
	go test -short -race -v \
		./internal/auth/domain \
		./internal/auth/service \
		./internal/leaderboard/domain \
		./internal/strategy \
		./internal/shared/utils \
		./internal/shared/middleware

test-integration: ## Run integration tests (requires PostgreSQL and Redis)
	@echo "$(GREEN)Running integration tests...$(NC)"
	@echo "$(YELLOW)Make sure PostgreSQL and Redis are running!$(NC)"
	go test -v -race \
		./internal/service \
		./internal/integration \
		./internal/shared/repository

test-all: ## Run all tests including integration
	@echo "$(GREEN)Running all tests...$(NC)"
	make test-unit
	make test-integration

test-coverage: ## Run tests with coverage report
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	go test -short -coverprofile=$(COVERAGE_FILE) -covermode=atomic \
		./internal/auth/domain \
		./internal/auth/service \
		./internal/leaderboard/domain \
		./internal/strategy \
		./internal/shared/utils \
		./internal/shared/middleware
	@echo "$(GREEN)Generating coverage report...$(NC)"
	go tool cover -func=$(COVERAGE_FILE)
	@echo "$(GREEN)Generating HTML coverage report...$(NC)"
	go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

test-coverage-all: ## Run all tests with full coverage
	@echo "$(GREEN)Running all tests with coverage...$(NC)"
	go test -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./internal/...
	go tool cover -func=$(COVERAGE_FILE)
	go tool cover -html=$(COVERAGE_FILE) -o coverage.html

lint: ## Run linter
	@echo "$(GREEN)Running linter...$(NC)"
	golangci-lint run --timeout=5m

fmt: ## Format code
	@echo "$(GREEN)Formatting code...$(NC)"
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	@echo "$(GREEN)Running go vet...$(NC)"
	go vet ./...

check: fmt vet lint test-unit ## Run all checks (format, vet, lint, test)

build: ## Build the application
	@echo "$(GREEN)Building application...$(NC)"
	go build -o bin/server ./cmd/server
	go build -o bin/simulator ./cmd/simulator
	@echo "$(GREEN)Build complete! Binaries in bin/$(NC)"

run: ## Run the server
	@echo "$(GREEN)Starting server...$(NC)"
	go run cmd/server/main.go

run-simulator: ## Run the simulator
	@echo "$(GREEN)Starting simulator...$(NC)"
	go run cmd/simulator/main.go

clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning...$(NC)"
	rm -rf bin/
	rm -f $(COVERAGE_FILE) coverage.html
	go clean

docker-build: ## Build Docker image
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container
	@echo "$(GREEN)Running Docker container...$(NC)"
	docker-compose up -d

docker-stop: ## Stop Docker containers
	@echo "$(GREEN)Stopping Docker containers...$(NC)"
	docker-compose down

docker-logs: ## Show Docker logs
	docker-compose logs -f

benchmark: ## Run benchmarks
	@echo "$(GREEN)Running benchmarks...$(NC)"
	go test -bench=. -benchmem -run=^$$ ./internal/service

ci: install check test-all ## Run CI pipeline locally

seed: ## Seed the database
	@echo "$(GREEN)Seeding database...$(NC)"
	go run cmd/seed/main.go

migrate: ## Run database migrations
	@echo "$(GREEN)Running migrations...$(NC)"
	# Add migration command here

.DEFAULT_GOAL := help
