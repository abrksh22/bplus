.PHONY: help build run test test-coverage lint fmt clean install dev docs

# Variables
BINARY_NAME=bplus
BUILD_DIR=bin
MAIN_PATH=cmd/bplus/main.go
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags="-s -w"

# Version information
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
BUILD_FLAGS=-ldflags="-s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

## help: Display this help message
help:
	@echo "b+ (Be Positive) - Development Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/##/  /' | column -t -s ':'

## build: Build the application binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✓ Built: $(BUILD_DIR)/$(BINARY_NAME)"

## run: Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	$(GO) run $(MAIN_PATH)

## test: Run all tests
test:
	@echo "Running tests..."
	$(GO) test $(GOFLAGS) ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

## test-verbose: Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	$(GO) test -v -race ./...

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

## lint: Run linters
lint:
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run --timeout=5m

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@which gofumpt > /dev/null && gofumpt -l -w . || echo "Install gofumpt for better formatting: go install mvdan.cc/gofumpt@latest"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@rm -rf dist/
	@echo "✓ Cleaned build artifacts"

## install: Install the binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(BUILD_FLAGS) $(MAIN_PATH)
	@echo "✓ Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

## dev: Run in development mode with hot reload (requires air)
dev:
	@which air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	air

## deps: Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "✓ Dependencies updated"

## deps-upgrade: Upgrade all dependencies to latest versions
deps-upgrade:
	@echo "Upgrading dependencies..."
	$(GO) get -u ./...
	$(GO) mod tidy
	@echo "✓ Dependencies upgraded"

## tools: Install development tools
tools:
	@echo "Installing development tools..."
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing air (live reload)..."
	@go install github.com/air-verse/air@latest
	@echo "Installing gofumpt (formatter)..."
	@go install mvdan.cc/gofumpt@latest
	@echo "Installing gotestsum (test runner)..."
	@go install gotest.tools/gotestsum@latest
	@echo "✓ Development tools installed"

## ci: Run CI checks (lint, test, build)
ci: lint test build
	@echo "✓ CI checks passed"

## release: Build for multiple platforms
release:
	@echo "Building for multiple platforms..."
	@which goreleaser > /dev/null || (echo "Installing goreleaser..." && go install github.com/goreleaser/goreleaser@latest)
	goreleaser build --snapshot --clean
	@echo "✓ Release builds complete"

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t bplus:$(VERSION) .
	@echo "✓ Docker image built: bplus:$(VERSION)"

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -it --rm bplus:$(VERSION)

## docs: Generate documentation
docs:
	@echo "Generating documentation..."
	@which godoc > /dev/null || (echo "Installing godoc..." && go install golang.org/x/tools/cmd/godoc@latest)
	@echo "Starting godoc server at http://localhost:6060"
	@echo "Browse to http://localhost:6060/pkg/github.com/abrksh22/bplus/"
	godoc -http=:6060

## check: Quick check before commit (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "✓ All checks passed - ready to commit!"

## all: Run fmt, vet, lint, test, and build
all: fmt vet lint test build
	@echo "✓ Build complete!"

# Default target
.DEFAULT_GOAL := help
