SHELL := /bin/bash
GO ?= go
PKG := ./...
BINARY := shotgun-cli
BUILD_DIR := build
COVERAGE_FILE := coverage.out
GO_FILES := $(shell find . -name '*.go' -not -path './vendor/*')

OS_ARCHES := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: help build build-all test test-race test-bench test-e2e lint fmt vet clean install deps generate coverage release

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
	golangci-lint run ./...

fmt: ## Format Go source files
	$(GO) fmt ./...

vet: ## Run go vet static analysis
	$(GO) vet $(PKG)

clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR) $(COVERAGE_FILE)

install: ## Install the binary into GOPATH/bin
	$(GO) install .

deps: ## Download and verify module dependencies
	$(GO) mod download
	$(GO) mod verify

generate: ## Run go generate
	$(GO) generate $(PKG)

coverage: ## Generate coverage profile and report
	$(GO) test -coverprofile=$(COVERAGE_FILE) $(PKG)
	$(GO) tool cover -func=$(COVERAGE_FILE)

release: ## Build release artifacts with Goreleaser
	goreleaser release --clean
