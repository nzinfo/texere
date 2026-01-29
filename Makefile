# Texere Makefile

.PHONY: all build test clean benchmark lint install run

# Variables
BINARY_NAME=texere-server
CLI_NAME=texere-cli
BUILD_DIR=bin
CMD_DIR=cmd
PKG_DIR=pkg

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%S)"
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Default target
all: lint build

## build: Build the server and CLI
build:
	@echo "Building Texere..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/texere-server/main.go
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(CLI_NAME) $(CMD_DIR)/texere-cli/main.go
	@echo "Build complete!"

## build-server: Build only the server
build-server:
	@echo "Building Texere server..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/texere-server/main.go
	@echo "Server build complete!"

## build-cli: Build only the CLI
build-cli:
	@echo "Building Texere CLI..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(CLI_NAME) $(CMD_DIR)/texere-cli/main.go
	@echo "CLI build complete!"

## test: Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Tests complete! Coverage report: coverage.html"

## test-unit: Run unit tests only
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v ./$(PKG_DIR)/...

## test-integration: Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v ./test/integration/...

## benchmark: Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./$(PKG_DIR)/... | tee benchmark.txt

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html benchmark.txt
	@echo "Clean complete!"

## lint: Run linters
lint:
	@echo "Running linters..."
	$(GOFMT) -l .
	$(GOLINT) run
	@echo "Linting complete!"

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -w .
	@echo "Formatting complete!"

## tidy: Tidy go.mod
tidy:
	@echo "Tidying go.mod..."
	$(GOMOD) tidy
	@echo "Tidy complete!"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	@echo "Dependencies downloaded!"

## install: Install binaries
install:
	@echo "Installing Texere..."
	$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) $(CMD_DIR)/texere-server/main.go
	$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(CLI_NAME) $(CMD_DIR)/texere-cli/main.go
	@echo "Installation complete!"

## run: Run the server
run: build-server
	@echo "Starting Texere server..."
	./$(BUILD_DIR)/$(BINARY_NAME)

## run-dev: Run in development mode with hot reload
run-dev:
	@echo "Starting development server..."
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t texere:latest -f deployments/docker/Dockerfile .
	@echo "Docker image built!"

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 texere:latest

## docker-compose-up: Start services with docker-compose
docker-compose-up:
	@echo "Starting services with docker-compose..."
	docker-compose -f deployments/docker/docker-compose.yml up -d

## docker-compose-down: Stop docker-compose services
docker-compose-down:
	@echo "Stopping docker-compose services..."
	docker-compose -f deployments/docker/docker-compose.yml down

## help: Show this help message
help:
	@echo "Texere Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
