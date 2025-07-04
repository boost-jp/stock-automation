# Stock Automation Backend Makefile
# Go version: 1.24.4

.PHONY: help install-tools build test test-coverage test-integration clean lint fmt vet security docker-build docker-up docker-down dev run migrate

# Variables
BINARY_NAME=stock-automation
GO_VERSION=1.24.4
DOCKER_COMPOSE_FILE=docker/docker-compose.yml
COVERAGE_FILE=coverage.out
MAIN_FILE=cmd/main.go

# Default target
help: ## Display help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Install tools
install-tools: ## Install required tools from tools workspace
	@echo "Installing tools..."
	@cd tools && go install github.com/aarondl/sqlboiler/v4
	@cd tools && go install github.com/aarondl/sqlboiler/v4/drivers/sqlboiler-mysql
	@cd tools && go install github.com/golangci/golangci-lint/cmd/golangci-lint
	@cd tools && go install github.com/sqldef/sqldef/cmd/mysqldef
	@cd tools && go install golang.org/x/tools/cmd/goimports
	@echo "Tools installed successfully"

# Build
build: ## Build the application
	@echo "Building application..."
	@go build -o $(BINARY_NAME) $(MAIN_FILE)
	@echo "Build completed: $(BINARY_NAME)"

build-linux: ## Build for Linux
	@echo "Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux $(MAIN_FILE)
	@echo "Linux build completed: $(BINARY_NAME)-linux"

# Testing
test: ## Run unit tests
	@echo "Running unit tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=$(COVERAGE_FILE) ./...
	@go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-integration: docker-up test-db-migrate ## Run integration tests
	@echo "Running integration tests..."
	@sleep 10  # Wait for database to be ready
	@DB_NAME=$(TEST_DB_NAME) go test -v -tags=integration ./...

test-all: test test-integration ## Run all tests

# Code Quality
lint: ## Run linter
	@echo "Running linter..."
	@go tool golangci-lint run --fix

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@go tool goimports -w .

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

security: ## Run security scan
	@echo "Running security scan..."
	@gosec ./...

check: fmt vet lint security ## Run all code quality checks

# Docker
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):latest .

docker-up: ## Start Docker containers
	@echo "Starting Docker containers..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) up -d

docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) down

docker-logs: ## Show Docker logs
	@docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

# Development
dev: docker-up ## Start development environment
	@echo "Starting development environment..."
	@sleep 5  # Wait for database to be ready
	@go run $(MAIN_FILE)

run: ## Run the application
	@echo "Running application..."
	@go run $(MAIN_FILE)

# Database
DB_HOST ?= localhost
DB_PORT ?= 3309
DB_USER ?= root
DB_PASS ?= password
DB_NAME ?= stock_automation
TEST_DB_NAME ?= stock_automation_test

db-setup: ## Create database if not exists
	@echo "Creating database if not exists..."
	@docker exec stock-automation-mysql mysql -u$(DB_USER) -p$(DB_PASS) -e "CREATE DATABASE IF NOT EXISTS $(DB_NAME) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

db-migrate: ## Apply database schema using mysqldef
	@echo "Applying database schema..."
	@mysqldef -u$(DB_USER) -p$(DB_PASS) -h$(DB_HOST) -P$(DB_PORT) $(DB_NAME) < schema.sql

test-db-setup: ## Create test database if not exists
	@echo "Creating test database if not exists..."
	@docker exec stock-automation-mysql mysql -u$(DB_USER) -p$(DB_PASS) -e "CREATE DATABASE IF NOT EXISTS $(TEST_DB_NAME) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

test-db-migrate: test-db-setup ## Apply database schema to test database
	@echo "Applying database schema to test database..."
	@mysqldef -u$(DB_USER) -p$(DB_PASS) -h$(DB_HOST) -P$(DB_PORT) $(TEST_DB_NAME) < schema.sql

gen-sqlboiler: ## Generate SQLBoiler models
	@echo "Generating SQLBoiler models..."
	@echo "Step 1: Generate DAO package..."
	@rm -rf app/infrastructure/dao/*.go
	@go tool sqlboiler mysql -o app/infrastructure/dao -p dao --no-driver-templates --wipe --templates templates/sqlboiler/main --templates templates/sqlboiler/additional
	@echo "Step 2: Generate domain models..."
	@go tool sqlboiler mysql -o app/infrastructure/dto -p models --no-driver-templates --wipe --templates templates/sqlboiler/models
	@mv -n app/infrastructure/dto/*.go app/domain/models 2>/dev/null || true
	@rm -rf app/infrastructure/dto/*.go

migrate: db-migrate ## Run database migrations (alias for db-migrate)

# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME) $(BINARY_NAME)-linux
	@rm -f $(COVERAGE_FILE) coverage.html
	@rm -rf app/models_generated
	@go clean

clean-all: clean docker-down ## Clean everything including Docker containers

# Dependencies
mod-tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	@go mod tidy

mod-download: ## Download go modules
	@echo "Downloading go modules..."
	@go mod download

# Hot reload for development (requires air)
install-air: ## Install air for hot reload
	@go install github.com/cosmtrek/air@latest

hot-reload: install-air docker-up ## Start hot reload development
	@echo "Starting hot reload development..."
	@air

# Generate documentation
docs: ## Generate API documentation
	@echo "Generating API documentation..."
	@swag init -g $(MAIN_FILE)

# Release
release: clean test check build ## Build release version

# CI/CD targets
ci-test: install-tools test-coverage security ## Run CI tests

ci-build: install-tools build ## Run CI build

# Health check
health: ## Check if application is running
	@echo "Checking application health..."
	@curl -f http://localhost:8080/health || echo "Application is not running"

# Show project info
info: ## Show project information
	@echo "Project: Stock Automation Backend"
	@echo "Go Version: $(GO_VERSION)"
	@echo "Binary: $(BINARY_NAME)"
	@echo "Main File: $(MAIN_FILE)"
	@echo "Docker Compose: $(DOCKER_COMPOSE_FILE)"
