# Coverage ramp (Go backend merged)

Stepped enforcement for `backend/scripts/coverage.sh` / CI `go-coverage` job.

| Phase | `MIN_COVERAGE` | Target release | Status |
|-------|----------------|----------------|--------|
| **0** | **8%** | v0.3.x (current) | enforce now (~9% merged) |
| **1** | **15%** | v0.4.0 | planned |
| **2** | **30%** | v0.5.0 | planned |
| **3** | **50%** | v0.6.0 | planned |
| **4** | **80%** | v1.0.0 | planned |

## CI

`.github/workflows/ci.yml` sets `MIN_COVERAGE` on the `go-coverage` job. Bump when merged coverage consistently exceeds the next milestone.

## Skipped modules

`testkit` is excluded by default (`COVERAGE_SKIP=testkit`) because integration tests require Docker and a separate job.

## Local

```bash
cd backend
MIN_COVERAGE=8 bash scripts/coverage.sh
```
