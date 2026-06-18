# cache

分布式/进程内缓存 Port（对标 Memorystore）。

## 状态

| Adapter | 状态 |
|---------|------|
| `redis` | 🟡 go-redis/v9 |
| `local` | 🟡 ristretto |
| `memcached` | 🟡 gomemcache |
| `memory` | ✅ 测试用 |
| `tiered` | ✅ L1 + L2 多级组合 |

## 多级缓存（tiered）

典型组合：**L1 `local`（ristretto）+ L2 `redis`**，读穿、写穿、逐层回填。

```go
l1, err := local.New(cache.Config{})
if err != nil {
    return err
}
l2 := redis.New(cache.Config{Address: "localhost:6379"})

store, err := tiered.New(l1, l2)
if err != nil {
    return err
}
defer store.Close()

ctx := context.Background()
_ = store.SetWithTTL(ctx, "user:42", []byte("profile"), time.Minute)
val, err := store.Get(ctx, "user:42")
```

行为：

- **Get**：自上而下查找；下层命中时回填上层
- **Set / SetWithTTL / Delete / Flush**：所有层同步
- **Exists**：任一层存在即 true

单测可用 `tiered.New(memory.New(), memory.New())` 跑契约。

性能对标与基准方法见 **[Memorystore 性能基准文档](../../docs/cache-memorystore-benchmark.md)**（含 `go test -bench`）。

## 测试

```bash
cd backend/cache && go test ./... -race
```

生产环境请为缓存 key 设置 TTL；避免 `KEYS *`。
