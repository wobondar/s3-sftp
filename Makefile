# Project variables
BINARY_NAME=s3-sftp
BUILD_DIR=build
GO_FILES=$(wildcard cmd/s3-sftp/*.go)
GO_CMD_DIR=cmd/s3-sftp/*.go
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet

# Colors for terminal output
COLOR_RESET=\033[0m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m
COLOR_RED=\033[31m

.DEFAULT_GOAL := help

.PHONY: all check-deps build build-all clean vet release-check fmt mod-tidy run help

all: check-deps clean build ## Check dependencies, clean, and build

check-deps: ## Check for required dependencies
	@echo "$(COLOR_BLUE)Checking dependencies...$(COLOR_RESET)"
	@which $(GOCMD) > /dev/null || (echo "$(COLOR_RED)Error: Go is not installed$(COLOR_RESET)" && exit 1)
	@$(GOCMD) version

build: ## Build the application
	@echo "$(COLOR_BLUE)Building $(BINARY_NAME) version $(VERSION)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(GO_CMD_DIR)
	@echo "$(COLOR_GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)"

build-all: check-deps clean ## Build for multiple platforms (Linux, macOS, Windows)
	@echo "$(COLOR_BLUE)Building for multiple platforms...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@./build.sh $(ARGS) || (echo "\n$(COLOR_RED)Use: make build-all ARGS=v0.0.1$(COLOR_RESET)\n$(COLOR_YELLOW)To set the version$(COLOR_RESET)\n" && exit 1)
	@echo "$(COLOR_GREEN)Multi-platform build complete!$(COLOR_RESET)"

clean: ## Remove build artifacts
	@echo "$(COLOR_BLUE)Cleaning...$(COLOR_RESET)"
	@rm -rf $(BUILD_DIR)
	@echo "$(COLOR_GREEN)Clean complete!$(COLOR_RESET)"

vet: ## Run go vet
	@echo "$(COLOR_BLUE)Running go vet...$(COLOR_RESET)"
	$(GOVET) ./...
	@echo "$(COLOR_GREEN)Vet complete!$(COLOR_RESET)"

release-check: ## Check goreleaser release without publishing
	@echo "$(COLOR_BLUE)Checking for release...$(COLOR_RESET)"
	@goreleaser check
	@echo "$(COLOR_GREEN)Release check complete!$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BLUE)Checking for release build without publishing...$(COLOR_RESET)"
	@goreleaser release --snapshot --skip=publish --clean
	@echo "$(COLOR_GREEN)Release build check complete!$(COLOR_RESET)"

fmt: ## Format code
	@echo "$(COLOR_BLUE)Formatting code...$(COLOR_RESET)"
	@gofmt -s -w $(GO_FILES)
	@echo "$(COLOR_GREEN)Format complete!$(COLOR_RESET)"

mod-tidy: ## Tidy Go modules
	@echo "$(COLOR_BLUE)Tidying modules...$(COLOR_RESET)"
	$(GOMOD) tidy
	@echo "$(COLOR_GREEN)Tidy complete!$(COLOR_RESET)"

run: build ## Build and run the application
	@echo "$(COLOR_BLUE)Running $(BINARY_NAME)...$(COLOR_RESET)"
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

help:	## Show this help message
	@echo "$(COLOR_BLUE)Available targets:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-18s$(COLOR_RESET) - %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_BLUE)Example usage:$(COLOR_RESET)"
	@echo "  make run ARGS=\"-csv input.csv -c config.json\""
	@echo "  make build-all ARGS=\"v0.0.1\""
