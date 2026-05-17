# Development Guide

## Local Setup

### 1. Clone & Install

```bash
git clone https://github.com/your-org/voxera-kit.git
cd voxera-kit

# Root dependencies (husky, commitlint)
npm install

# Frontend dependencies
cd frontend && pnpm install && cd ..

# Backend workspace sync
cd backend && go work sync && cd ..
```

### 2. Verify Installation

```bash
make lint    # Should pass with no errors
make test    # Should pass all tests
make build   # Should compile everything
```

## Running Frontend Dev

```bash
cd frontend

# Build all packages
pnpm turbo build

# Run linter
pnpm turbo lint

# Run type checking
pnpm turbo typecheck

# Run tests
pnpm turbo test

# Format code
pnpm format
```

To work on a specific package:

```bash
cd frontend/packages/api-client
pnpm build
pnpm test
```

## Running Go Modules

```bash
cd backend

# Build a specific module
cd database && go build ./...

# Test a specific module
cd cache && go test ./... -race -v

# Run all tests via Makefile
cd .. && make test-go

# Lint all Go code
make lint-go

# Format all Go code
make fmt-go
```

## Testing

### Go Tests

```bash
# All modules
make test-go

# Single module with verbose output
cd backend/database && go test ./... -race -v -count=1

# With coverage
cd backend/database && go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### TypeScript Tests

```bash
# All packages
make test-ts

# Single package
cd frontend/packages/api-client && pnpm vitest run

# Watch mode
cd frontend/packages/api-client && pnpm vitest
```

## Debugging Tips

- **Go:** Use `dlv debug` or VS Code's Go debugger. The `.vscode/settings.json` is pre-configured.
- **TypeScript:** Use VS Code's built-in debugger. Set breakpoints directly in `.ts` files.
- **Lint issues:** Run `make lint` to see all issues. Use `make fmt` to auto-fix formatting.
- **Module not found:** Run `go work sync` in `backend/` or `pnpm install` in `frontend/`.
- **Turbo cache stale:** Run `pnpm turbo clean` then rebuild.
