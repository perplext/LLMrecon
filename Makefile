# LLM Red Team Makefile

# Variables
BINARY_NAME=LLMrecon
VERSION?=0.1.0
BUILD_DIR=build
MAIN_PATH=src/main.go
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet
GOLINT=golangci-lint

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)"

# OS specific variables
ifeq ($(OS),Windows_NT)
    BINARY_EXT=.exe
else
    BINARY_EXT=
endif

# Default target
.DEFAULT_GOAL := help

.PHONY: help
help: ## Display this help message
	@echo "LLM Red Team - Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: all
all: clean deps test build ## Clean, install dependencies, test, and build

.PHONY: deps
deps: ## Download and tidy dependencies
	$(GOMOD) download
	$(GOMOD) tidy

.PHONY: build
build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT)"

.PHONY: build-all
build-all: ## Build for multiple platforms
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	# macOS ARM64
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Multi-platform build complete"

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

.PHONY: test-short
test-short: ## Run short tests
	$(GOTEST) -v -short ./...

.PHONY: coverage
coverage: test ## Generate coverage report
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: benchmark
benchmark: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

.PHONY: lint
lint: ## Run linter
	@echo "Running linter..."
	$(GOLINT) run ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...

.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

.PHONY: install
install: build ## Install binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(MAIN_PATH)

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t LLMrecon:$(VERSION) .

.PHONY: docker-push
docker-push: ## Push Docker image
	@echo "Pushing Docker image..."
	docker tag LLMrecon:$(VERSION) LLMrecon:latest
	docker push LLMrecon:$(VERSION)
	docker push LLMrecon:latest

.PHONY: release
release: clean test build-all ## Create release artifacts
	@echo "Creating release artifacts..."
	@mkdir -p releases/$(VERSION)
	# Create archives for each platform
	cd $(BUILD_DIR) && tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	cd $(BUILD_DIR) && tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64
	cd $(BUILD_DIR) && tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64
	cd $(BUILD_DIR) && tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	cd $(BUILD_DIR) && zip ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	# Create checksums
	cd releases/$(VERSION) && shasum -a 256 * > checksums.txt
	@echo "Release artifacts created in releases/$(VERSION)"

.PHONY: run
run: build ## Build and run the binary
	$(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT)

.PHONY: dev
dev: ## Run with hot reload (requires air)
	air -c .air.toml

# Check if all tools are installed
.PHONY: check-tools
check-tools: ## Check if required tools are installed
	@echo "Checking tools..."
	@which $(GOCMD) > /dev/null || (echo "Go is not installed" && exit 1)
	@which $(GOLINT) > /dev/null || echo "golangci-lint is not installed (optional)"
	@which docker > /dev/null || echo "Docker is not installed (optional)"
	@which air > /dev/null || echo "Air is not installed (optional for hot reload)"
	@echo "Basic tools check complete"

# Initialize project for development
.PHONY: init
init: check-tools deps ## Initialize project for development
	@echo "Project initialized successfully!"

# Update dependencies
.PHONY: update
update: ## Update all dependencies
	$(GOGET) -u ./...
	$(GOMOD) tidy