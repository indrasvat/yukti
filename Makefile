# Makefile for Yukti
# Terminal User Interface for Google Apps Script

# ══════════════════════════════════════════════════════════════════════════════
# Variables
# ══════════════════════════════════════════════════════════════════════════════

BINARY_NAME := yukti
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION := $(shell go version | cut -d' ' -f3)

# Build flags
LDFLAGS := -s -w \
	-X yukti/internal/buildinfo.Version=$(VERSION) \
	-X yukti/internal/buildinfo.Commit=$(COMMIT) \
	-X yukti/internal/buildinfo.BuildDate=$(BUILD_DATE) \
	-X yukti/internal/buildinfo.GoVersion=$(GO_VERSION)

# Directories
BIN_DIR := bin
DIST_DIR := dist
COVERAGE_DIR := coverage

# Tools
GOLANGCI_LINT := golangci-lint
GORELEASER := goreleaser
GOTESTSUM := $(shell command -v gotestsum 2> /dev/null)

# Colors for output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[34m
COLOR_MAGENTA := \033[35m

# ══════════════════════════════════════════════════════════════════════════════
# Default target
# ══════════════════════════════════════════════════════════════════════════════

.DEFAULT_GOAL := help

# ══════════════════════════════════════════════════════════════════════════════
# Help
# ══════════════════════════════════════════════════════════════════════════════

.PHONY: help
help: ## Show this help message
	@echo ""
	@echo "$(COLOR_BOLD)⚡ Yukti - TUI for Google Apps Script$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Usage:$(COLOR_RESET)"
	@echo "  make $(COLOR_GREEN)<target>$(COLOR_RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(COLOR_GREEN)%-18s$(COLOR_RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""

# ══════════════════════════════════════════════════════════════════════════════
# Development
# ══════════════════════════════════════════════════════════════════════════════

.PHONY: build
build: ## Build the binary
	@echo "$(COLOR_BLUE)▶ Building $(BINARY_NAME)...$(COLOR_RESET)"
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/yukti
	@echo "$(COLOR_GREEN)✓ Built $(BIN_DIR)/$(BINARY_NAME)$(COLOR_RESET)"

.PHONY: build-all
build-all: ## Build for all platforms
	@echo "$(COLOR_BLUE)▶ Building for all platforms...$(COLOR_RESET)"
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/yukti
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/yukti
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/yukti
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/yukti
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/yukti
	@echo "$(COLOR_GREEN)✓ Built all platforms in $(DIST_DIR)/$(COLOR_RESET)"

.PHONY: install
install: build ## Install to GOPATH/bin
	@echo "$(COLOR_BLUE)▶ Installing $(BINARY_NAME)...$(COLOR_RESET)"
	go install -ldflags "$(LDFLAGS)" ./cmd/yukti
	@echo "$(COLOR_GREEN)✓ Installed to $$(go env GOPATH)/bin/$(BINARY_NAME)$(COLOR_RESET)"

.PHONY: run
run: build ## Run the TUI (development)
	@$(BIN_DIR)/$(BINARY_NAME)

.PHONY: dev
dev: ## Run in development mode with live reload (requires air)
	@if command -v air > /dev/null; then \
		air -c .air.toml; \
	else \
		echo "air not installed. Run: go install github.com/air-verse/air@latest"; \
		exit 1; \
	fi

# ══════════════════════════════════════════════════════════════════════════════
# Testing
# ══════════════════════════════════════════════════════════════════════════════

.PHONY: test
test: ## Run tests
	@echo "$(COLOR_BLUE)▶ Running tests...$(COLOR_RESET)"
ifdef GOTESTSUM
	gotestsum --format pkgname-and-test-fails --format-icons hivis -- -race -shuffle=on ./...
else
	go test -race -shuffle=on ./...
endif
	@echo "$(COLOR_GREEN)✓ Tests passed$(COLOR_RESET)"

.PHONY: test-v
test-v: ## Run tests with verbose output (BDD-style)
	@echo "$(COLOR_BLUE)▶ Running tests (verbose)...$(COLOR_RESET)"
ifdef GOTESTSUM
	gotestsum --format testdox --format-icons hivis -- -race -shuffle=on ./...
else
	go test -race -shuffle=on -v ./...
endif

.PHONY: test-cover
test-cover: ## Run tests with coverage
	@echo "$(COLOR_BLUE)▶ Running tests with coverage...$(COLOR_RESET)"
	@mkdir -p $(COVERAGE_DIR)
ifdef GOTESTSUM
	gotestsum --format pkgname-and-test-fails --format-icons hivis -- -race -shuffle=on -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
else
	go test -race -shuffle=on -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
endif
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "$(COLOR_GREEN)✓ Coverage report: $(COVERAGE_DIR)/coverage.html$(COLOR_RESET)"
	@go tool cover -func=$(COVERAGE_DIR)/coverage.out | tail -1

.PHONY: test-short
test-short: ## Run short tests only
	@echo "$(COLOR_BLUE)▶ Running short tests...$(COLOR_RESET)"
ifdef GOTESTSUM
	gotestsum --format pkgname-and-test-fails --format-icons hivis -- -race -shuffle=on -short ./...
else
	go test -race -shuffle=on -short ./...
endif

.PHONY: bench
bench: ## Run benchmarks
	@echo "$(COLOR_BLUE)▶ Running benchmarks...$(COLOR_RESET)"
	go test -bench=. -benchmem ./...

# ══════════════════════════════════════════════════════════════════════════════
# Code Quality
# ══════════════════════════════════════════════════════════════════════════════

.PHONY: lint
lint: ## Run linter
	@echo "$(COLOR_BLUE)▶ Running linter...$(COLOR_RESET)"
	$(GOLANGCI_LINT) run ./...
	@echo "$(COLOR_GREEN)✓ Linting passed$(COLOR_RESET)"

.PHONY: lint-fix
lint-fix: ## Run linter with auto-fix
	@echo "$(COLOR_BLUE)▶ Running linter with auto-fix...$(COLOR_RESET)"
	$(GOLANGCI_LINT) run --fix ./...
	@echo "$(COLOR_GREEN)✓ Linting complete$(COLOR_RESET)"

.PHONY: fmt
fmt: ## Format code
	@echo "$(COLOR_BLUE)▶ Formatting code...$(COLOR_RESET)"
	go fmt ./...
	$(GOLANGCI_LINT) fmt ./...
	@echo "$(COLOR_GREEN)✓ Formatting complete$(COLOR_RESET)"

.PHONY: vet
vet: ## Run go vet
	@echo "$(COLOR_BLUE)▶ Running go vet...$(COLOR_RESET)"
	go vet ./...
	@echo "$(COLOR_GREEN)✓ Vet passed$(COLOR_RESET)"

.PHONY: tidy
tidy: ## Tidy go.mod
	@echo "$(COLOR_BLUE)▶ Tidying go.mod...$(COLOR_RESET)"
	go mod tidy
	@echo "$(COLOR_GREEN)✓ go.mod tidied$(COLOR_RESET)"

.PHONY: verify
verify: ## Verify dependencies
	@echo "$(COLOR_BLUE)▶ Verifying dependencies...$(COLOR_RESET)"
	go mod verify
	@echo "$(COLOR_GREEN)✓ Dependencies verified$(COLOR_RESET)"

# ══════════════════════════════════════════════════════════════════════════════
# CI Pipeline
# ══════════════════════════════════════════════════════════════════════════════

.PHONY: ci
ci: tidy verify vet lint test build ## Run full CI pipeline
	@echo ""
	@echo "$(COLOR_GREEN)$(COLOR_BOLD)✓ CI pipeline passed!$(COLOR_RESET)"
	@echo ""

.PHONY: ci-fast
ci-fast: vet lint-fix test-short build ## Run fast CI (for local development)
	@echo "$(COLOR_GREEN)✓ Fast CI passed$(COLOR_RESET)"

# ══════════════════════════════════════════════════════════════════════════════
# Release
# ══════════════════════════════════════════════════════════════════════════════

.PHONY: release-check
release-check: ## Check GoReleaser configuration
	@echo "$(COLOR_BLUE)▶ Checking GoReleaser config...$(COLOR_RESET)"
	goreleaser check
	@echo "$(COLOR_GREEN)✓ Config valid$(COLOR_RESET)"

.PHONY: release-snapshot
release-snapshot: ## Build snapshot release (no publish)
	@echo "$(COLOR_BLUE)▶ Building snapshot release...$(COLOR_RESET)"
	GO_VERSION=$(GO_VERSION) goreleaser release --snapshot --clean
	@echo "$(COLOR_GREEN)✓ Snapshot built in dist/$(COLOR_RESET)"

.PHONY: release-dry-run
release-dry-run: ## Dry run release (no publish)
	@echo "$(COLOR_BLUE)▶ Dry run release...$(COLOR_RESET)"
	GO_VERSION=$(GO_VERSION) goreleaser release --skip=publish --clean
	@echo "$(COLOR_GREEN)✓ Dry run complete$(COLOR_RESET)"

.PHONY: release
release: ## Create release (requires GITHUB_TOKEN)
	@echo "$(COLOR_BLUE)▶ Creating release...$(COLOR_RESET)"
	$(GORELEASER) release --clean

# ══════════════════════════════════════════════════════════════════════════════
# Utilities
# ══════════════════════════════════════════════════════════════════════════════

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(COLOR_BLUE)▶ Cleaning...$(COLOR_RESET)"
	rm -rf $(BIN_DIR) $(DIST_DIR) $(COVERAGE_DIR)
	go clean -cache -testcache
	@echo "$(COLOR_GREEN)✓ Cleaned$(COLOR_RESET)"

.PHONY: deps
deps: ## Download dependencies
	@echo "$(COLOR_BLUE)▶ Downloading dependencies...$(COLOR_RESET)"
	go mod download
	@echo "$(COLOR_GREEN)✓ Dependencies downloaded$(COLOR_RESET)"

.PHONY: tools
tools: ## Install development tools
	@echo "$(COLOR_BLUE)▶ Installing tools...$(COLOR_RESET)"
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
	go install github.com/goreleaser/goreleaser/v2@latest
	go install gotest.tools/gotestsum@latest
	go install github.com/air-verse/air@latest
	@echo "$(COLOR_GREEN)✓ Tools installed$(COLOR_RESET)"

.PHONY: tools-ci
tools-ci: ## Install CI tools (minimal)
	@echo "$(COLOR_BLUE)▶ Installing CI tools...$(COLOR_RESET)"
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
	go install gotest.tools/gotestsum@latest
	@echo "$(COLOR_GREEN)✓ CI tools installed$(COLOR_RESET)"

.PHONY: version
version: ## Show version info
	@echo "$(COLOR_MAGENTA)Version:    $(VERSION)$(COLOR_RESET)"
	@echo "$(COLOR_MAGENTA)Commit:     $(COMMIT)$(COLOR_RESET)"
	@echo "$(COLOR_MAGENTA)Build Date: $(BUILD_DATE)$(COLOR_RESET)"
	@echo "$(COLOR_MAGENTA)Go Version: $(GO_VERSION)$(COLOR_RESET)"

.PHONY: info
info: ## Show project info
	@echo ""
	@echo "$(COLOR_BOLD)⚡ Yukti (युक्ति)$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Beautiful TUI for managing Google Apps Script projects.$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BLUE)Repository:$(COLOR_RESET) https://github.com/robinsharma/yukti"
	@echo "$(COLOR_BLUE)Go Version:$(COLOR_RESET) $(GO_VERSION)"
	@echo "$(COLOR_BLUE)Build:$(COLOR_RESET)      $(VERSION) ($(COMMIT))"
	@echo ""

# ══════════════════════════════════════════════════════════════════════════════
# Git Hooks (Lefthook)
# ══════════════════════════════════════════════════════════════════════════════

LEFTHOOK := $(shell command -v lefthook 2> /dev/null)

.PHONY: lefthook-install
lefthook-install: ## Install lefthook if not present
ifndef LEFTHOOK
	@echo "$(COLOR_BLUE)▶ Installing lefthook...$(COLOR_RESET)"
	@go install github.com/evilmartians/lefthook@latest
endif
	@echo "$(COLOR_GREEN)✓ Lefthook is installed$(COLOR_RESET)"

.PHONY: hooks
hooks: lefthook-install ## Setup git hooks with lefthook
	@echo "$(COLOR_BLUE)▶ Installing git hooks via lefthook...$(COLOR_RESET)"
	@lefthook install
	@echo "$(COLOR_GREEN)✓ Git hooks installed (pre-push runs 'make ci')$(COLOR_RESET)"

.PHONY: hooks-uninstall
hooks-uninstall: ## Remove lefthook git hooks
	@echo "$(COLOR_BLUE)▶ Removing git hooks...$(COLOR_RESET)"
	@lefthook uninstall
	@echo "$(COLOR_GREEN)✓ Git hooks removed$(COLOR_RESET)"

.PHONY: hooks-run
hooks-run: ## Run pre-push hook manually
	@echo "$(COLOR_BLUE)▶ Running pre-push hook...$(COLOR_RESET)"
	@lefthook run pre-push
