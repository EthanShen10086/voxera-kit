# llm

**Status:** Production (port + router) · provider adapters vary

Unified LLM provider interface with OpenAI, Qwen, Claude, DeepSeek, Hunyuan adapters.

| Adapter | Status |
|---------|--------|
| `noop/` | Stub (tests only) |
| `openai/`, `qwen/` | Beta |
| Others | Beta — verify before production |

## Tests

`noop/adapter_test.go` — provider contract smoke test.
