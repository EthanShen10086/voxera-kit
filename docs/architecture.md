# Voxera Kit Architecture

## Overview

Voxera Kit is a **pluggable SDK** that provides production-ready building blocks for backend services and frontend applications. It follows a modular architecture where each capability (database, cache, auth, etc.) is an independent module with well-defined interfaces.

## Design Principles

- **Ports & Adapters (Hexagonal Architecture):** Each module defines a port (Go interface / TS type) and one or more adapters (concrete implementations).
- **Zero vendor lock-in:** Applications depend on ports, not adapters. Swapping Redis for Memcached means changing one line of config.
- **Independent modules:** Each Go module has its own `go.mod`; each TS package has its own `package.json`. No circular dependencies.
- **Convention over configuration:** Sensible defaults with opt-in customization via functional options or config structs.

## Module Categories

```
┌─────────────────────────────────────────────────┐
│                  Voxera Kit                      │
├──────────────┬──────────────┬───────────────────┤
│  Data Layer  │  Infra Layer │  App Layer        │
├──────────────┼──────────────┼───────────────────┤
│  database    │  auth        │  framework        │
│  cache       │  security    │  config           │
│  storage     │  crypto      │  errors           │
│  mq          │  observability│ share            │
│  dataparser  │  ratelimiter │  shorturl         │
│  dataprovider│  circuitbreaker│ scheduler       │
│              │  concurrency │  registry         │
│              │  compression │  messaging        │
│              │  asr         │                   │
│              │  translation │                   │
│              │  payment     │                   │
└──────────────┴──────────────┴───────────────────┘
```

## Backend Module Structure

Each Go module follows a consistent layout:

```
backend/<module>/
├── go.mod              # Module definition
├── go.sum              # Dependency checksums
├── port.go             # Interface definitions (the "port")
├── adapter_*.go        # Concrete implementations (the "adapters")
├── factory.go          # Factory functions / constructors
├── options.go          # Functional options (if needed)
├── *_test.go           # Tests
└── README.md           # Module-specific docs (optional)
```

**Key pattern:** `port.go` defines the interface that application code depends on. Adapters implement that interface for specific technologies.

## Frontend Package Structure

Each TypeScript package lives under `frontend/packages/`:

```
frontend/packages/<package>/
├── package.json        # Package metadata & dependencies
├── tsconfig.json       # Extends shared base config
├── src/
│   ├── index.ts        # Public API entry point
│   ├── types.ts        # Type definitions
│   └── *.ts            # Implementation files
└── __tests__/          # Test files (optional, or co-located)
```

Packages are managed via pnpm workspaces and built/tested with Turborepo.

## How Apps Consume the Kit

### Go Applications

```go
import (
    "github.com/your-org/voxera-kit/database"
    "github.com/your-org/voxera-kit/cache"
)

func main() {
    db := database.NewPostgres(cfg)    // Pick the adapter you need
    c := cache.NewRedis(redisCfg)      // Swap to Memcached anytime
    app := NewApp(db, c)               // Inject ports, not adapters
}
```

### TypeScript Applications

```typescript
import { createApiClient } from '@voxera-kit/api-client';
import { createCache } from '@voxera-kit/cache';

const api = createApiClient({ baseURL: '/api' });
const cache = createCache({ strategy: 'lru' });
```

## Dependency Flow

```
Application → Kit Port (interface) → Kit Adapter (implementation) → External Service
```

Applications never import adapters directly in production code. Adapter selection happens at the composition root (main function / DI container).
