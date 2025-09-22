# FauxFoundry - Professional Makefile
# Created by copyleftdev - Building tools for developers, by developers
# https://github.com/copyleftdev/faux-foundry

# ============================================================================
# Configuration Variables
# ============================================================================

# Application
APP_NAME := fauxfoundry
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# Go Configuration
GO_VERSION := 1.21
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
CGO_ENABLED := 0

# Directories
BIN_DIR := bin
BUILD_DIR := build
DIST_DIR := dist
DOCS_DIR := docs
EXAMPLES_DIR := examples
OUTPUTS_DIR := outputs
MEDIA_DIR := media

# Build Configuration
MAIN_PACKAGE := ./cmd/fauxfoundry
LDFLAGS := -s -w \
	-X 'main.version=$(VERSION)' \
	-X 'main.buildDate=$(BUILD_DATE)' \
	-X 'main.gitCommit=$(GIT_COMMIT)' \
	-X 'main.gitBranch=$(GIT_BRANCH)'

# Test Configuration
TEST_TIMEOUT := 30m
COVERAGE_OUT := coverage.out
COVERAGE_HTML := coverage.html

# Docker Configuration
DOCKER_IMAGE := copyleftdev/fauxfoundry
DOCKER_TAG := $(VERSION)

# Release Configuration
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64
RELEASE_DIR := $(DIST_DIR)/release

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
PURPLE := \033[0;35m
CYAN := \033[0;36m
WHITE := \033[1;37m
NC := \033[0m # No Color

# ============================================================================
# Default Target
# ============================================================================

.DEFAULT_GOAL := help

# ============================================================================
# Help Target
# ============================================================================

.PHONY: help
help: ## Display this help message
	@echo "$(CYAN)FauxFoundry - Professional Makefile$(NC)"
	@echo "$(YELLOW)Created by copyleftdev$(NC)"
	@echo ""
	@echo "$(WHITE)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(BLUE)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(WHITE)Build Information:$(NC)"
	@echo "  Version:     $(VERSION)"
	@echo "  Go Version:  $(GO_VERSION)"
	@echo "  Platform:    $(GOOS)/$(GOARCH)"
	@echo "  Git Commit:  $(GIT_COMMIT)"
	@echo "  Git Branch:  $(GIT_BRANCH)"

# ============================================================================
# Environment Setup
# ============================================================================

.PHONY: setup
setup: ## Set up development environment
	@echo "$(CYAN)Setting up development environment...$(NC)"
	@go version | grep -q "go$(GO_VERSION)" || (echo "$(RED)Go $(GO_VERSION) required$(NC)" && exit 1)
	@go mod download
	@go mod tidy
	@mkdir -p $(BIN_DIR) $(BUILD_DIR) $(DIST_DIR) $(OUTPUTS_DIR)
	@echo "$(GREEN)Development environment ready!$(NC)"

.PHONY: deps
deps: ## Download and verify dependencies
	@echo "$(CYAN)Downloading dependencies...$(NC)"
	@go mod download
	@go mod verify
	@go mod tidy
	@echo "$(GREEN)Dependencies updated!$(NC)"

.PHONY: tools
tools: ## Install development tools
	@echo "$(CYAN)Installing development tools...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "$(GREEN)Development tools installed!$(NC)"

# ============================================================================
# Build Targets
# ============================================================================

.PHONY: build
build: ## Build the application for current platform
	@echo "$(CYAN)Building $(APP_NAME) for $(GOOS)/$(GOARCH)...$(NC)"
	@CGO_ENABLED=$(CGO_ENABLED) go build \
		-ldflags "$(LDFLAGS)" \
		-o $(BIN_DIR)/$(APP_NAME) \
		$(MAIN_PACKAGE)
	@echo "$(GREEN)Build complete: $(BIN_DIR)/$(APP_NAME)$(NC)"

.PHONY: build-all
build-all: ## Build for all supported platforms
	@echo "$(CYAN)Building for all platforms...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		output_name=$(APP_NAME)-$$GOOS-$$GOARCH; \
		if [ $$GOOS = "windows" ]; then output_name=$$output_name.exe; fi; \
		echo "Building for $$GOOS/$$GOARCH..."; \
		CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH go build \
			-ldflags "$(LDFLAGS)" \
			-o $(BUILD_DIR)/$$output_name \
			$(MAIN_PACKAGE); \
	done
	@echo "$(GREEN)Multi-platform build complete!$(NC)"

.PHONY: install
install: build ## Install the application to GOPATH/bin
	@echo "$(CYAN)Installing $(APP_NAME)...$(NC)"
	@go install -ldflags "$(LDFLAGS)" $(MAIN_PACKAGE)
	@echo "$(GREEN)Installation complete!$(NC)"

# ============================================================================
# Development Targets
# ============================================================================

.PHONY: run
run: ## Run the application with development flags
	@echo "$(CYAN)Running $(APP_NAME)...$(NC)"
	@go run $(MAIN_PACKAGE) --help

.PHONY: dev
dev: ## Run in development mode with hot reload
	@echo "$(CYAN)Starting development server...$(NC)"
	@go run $(MAIN_PACKAGE) tui --verbose

.PHONY: demo
demo: build ## Run a demonstration of key features
	@echo "$(CYAN)Running FauxFoundry demonstration...$(NC)"
	@./$(BIN_DIR)/$(APP_NAME) doctor
	@echo "$(YELLOW)Validating example specifications...$(NC)"
	@./$(BIN_DIR)/$(APP_NAME) validate $(EXAMPLES_DIR)/*.yaml
	@echo "$(YELLOW)Generating sample data...$(NC)"
	@./$(BIN_DIR)/$(APP_NAME) generate --spec $(EXAMPLES_DIR)/medical-demo.yaml --count 5 --output $(OUTPUTS_DIR)/demo.jsonl
	@echo "$(GREEN)Demo complete! Check $(OUTPUTS_DIR)/demo.jsonl$(NC)"

# ============================================================================
# Testing Targets
# ============================================================================

.PHONY: test
test: ## Run all tests
	@echo "$(CYAN)Running tests...$(NC)"
	@go test -v -timeout $(TEST_TIMEOUT) ./...

.PHONY: test-race
test-race: ## Run tests with race detection
	@echo "$(CYAN)Running tests with race detection...$(NC)"
	@go test -v -race -timeout $(TEST_TIMEOUT) ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "$(CYAN)Running tests with coverage...$(NC)"
	@go test -v -coverprofile=$(COVERAGE_OUT) -covermode=atomic ./...
	@go tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@echo "$(GREEN)Coverage report generated: $(COVERAGE_HTML)$(NC)"

.PHONY: test-integration
test-integration: build ## Run integration tests
	@echo "$(CYAN)Running integration tests...$(NC)"
	@go test -v -tags=integration -timeout $(TEST_TIMEOUT) ./tests/integration/...

.PHONY: test-e2e
test-e2e: build ## Run end-to-end tests
	@echo "$(CYAN)Running end-to-end tests...$(NC)"
	@go test -v -tags=e2e -timeout $(TEST_TIMEOUT) ./tests/e2e/...

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "$(CYAN)Running benchmarks...$(NC)"
	@go test -v -bench=. -benchmem ./...

# ============================================================================
# Quality Assurance
# ============================================================================

.PHONY: lint
lint: ## Run linter
	@echo "$(CYAN)Running linter...$(NC)"
	@golangci-lint run ./...

.PHONY: fmt
fmt: ## Format code
	@echo "$(CYAN)Formatting code...$(NC)"
	@go fmt ./...
	@goimports -w .

.PHONY: vet
vet: ## Run go vet
	@echo "$(CYAN)Running go vet...$(NC)"
	@go vet ./...

.PHONY: security
security: ## Run security scan
	@echo "$(CYAN)Running security scan...$(NC)"
	@gosec ./...

.PHONY: quality
quality: fmt vet lint security test ## Run all quality checks

# ============================================================================
# Documentation
# ============================================================================

.PHONY: docs
docs: ## Generate documentation
	@echo "$(CYAN)Generating documentation...$(NC)"
	@go doc -all ./... > $(DOCS_DIR)/api.txt
	@echo "$(GREEN)Documentation generated in $(DOCS_DIR)/$(NC)"

.PHONY: docs-serve
docs-serve: ## Serve documentation locally
	@echo "$(CYAN)Serving documentation...$(NC)"
	@godoc -http=:6060 -play

.PHONY: examples
examples: build ## Validate all example specifications
	@echo "$(CYAN)Validating example specifications...$(NC)"
	@for spec in $(EXAMPLES_DIR)/*.yaml; do \
		echo "Validating $$spec..."; \
		./$(BIN_DIR)/$(APP_NAME) validate "$$spec" || exit 1; \
	done
	@echo "$(GREEN)All examples validated successfully!$(NC)"

# ============================================================================
# Docker Targets
# ============================================================================

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(CYAN)Building Docker image...$(NC)"
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(NC)"

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "$(CYAN)Running Docker container...$(NC)"
	@docker run --rm -it \
		-v $(PWD)/$(EXAMPLES_DIR):/app/examples \
		-v $(PWD)/$(OUTPUTS_DIR):/app/outputs \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-push
docker-push: docker-build ## Push Docker image to registry
	@echo "$(CYAN)Pushing Docker image...$(NC)"
	@docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	@docker push $(DOCKER_IMAGE):latest
	@echo "$(GREEN)Docker image pushed!$(NC)"

# ============================================================================
# Release Targets
# ============================================================================

.PHONY: release-prepare
release-prepare: quality test-coverage ## Prepare for release
	@echo "$(CYAN)Preparing release $(VERSION)...$(NC)"
	@mkdir -p $(RELEASE_DIR)
	@echo "$(GREEN)Release preparation complete!$(NC)"

.PHONY: release-build
release-build: release-prepare build-all ## Build release artifacts
	@echo "$(CYAN)Building release artifacts...$(NC)"
	@mkdir -p $(RELEASE_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		binary_name=$(APP_NAME)-$$GOOS-$$GOARCH; \
		if [ $$GOOS = "windows" ]; then binary_name=$$binary_name.exe; fi; \
		archive_name=$(APP_NAME)-$(VERSION)-$$GOOS-$$GOARCH; \
		if [ $$GOOS = "windows" ]; then \
			zip -j $(RELEASE_DIR)/$$archive_name.zip $(BUILD_DIR)/$$binary_name README.md LICENSE; \
		else \
			tar -czf $(RELEASE_DIR)/$$archive_name.tar.gz -C $(BUILD_DIR) $$binary_name -C .. README.md LICENSE; \
		fi; \
		echo "Created $$archive_name"; \
	done
	@echo "$(GREEN)Release artifacts built in $(RELEASE_DIR)$(NC)"

.PHONY: release-checksums
release-checksums: ## Generate checksums for release artifacts
	@echo "$(CYAN)Generating checksums...$(NC)"
	@cd $(RELEASE_DIR) && sha256sum * > checksums.txt
	@echo "$(GREEN)Checksums generated!$(NC)"

.PHONY: release
release: release-build release-checksums ## Create complete release
	@echo "$(CYAN)Release $(VERSION) ready!$(NC)"
	@ls -la $(RELEASE_DIR)

# ============================================================================
# Maintenance Targets
# ============================================================================

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(CYAN)Cleaning build artifacts...$(NC)"
	@rm -rf $(BIN_DIR) $(BUILD_DIR) $(DIST_DIR)
	@rm -f $(COVERAGE_OUT) $(COVERAGE_HTML)
	@go clean -cache -testcache -modcache
	@echo "$(GREEN)Clean complete!$(NC)"

.PHONY: clean-outputs
clean-outputs: ## Clean generated output files
	@echo "$(CYAN)Cleaning output files...$(NC)"
	@rm -rf $(OUTPUTS_DIR)/*.jsonl $(OUTPUTS_DIR)/*.jsonl.gz
	@echo "$(GREEN)Output files cleaned!$(NC)"

.PHONY: update
update: ## Update dependencies
	@echo "$(CYAN)Updating dependencies...$(NC)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)Dependencies updated!$(NC)"

# ============================================================================
# CI/CD Targets
# ============================================================================

.PHONY: ci
ci: setup quality test-coverage ## Run CI pipeline
	@echo "$(GREEN)CI pipeline completed successfully!$(NC)"

.PHONY: cd
cd: ci release ## Run CD pipeline
	@echo "$(GREEN)CD pipeline completed successfully!$(NC)"

# ============================================================================
# Utility Targets
# ============================================================================

.PHONY: version
version: ## Display version information
	@echo "$(CYAN)FauxFoundry Version Information$(NC)"
	@echo "Version:     $(VERSION)"
	@echo "Build Date:  $(BUILD_DATE)"
	@echo "Git Commit:  $(GIT_COMMIT)"
	@echo "Git Branch:  $(GIT_BRANCH)"
	@echo "Go Version:  $(shell go version)"
	@echo "Platform:    $(GOOS)/$(GOARCH)"

.PHONY: size
size: build ## Display binary size information
	@echo "$(CYAN)Binary Size Information$(NC)"
	@ls -lh $(BIN_DIR)/$(APP_NAME)
	@file $(BIN_DIR)/$(APP_NAME)

.PHONY: deps-graph
deps-graph: ## Generate dependency graph
	@echo "$(CYAN)Generating dependency graph...$(NC)"
	@go mod graph | dot -T png -o $(DOCS_DIR)/deps.png
	@echo "$(GREEN)Dependency graph saved to $(DOCS_DIR)/deps.png$(NC)"

# ============================================================================
# Health Check Targets
# ============================================================================

.PHONY: health
health: build ## Run health checks
	@echo "$(CYAN)Running health checks...$(NC)"
	@./$(BIN_DIR)/$(APP_NAME) doctor --verbose
	@echo "$(GREEN)Health checks complete!$(NC)"

.PHONY: smoke-test
smoke-test: build ## Run smoke tests
	@echo "$(CYAN)Running smoke tests...$(NC)"
	@./$(BIN_DIR)/$(APP_NAME) --version
	@./$(BIN_DIR)/$(APP_NAME) --help
	@./$(BIN_DIR)/$(APP_NAME) validate $(EXAMPLES_DIR)/medical-demo.yaml
	@echo "$(GREEN)Smoke tests passed!$(NC)"

# ============================================================================
# Performance Targets
# ============================================================================

.PHONY: profile
profile: build ## Run performance profiling
	@echo "$(CYAN)Running performance profiling...$(NC)"
	@go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...
	@echo "$(GREEN)Profiling complete! Use 'go tool pprof' to analyze$(NC)"

.PHONY: stress-test
stress-test: build ## Run stress tests
	@echo "$(CYAN)Running stress tests...$(NC)"
	@./$(BIN_DIR)/$(APP_NAME) generate --spec $(EXAMPLES_DIR)/medical-demo.yaml --count 1000 --output $(OUTPUTS_DIR)/stress-test.jsonl
	@echo "$(GREEN)Stress test complete!$(NC)"

# ============================================================================
# Special Targets
# ============================================================================

# Prevent make from deleting intermediate files
.PRECIOUS: $(BUILD_DIR)/% $(RELEASE_DIR)/%

# Ensure these targets always run
.PHONY: help setup deps tools build build-all install run dev demo test test-race test-coverage test-integration test-e2e benchmark lint fmt vet security quality docs docs-serve examples docker-build docker-run docker-push release-prepare release-build release-checksums release clean clean-outputs update ci cd version size deps-graph health smoke-test profile stress-test
