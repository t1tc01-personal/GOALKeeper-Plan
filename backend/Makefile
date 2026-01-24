.PHONY: help build run test clean deps lint fmt

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building application..."
	@go build -o bin/goalkeeper-plan-api cmd/main.go

run: ## Run the application
	@echo "Running application..."
	@go run cmd/main.go api

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run ./... || echo "golangci-lint not installed, skipping..."

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

mod-verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	@go mod verify

