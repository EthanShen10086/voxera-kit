# task

Cloud Tasks–style delayed queue port with in-memory and Redis backends.

## Port

| Interface | Methods |
|-----------|---------|
| `TaskQueue` | `Enqueue`, `Schedule`, `Cancel` |
| `DeadLetterQueue` | `DeadLetterLen` (memory/redis adapters) |

`Task` fields:

- `IdempotencyKey` — duplicate enqueue is ignored when the key was already processed or pending
- `Retry` — `MaxAttempts` (default 3) and `Backoff` (default 100ms in memory)
- `Attempt` — set by the worker on each execution

## Adapters

| Package | Role |
|---------|------|
| `task/memory` | Goroutine + `time.AfterFunc`; built-in retry and DLQ |
| `task/redis` | ZSET schedule + payload keys; use with `redis.Worker` |

## Worker (Redis)

```go
worker, err := redis.NewWorker(redis.WorkerConfig{
    Adapter: queue,
    Handler: func(ctx context.Context, t task.Task) error {
        return process(t)
    },
})
go worker.Run(ctx)
```

The worker polls due tasks, runs the handler, retries with backoff, and pushes to `{prefix}:dlq` when attempts are exhausted.

## vs `scheduler`

| Module | Use when |
|--------|----------|
| **`scheduler/cron`** | Fixed cron expressions, periodic jobs |
| **`task`** | One-off or delayed work, retries, idempotency, DLQ |
| **`mq`** | Event-driven fan-out; combine with W1.5 bucket notifications |

Do not use `task` for complex DAG workflows — compose `task` + `mq` + `scheduler` instead.

## Contract tests

```bash
cd backend/task && go test ./contract/...
```

- `RunTaskContract` — enqueue, schedule, cancel
- `RunTaskAdvancedContract` — idempotency, retry, dead letter

## Production checklist

- Redis: dedicated DB index, key prefix per service, monitor `{prefix}:dlq` length
- Idempotency keys: stable business IDs (order ID, payment ID)
- Handler must be safe to retry (at-least-once semantics)
