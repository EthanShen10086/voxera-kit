# Coverage ramp (Go backend merged)

Stepped enforcement for `backend/scripts/coverage.sh` / CI `go-coverage` job.

| Phase | `MIN_COVERAGE` | Target release | Status |
|-------|----------------|----------------|--------|
| **0** | **8%** | v0.3.x | done (~9% merged) |
| **1** | **15%** | v0.4.0 | done (~16% merged) |
| **2** | **30%** | v0.5.0 | done (~32% merged) |
| **3** | **50%** | v0.6.0 | done (~50% merged) |
| **4** | **80%** | v1.0.0 | done (~80% merged) |

## Phase 4 sprints (50% → 80%)

| Sprint | Scope | Test strategy |
|--------|-------|---------------|
| **4.1** | `storage/s3`, `mq/nats` | gofakes3 httptest fixture; embedded `nats-server` | done (~53%) |
| **4.2** | `storage/minio`, `database/postgres` | testcontainers (`-tags=integration`) | done |
| **4.3** | `storage/cos/oss`, `mq/kafka/rabbitmq`, partial `fs/middleware/llm` | httptest vendor mocks + testcontainers | done (~57%) |
| **4.4** | `secret/aws/gcp/tencent`, `asr/whisper`, cos/oss lifecycle, `llm/claude/qwen` stream, testfixture smoke | unit + httptest | done (~61%) |
| **4.5** | noop/stub 快赢、`uploadlarge/opts`、`mq` 离线、`hunyuan` stream、`task/redis` Worker、`s3` 版本、ad/cache/scraper 等 | unit + httptest | done (~67%) |
| **4.6** | `secret/vault/gcp` httptest/grpc mock、`storage/minio` internal、database mysql/mongodb、asr stubs、scraper/share/shorturl | unit + httptest | done (~70%) |
| **4.7** | `storage/minio` gofakes3 单测、`secret/tencent` httptest 契约、`mq` 深化、快赢 stub、向 80% 推进 | httptest + gofakes3 | done (~80%) |

`MIN_COVERAGE` **80** enforced in CI and `coverage.sh` default (merged **80.0%** on 2026-06-18).

## CI

`.github/workflows/ci.yml` sets `MIN_COVERAGE` on the `go-coverage` job. Integration tests (`-tags=integration`) run in `go-integration` with Docker; they do not affect merged unit coverage unless promoted to default `go test`.

## Skipped modules

`testkit` is excluded by default (`COVERAGE_SKIP=testkit`) because integration tests require Docker and a separate job.

## Local

```bash
cd backend
MIN_COVERAGE=80 bash scripts/coverage.sh
MIN_COVERAGE=80 COVERAGE_ENFORCE=false bash scripts/coverage.sh  # probe without failing
COVERAGE_TAGS=integration MIN_COVERAGE=80 COVERAGE_ENFORCE=false bash scripts/coverage.sh  # + testcontainers adapters
```
