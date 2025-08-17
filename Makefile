# go-wc Makefile
# Production-ready build system with optimization and standard processes

# Project configuration
PROJECT_NAME := go-wc
BINARY_NAME := go_wc
MAIN_PATH := ./cmd/go_wc
PKG_LIST := $(shell go list ./... | grep -v /vendor/)

# Build configuration
BUILD_DIR := bin
DIST_DIR := dist
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION := $(shell go version | cut -d' ' -f3)

# Build flags for optimization
LDFLAGS := -s -w \
	-X 'main.version=$(VERSION)' \
	-X 'main.commit=$(COMMIT)' \
	-X 'main.buildTime=$(BUILD_TIME)' \
	-X 'main.goVersion=$(GO_VERSION)'

# CGO and build tags
CGO_ENABLED := 0
BUILD_TAGS := netgo osusergo static_build

# Cross-compilation targets
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64
ARCH_TARGETS := $(addprefix build-,$(subst /,-,$(PLATFORMS)))

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
PURPLE := \033[0;35m
CYAN := \033[0;36m
NC := \033[0m # No Color

# Default target
.DEFAULT_GOAL := help

## help: Show this help message
.PHONY: help
help:
	@echo "$(CYAN)$(PROJECT_NAME) Build System$(NC)"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
	@echo ""
	@echo "$(YELLOW)Cross-compilation targets:$(NC)"
	@echo "  $(PLATFORMS)" | tr ' ' '\n' | sed 's/^/  build-/' | tr '/' '-'

## clean: Remove build artifacts and cache
.PHONY: clean
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@go clean -cache -testcache -modcache
	@echo "$(GREEN)✓ Clean completed$(NC)"

## deps: Download and verify dependencies
.PHONY: deps
deps:
	@echo "$(YELLOW)Downloading dependencies...$(NC)"
	@go mod download
	@go mod verify
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

## fmt: Format Go code
.PHONY: fmt
fmt:
	@echo "$(YELLOW)Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

## vet: Run go vet
.PHONY: vet
vet:
	@echo "$(YELLOW)Running go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✓ Vet completed$(NC)"

## lint: Run golangci-lint (install if needed)
.PHONY: lint
lint:
	@echo "$(YELLOW)Running linter...$(NC)"
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "$(BLUE)Installing golangci-lint...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@golangci-lint run --timeout=5m
	@echo "$(GREEN)✓ Linting completed$(NC)"

## test: Run tests
.PHONY: test
test:
	@echo "$(YELLOW)Running tests...$(NC)"
	@go test -v -race -timeout=30s ./...
	@echo "$(GREEN)✓ Tests completed$(NC)"

## test-coverage: Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "$(YELLOW)Running tests with coverage...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go test -v -race -timeout=30s -coverprofile=$(BUILD_DIR)/coverage.out ./...
	@go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@go tool cover -func=$(BUILD_DIR)/coverage.out | tail -1
	@echo "$(GREEN)✓ Coverage report generated: $(BUILD_DIR)/coverage.html$(NC)"

## bench: Run benchmarks
.PHONY: bench
bench:
	@echo "$(YELLOW)Running benchmarks...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go test -bench=. -benchmem -timeout=10m ./... | tee $(BUILD_DIR)/bench.out
	@echo "$(GREEN)✓ Benchmarks completed$(NC)"

## security: Run security checks
.PHONY: security
security:
	@echo "$(YELLOW)Running security checks...$(NC)"
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "$(BLUE)Installing gosec...$(NC)"; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	@gosec -quiet ./...
	@echo "$(GREEN)✓ Security checks completed$(NC)"

## check: Run all checks (fmt, vet, lint, test, security)
.PHONY: check
check: fmt vet lint test security
	@echo "$(GREEN)✓ All checks passed$(NC)"

## build: Build optimized binary for current platform
.PHONY: build
build:
	@echo "$(YELLOW)Building $(BINARY_NAME) for current platform...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) go build \
		-tags "$(BUILD_TAGS)" \
		-ldflags "$(LDFLAGS)" \
		-trimpath \
		-o $(BUILD_DIR)/$(BINARY_NAME) \
		$(MAIN_PATH)
	@echo "$(GREEN)✓ Binary built: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)

## build-debug: Build binary with debug symbols
.PHONY: build-debug
build-debug:
	@echo "$(YELLOW)Building debug binary...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) go build \
		-tags "$(BUILD_TAGS)" \
		-ldflags "-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)'" \
		-gcflags "all=-N -l" \
		-o $(BUILD_DIR)/$(BINARY_NAME)-debug \
		$(MAIN_PATH)
	@echo "$(GREEN)✓ Debug binary built: $(BUILD_DIR)/$(BINARY_NAME)-debug$(NC)"

## install: Install binary to GOPATH/bin
.PHONY: install
install: build
	@echo "$(YELLOW)Installing $(BINARY_NAME)...$(NC)"
	@go install -tags "$(BUILD_TAGS)" -ldflags "$(LDFLAGS)" $(MAIN_PATH)
	@echo "$(GREEN)✓ Binary installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)$(NC)"

## release: Build optimized binaries for all platforms
.PHONY: release
release: clean $(ARCH_TARGETS)
	@echo "$(GREEN)✓ Release build completed$(NC)"
	@echo "$(CYAN)Built binaries:$(NC)"
	@find $(DIST_DIR) -name "$(BINARY_NAME)*" -exec ls -lh {} \;

# Cross-compilation targets
$(ARCH_TARGETS):
	$(eval PLATFORM := $(subst build-,,$@))
	$(eval GOOS := $(word 1,$(subst -, ,$(PLATFORM))))
	$(eval GOARCH := $(word 2,$(subst -, ,$(PLATFORM))))
	$(eval OUTPUT := $(DIST_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH))
	$(eval OUTPUT := $(if $(filter windows,$(GOOS)),$(OUTPUT).exe,$(OUTPUT)))
	@echo "$(YELLOW)Building for $(GOOS)/$(GOARCH)...$(NC)"
	@mkdir -p $(DIST_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-tags "$(BUILD_TAGS)" \
		-ldflags "$(LDFLAGS)" \
		-trimpath \
		-o $(OUTPUT) \
		$(MAIN_PATH)
	@echo "$(GREEN)✓ Built: $(OUTPUT)$(NC)"

## docker-build: Build Docker image
.PHONY: docker-build
docker-build:
	@echo "$(YELLOW)Building Docker image...$(NC)"
	@docker build -t $(PROJECT_NAME):$(VERSION) -t $(PROJECT_NAME):latest .
	@echo "$(GREEN)✓ Docker image built$(NC)"

## docker-run: Run Docker container
.PHONY: docker-run
docker-run:
	@echo "$(YELLOW)Running Docker container...$(NC)"
	@docker run --rm -it $(PROJECT_NAME):latest --help

## size: Show binary size analysis
.PHONY: size
size: build
	@echo "$(CYAN)Binary size analysis:$(NC)"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)
	@echo ""
	@echo "$(CYAN)Stripped vs unstripped comparison:$(NC)"
	@if command -v strip >/dev/null 2>&1; then \
		cp $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/$(BINARY_NAME)-stripped; \
		strip $(BUILD_DIR)/$(BINARY_NAME)-stripped 2>/dev/null || true; \
		ls -lh $(BUILD_DIR)/$(BINARY_NAME)*; \
	fi

## profile: Build with profiling enabled
.PHONY: profile
profile:
	@echo "$(YELLOW)Building with profiling...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=1 go build \
		-tags "$(BUILD_TAGS)" \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY_NAME)-profile \
		$(MAIN_PATH)
	@echo "$(GREEN)✓ Profile binary built: $(BUILD_DIR)/$(BINARY_NAME)-profile$(NC)"

## mod-update: Update all dependencies to latest versions
.PHONY: mod-update
mod-update:
	@echo "$(YELLOW)Updating dependencies...$(NC)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

## mod-graph: Show dependency graph
.PHONY: mod-graph
mod-graph:
	@echo "$(CYAN)Dependency graph:$(NC)"
	@go mod graph

## info: Show build information
.PHONY: info
info:
	@echo "$(CYAN)Build Information:$(NC)"
	@echo "  Project: $(PROJECT_NAME)"
	@echo "  Binary:  $(BINARY_NAME)"
	@echo "  Version: $(VERSION)"
	@echo "  Commit:  $(COMMIT)"
	@echo "  Go:      $(GO_VERSION)"
	@echo "  CGO:     $(CGO_ENABLED)"
	@echo "  Tags:    $(BUILD_TAGS)"
	@echo "  LDFLAGS: $(LDFLAGS)"

## ci: Run CI pipeline (check + build + test-coverage)
.PHONY: ci
ci: deps check build test-coverage
	@echo "$(GREEN)✓ CI pipeline completed successfully$(NC)"

# Development workflow targets
## dev: Quick development build and test
.PHONY: dev
dev: fmt vet test build
	@echo "$(GREEN)✓ Development build completed$(NC)"

## watch: Watch for changes and rebuild (requires entr)
.PHONY: watch
watch:
	@if ! command -v entr >/dev/null 2>&1; then \
		echo "$(RED)Error: entr is required for watch mode$(NC)"; \
		echo "Install with: apt-get install entr (Ubuntu) or brew install entr (macOS)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Watching for changes... (Ctrl+C to stop)$(NC)"
	@find . -name "*.go" | entr -r make dev

# Ensure build directories exist
$(BUILD_DIR):
	@mkdir -p $@

$(DIST_DIR):
	@mkdir -p $@

# Prevent make from treating files as targets
.PHONY: $(ARCH_TARGETS)