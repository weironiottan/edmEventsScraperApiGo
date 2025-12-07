.PHONY: default test test-verbose test-cover test-coverage-html clean build run help

# Default target
.DEFAULT_GOAL := default

## default: Run fmt and test
default: fmt test
	@echo "Format and tests complete!"

## help: Display this help message
help:
	@echo "Available targets:"
	@echo "  make test              - Run all tests"
	@echo "  make test-verbose      - Run tests with verbose output"
	@echo "  make test-cover        - Run tests with coverage report"
	@echo "  make test-coverage-html - Generate HTML coverage report"
	@echo "  make build             - Build the application binary"
	@echo "  make run               - Run the application"
	@echo "  make clean             - Remove build artifacts and test cache"
	@echo "  make fmt               - Format Go code"
	@echo "  make lint              - Run go vet"
	@echo "  make deps              - Download dependencies"
	@echo "  make docker-build      - Build Docker image"
	@echo "  make docker-run        - Run Docker container"

## test: Run all tests
test:
	@echo "Running tests..."
	go test ./cmd

## test-verbose: Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	go test -v ./cmd

## test-cover: Run tests with coverage report
test-cover:
	@echo "Running tests with coverage..."
	go test -cover ./cmd

## test-coverage-html: Generate HTML coverage report
test-coverage-html:
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./cmd
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## test-coverage-func: Show coverage by function
test-coverage-func:
	@echo "Generating function coverage report..."
	go test -coverprofile=coverage.out ./cmd
	go tool cover -func=coverage.out

## build: Build the application binary
build:
	@echo "Building application..."
	go build -v -o edmEventsScraperJob ./cmd
	@echo "Binary created: edmEventsScraperJob"

## run: Run the application
run: build
	@echo "Running application..."
	./edmEventsScraperJob

## clean: Remove build artifacts and test cache
clean:
	@echo "Cleaning up..."
	rm -f edmEventsScraperJob
	rm -f coverage.out coverage.html
	go clean -testcache
	@echo "Clean complete"

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	go fmt ./...

## lint: Run go vet
lint:
	@echo "Running go vet..."
	go vet ./...

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t edm-events-scraper .

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run --env-file .env edm-events-scraper

## pre-commit: Run pre-commit hooks
pre-commit:
	@echo "Running pre-commit hooks..."
	pre-commit run --all-files

## all: Run fmt, lint, and test
all: fmt lint test
	@echo "All checks passed!"
