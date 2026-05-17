# Contributing to Voxera Kit

Thank you for contributing to Voxera Kit! This guide will help you get started.

## Prerequisites

| Tool    | Version  | Install                                |
| ------- | -------- | -------------------------------------- |
| Node.js | 22+      | https://nodejs.org                     |
| pnpm    | 9+       | `npm i -g pnpm`                       |
| Go      | 1.22+    | https://go.dev/dl                      |
| Make    | any      | Pre-installed on macOS / Linux         |
| Docker  | 24+      | https://docs.docker.com/get-docker     |

## Getting Started

```bash
git clone https://github.com/your-org/voxera-kit.git
cd voxera-kit

# Install root dev dependencies (husky, commitlint)
npm install

# Frontend
cd frontend
pnpm install
pnpm build
cd ..

# Backend
cd backend
go work sync
cd ..

# Verify everything works
make lint
make test
```

## Project Structure

```
voxera-kit/
├── backend/           # Go modules (database, cache, auth, mq, ...)
│   ├── go.work        # Go workspace definition
│   ├── database/      # Database port & adapters
│   ├── cache/         # Cache port & adapters
│   └── ...            # 25 modules total
├── frontend/          # TypeScript packages (pnpm workspace + Turborepo)
│   ├── packages/
│   │   ├── config/    # Shared ESLint, Prettier, TSConfig
│   │   ├── api-client/
│   │   └── ...        # 15 packages total
│   └── package.json
├── .github/workflows/ # CI/CD pipelines
├── Makefile           # Build & lint shortcuts
└── docs/              # Architecture & development guides
```

## Commit Conventions

We use [Conventional Commits](https://www.conventionalcommits.org/) enforced by commitlint.

**Format:** `type(scope): description`

Allowed types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`

Examples:
- `feat(database): add PostgreSQL connection pooling adapter`
- `fix(cache): handle Redis cluster failover`
- `docs: update contributing guide`

## Branch Strategy

| Branch        | Purpose                          |
| ------------- | -------------------------------- |
| `main`        | Stable release branch            |
| `develop`     | Integration branch               |
| `feature/*`   | New features                     |
| `fix/*`       | Bug fixes                        |
| `chore/*`     | Maintenance & tooling            |

## Pull Request Process

1. Create a feature branch from `develop`: `git checkout -b feature/my-feature develop`
2. Make your changes with conventional commits
3. Ensure CI passes: `make lint && make test`
4. Push and create a pull request targeting `develop`
5. Fill out the PR template completely
6. Request review from at least one maintainer
7. Address review feedback
8. Squash-merge once approved

## Code Style

### Go

- Format with `gofmt` (enforced by CI)
- Lint with `golangci-lint` using the shared config at `backend/.golangci.yml`
- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Run: `make lint-go` and `make fmt-go`

### TypeScript

- Lint with ESLint (shared config at `frontend/packages/config/eslint.config.ts`)
- Format with Prettier (config at `frontend/.prettierrc.json`)
- Run: `make lint-ts` and `make fmt-ts`

## Adding a New Module

### Go Module

1. Create a directory under `backend/`: `mkdir backend/mymodule`
2. Initialize: `cd backend/mymodule && go mod init github.com/your-org/voxera-kit/mymodule`
3. Create `port.go` with the interface definition
4. Create adapter implementations (e.g., `adapter_redis.go`)
5. Add factory function in `factory.go`
6. Write tests: `*_test.go`
7. Add to `backend/go.work`: `use ./mymodule`

### TypeScript Package

1. Create a directory under `frontend/packages/`: `mkdir frontend/packages/mypackage`
2. Create `package.json` with name `@voxera-kit/mypackage`
3. Create `tsconfig.json` extending `@voxera-kit/config/tsconfig.base.json`
4. Create `src/index.ts` as entry point
5. Implement and export your API
6. Write tests alongside source files
7. Package is auto-discovered by the pnpm workspace

## Questions?

Open an issue or reach out to the maintainers. We're happy to help!
