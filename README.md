# voxera-kit

Pluggable infrastructure SDK for the Voxera platform.

## Overview

voxera-kit is a monorepo providing reusable frontend packages and backend modules that power the Voxera platform. It is designed with a modular architecture so teams can adopt individual packages independently or use them together as a cohesive toolkit.

## Architecture

```
voxera-kit/
├── frontend/              # TypeScript packages (pnpm workspace + Turborepo)
│   └── packages/
│       └── config/        # Shared TS, ESLint & Prettier configs
├── backend/               # Go modules (Go workspace)
├── .github/workflows/     # CI & Release pipelines
└── .vscode/               # Editor settings & recommended extensions
```

## Frontend Packages

| Package | Description |
| --- | --- |
| `@voxera-kit/config` | Shared TypeScript, ESLint, and Prettier configurations |

## Backend Modules

| Module | Description |
| --- | --- |
| *(coming soon)* | Go service modules will be added here |

## Getting Started

### Prerequisites

- **Node.js** >= 20
- **pnpm** >= 9
- **Go** >= 1.22
- **Turborepo** (installed as a dev dependency)

### Install

```bash
# Frontend
cd frontend
pnpm install

# Backend
cd backend
go work sync
```

## Development

### Frontend Commands

```bash
cd frontend

pnpm build        # Build all packages
pnpm lint         # Lint all packages
pnpm test         # Run all tests
pnpm typecheck    # Type-check all packages
pnpm clean        # Clean build artifacts
```

### Backend Commands

```bash
cd backend

go build ./...    # Build all modules
go test ./...     # Run all tests
go vet ./...      # Vet all modules
```

## License

[MIT](./LICENSE)
