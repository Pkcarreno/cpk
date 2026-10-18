.PHONY: help test fmt vet check build install
PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

NAME := cpk
PREFIX := github.com/pkcarreno/
PROJECT := $(PREFIX)$(NAME)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --verify HEAD 2>/dev/null || echo "none")
DATE := $(shell git log -1 --format=%cI 2>/dev/null || date -u +%Y-%m-%dT%H:%M:%SZ)

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-40s\033[0m %s\n", $$1, $$2}'

test: ## Run unit tests
	go test ./...

fmt: ## Format source files
	gofmt -s -l -w .

vet: ## Inspect code with static analyzer
	go vet ./...

check: fmt vet test ## Run fmt, vet, and test

build: ## Compile binary with build metadata
	@echo "Building with the following parameters:"
	@echo "VERSION = $(VERSION)"
	@echo "COMMIT  = $(COMMIT)"
	@echo "DATE    = $(DATE)"
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -ldflags "-s -w -X \"${PROJECT}/internal/version.version=${VERSION}\" -X \"${PROJECT}/internal/version.commit=${COMMIT}\" -X \"${PROJECT}/internal/version.date=${DATE}\"" -o $(PWD)/build/$(NAME)

install: ## Install binary to GOBIN with build metadata
	@echo "Installing with the following parameters:"
	@echo "VERSION = $(VERSION)"
	@echo "COMMIT  = $(COMMIT)"
	@echo "DATE    = $(DATE)"
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go install -trimpath -ldflags "-s -w -X \"${PROJECT}/internal/version.version=${VERSION}\" -X \"${PROJECT}/internal/version.commit=${COMMIT}\" -X \"${PROJECT}/internal/version.date=${DATE}\""
