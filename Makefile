.PHONY: help test fmt vet check

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-40s\033[0m %s\n", $$1, $$2}'

test: ## Run unit tests
	go test ./...

fmt: ## Format source files
	gofmt -s -w .

vet: ## Inspect code with static analyzer
	go vet ./...

check: fmt vet test ## Run fmt, vet, and test
