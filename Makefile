SHELL := /bin/bash
GO ?= go
PKG := ./...
BINARY := shotgun-cli
BUILD_DIR := build
COVERAGE_FILE := coverage.out
GO_FILES := $(shell find . -name '*.go' -not -path './vendor/*')

# Installation paths
PREFIX ?= /usr/local
INSTALL_DIR := $(PREFIX)/bin

OS_ARCHES := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: help build build-all test test-race test-bench test-e2e lint fmt vet clean install install-local install-system uninstall deps generate coverage release version-bump version-patch version-minor version-major release-tag release-push release-snapshot

help:
	@echo "Usage: make <target>"
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?##' Makefile | sed 's/:.*##/: /'

build: ## Build the binary for the current platform
	$(GO) build -o $(BUILD_DIR)/$(BINARY) .

build-all: clean ## Cross-compile for common platforms
	@mkdir -p $(BUILD_DIR)
	@for target in $(OS_ARCHES); do \
		IFS=/ read -r os arch <<< $$target; \
		echo "Building $$os/$$arch"; \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		GOOS=$$os GOARCH=$$arch $(GO) build -o $(BUILD_DIR)/$(BINARY)-$$os-$$arch$$ext .; \
	done

test: ## Run unit tests with default settings
	$(GO) test $(PKG)

test-race: ## Run tests with the race detector
	$(GO) test -race $(PKG)

test-bench: ## Run package benchmarks
	$(GO) test -bench=. -run=^$$ $(PKG)

test-e2e: ## Execute end-to-end CLI tests
	$(GO) test ./test/e2e -v

lint: ## Run golangci-lint
	if [ -f .golangci-local.yml ]; then \
		golangci-lint run --config .golangci-local.yml ./...; \
	else \
		golangci-lint run ./...; \
	fi

fmt: ## Format Go source files
	$(GO) fmt ./...

vet: ## Run go vet static analysis
	$(GO) vet $(PKG)

clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR) $(COVERAGE_FILE)

install: install-local ## Install the binary (default: local GOPATH/bin)

install-local: ## Install the binary into GOPATH/bin
	@echo "Installing $(BINARY) to GOPATH/bin..."
	$(GO) install .
	@echo "✅ $(BINARY) installed successfully to GOPATH/bin"
	@echo "Make sure $(shell go env GOPATH)/bin is in your PATH"

