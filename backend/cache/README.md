# cache

分布式/进程内缓存 Port（对标 Memorystore）。

## 状态

| Adapter | 状态 |
|---------|------|
| `redis` | 🟡 go-redis/v9 |
| `local` | 🟡 ristretto |
| `memcached` | 🟡 gomemcache |
| `memory` | ✅ 测试用 |

## 测试

```bash
cd backend/cache && go test ./... -race
```

生产环境请为缓存 key 设置 TTL；避免 `KEYS *`。
