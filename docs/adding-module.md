# Adding a New Module

## Adding a Go Module

### Step 1: Create the Module Directory

```bash
mkdir backend/mymodule
cd backend/mymodule
go mod init github.com/your-org/voxera-kit/mymodule
```

### Step 2: Define the Port

Create `port.go` with your interface:

```go
package mymodule

import "context"

type MyService interface {
    DoSomething(ctx context.Context, input string) (string, error)
    Close() error
}
```

### Step 3: Implement an Adapter

Create `adapter_default.go`:

```go
package mymodule

type defaultAdapter struct {
    // fields
}

func (a *defaultAdapter) DoSomething(ctx context.Context, input string) (string, error) {
    // implementation
}

func (a *defaultAdapter) Close() error {
    // cleanup
}
```

### Step 4: Add a Factory

Create `factory.go`:

```go
package mymodule

func NewDefault(opts ...Option) MyService {
    a := &defaultAdapter{}
    for _, opt := range opts {
        opt(a)
    }
    return a
}
```

### Step 5: Write Tests

Create `mymodule_test.go` with table-driven tests covering the port contract.

### Step 6: Register in Workspace

Add to `backend/go.work`:

```
use ./mymodule
```

Run `go work sync` to verify.

## Adding a TypeScript Package

### Step 1: Create the Package Directory

```bash
mkdir -p frontend/packages/mypackage/src
```

### Step 2: Create package.json

```json
{
  "name": "@voxera-kit/mypackage",
  "version": "0.0.0",
  "private": true,
  "type": "module",
  "main": "./src/index.ts",
  "scripts": {
    "build": "tsc --build",
    "lint": "eslint src/",
    "test": "vitest run",
    "typecheck": "tsc --noEmit"
  }
}
```

### Step 3: Create tsconfig.json

```json
{
  "extends": "@voxera-kit/config/tsconfig.base.json",
  "compilerOptions": {
    "outDir": "dist",
    "rootDir": "src"
  },
  "include": ["src"]
}
```

### Step 4: Create Entry Point

Create `src/index.ts` and export your public API.

### Step 5: Write Tests

Add test files alongside source or in a `__tests__/` directory.

### Step 6: Install & Verify

```bash
cd frontend
pnpm install
pnpm turbo build --filter=@voxera-kit/mypackage
pnpm turbo test --filter=@voxera-kit/mypackage
```

## Checklist

- [ ] Port/interface defined (`port.go` or `types.ts`)
- [ ] At least one adapter/implementation
- [ ] Factory function or constructor
- [ ] Tests with reasonable coverage
- [ ] Exported public API via `index.ts` or package-level exports
- [ ] Registered in workspace (`go.work` or auto-discovered by pnpm)
- [ ] Lints and builds cleanly (`make lint && make build`)