install-system: build ## Install the binary to system-wide location (requires sudo)
	@echo "Installing $(BINARY) to $(INSTALL_DIR)..."
	@if [ ! -f "$(BUILD_DIR)/$(BINARY)" ]; then \
		echo "Binary not found. Building first..."; \
		$(MAKE) build; \
	fi
	sudo install -m 755 $(BUILD_DIR)/$(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "✅ $(BINARY) installed successfully to $(INSTALL_DIR)"
	@echo "You can now use '$(BINARY)' from anywhere"

uninstall: ## Remove installed binary from system
	@echo "Removing $(BINARY) from system..."
	@if [ -f "$(INSTALL_DIR)/$(BINARY)" ]; then \
		sudo rm -f $(INSTALL_DIR)/$(BINARY); \
		echo "✅ $(BINARY) removed from $(INSTALL_DIR)"; \
	else \
		echo "$(BINARY) not found in $(INSTALL_DIR)"; \
	fi
	@echo "Note: To remove from GOPATH/bin, run: rm $(shell go env GOPATH)/bin/$(BINARY)"

deps: ## Download and verify module dependencies
	$(GO) mod download
	$(GO) mod verify

generate: ## Run go generate
	$(GO) generate $(PKG)

coverage: ## Generate coverage profile and report
	$(GO) test -coverprofile=$(COVERAGE_FILE) $(PKG)
	$(GO) tool cover -func=$(COVERAGE_FILE)

# Version management
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
VERSION_RAW := $(shell echo "$(VERSION)" | sed 's/^v//')
VERSION_NEXT := $(shell bash -c 'version="$(VERSION_RAW)"; IFS=. read -r major minor patch <<< "$$version"; echo "$$major.$$((minor+1)).0"')

version-bump: ## Display current and next version
	@echo "Current version: $(VERSION)"
	@echo "Next minor version: v$(VERSION_NEXT)"
	@echo ""
	@echo "Use one of:"
	@echo "  make version-patch  # Bump patch version (0.0.X -> 0.0.Y)"
	@echo "  make version-minor  # Bump minor version (0.X.0 -> 0.Y.0)"
	@echo "  make version-major  # Bump major version (X.0.0 -> Y.0.0)"

version-patch: ## Bump patch version (e.g., v1.2.3 -> v1.2.4)
	@bash -c 'version="$(VERSION_RAW)"; IFS=. read -r major minor patch <<< "$$version"; new_version="$$major.$$minor.$$((patch+1))"; \
	echo "Bumping version from $(VERSION) to v$$new_version"; \
	echo "$$new_version" > .version.tmp && mv .version.tmp VERSION'

version-minor: ## Bump minor version (e.g., v1.2.3 -> v1.3.0)
	@bash -c 'version="$(VERSION_RAW)"; IFS=. read -r major minor patch <<< "$$version"; new_version="$$major.$$((minor+1)).0"; \
	echo "Bumping version from $(VERSION) to v$$new_version"; \
	echo "$$new_version" > .version.tmp && mv .version.tmp VERSION'

version-major: ## Bump major version (e.g., v1.2.3 -> v2.0.0)
	@bash -c 'version="$(VERSION_RAW)"; IFS=. read -r major minor patch <<< "$$version"; new_version="$$((major+1)).0.0"; \
	echo "Bumping version from $(VERSION) to v$$new_version"; \
	echo "$$new_version" > .version.tmp && mv .version.tmp VERSION'

version-set: ## Set a specific version (use VERSION=1.2.3)
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION parameter is required. Usage: make version-set VERSION=1.2.3"; \
		exit 1; \
	fi
	@echo "Setting version to v$(VERSION)"
	@echo "$(VERSION)" > .version.tmp && mv .version.tmp VERSION

release-tag: ## Create and push a new git tag (use VERSION=1.2.3)
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION parameter is required. Usage: make release-tag VERSION=1.2.3"; \
		exit 1; \
	fi
	@echo "Checking working tree status..."
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Error: Working tree is not clean. Please commit or stash changes first."; \
		git status --short; \
		exit 1; \
	fi
	@echo "Current version: $(VERSION)"
	@echo "Creating tag v$(VERSION)..."
	@git tag -a "v$(VERSION)" -m "Release v$(VERSION)"
	@echo "Tag v$(VERSION) created successfully"
	@echo ""
	@echo "To push the tag and trigger the release, run:"
	@echo "  make release-push VERSION=$(VERSION)"

release-push: ## Push tag to remote and trigger GitHub release
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION parameter is required. Usage: make release-push VERSION=1.2.3"; \
		exit 1; \
	fi
	@echo "Pushing tag v$(VERSION) to origin..."
	@git push origin "v$(VERSION)"
	@echo ""
	@echo "Tag pushed! Release workflow should start at:"
	@echo "  https://github.com/quantmind-br/shotgun-cli/actions"

release-snapshot: ## Build release artifacts locally without creating a tag
	goreleaser release --snapshot --clean

release-test: ## Test release configuration without publishing
	goreleaser release --snapshot --clean --skip-publish

release: ## Create and push new release (use VERSION=1.2.3)
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION parameter is required. Usage: make release VERSION=1.2.3"; \
		exit 1; \
	fi
	@echo "🚀 Starting release process for v$(VERSION)..."
	@echo ""
	@$(MAKE) release-tag VERSION=$(VERSION)
	@echo ""
	@$(MAKE) release-push VERSION=$(VERSION)
	@echo ""
	@echo "✅ Release v$(VERSION) initiated!"
	@echo "   Watch the release at: https://github.com/quantmind-br/shotgun-cli/releases"
