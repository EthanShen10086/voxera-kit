.PHONY: help lint fmt test build vet typecheck clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

lint: lint-go lint-ts ## Run all linters

lint-go: ## Run golangci-lint on backend
	cd backend && golangci-lint run ./...

lint-ts: ## Run ESLint on frontend
	cd frontend && pnpm turbo lint

fmt: fmt-go fmt-ts ## Format all code

fmt-go: ## Format Go code
	cd backend && gofmt -w .

fmt-ts: ## Format TypeScript code
	cd frontend && pnpm prettier --write "packages/*/src/**/*.ts"

test: test-go test-ts ## Run all tests

test-go: ## Run Go tests
	cd backend && for dir in $$(grep -oP '\./\K[a-z]+' go.work); do cd "$$dir" && go test ./... -race && cd ..; done

test-ts: ## Run TypeScript tests
	cd frontend && pnpm turbo test

build: build-go build-ts ## Build everything

build-go: ## Build all Go modules
	cd backend && for dir in $$(grep -oP '\./\K[a-z]+' go.work); do cd "$$dir" && go build ./... && cd ..; done

build-ts: ## Build all TypeScript packages
	cd frontend && pnpm turbo build

vet: ## Run go vet on all modules
	cd backend && for dir in $$(grep -oP '\./\K[a-z]+' go.work); do cd "$$dir" && go vet ./... && cd ..; done

typecheck: ## Run TypeScript type checking
	cd frontend && pnpm turbo typecheck

clean: ## Clean build artifacts
	cd frontend && pnpm turbo clean
	find backend -name "*.test" -delete
	find backend -name "coverage.out" -delete
