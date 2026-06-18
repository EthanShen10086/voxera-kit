# fixture

Wave T3 测试造数包：通用 ID / 时间 / HTTP 请求 builder，以及媒体对象 key 与样例字节。

## 用法

```go
import (
    "github.com/EthanShen10086/voxera-kit/fixture"
    mediafixture "github.com/EthanShen10086/voxera-kit/fixture/media"
)

id := fixture.NewPrefixedID("job")
now := fixture.FixedTime()

req, err := fixture.NewHTTPRequest("POST", "http://localhost/v1/items").
    WithBearerToken("test-token").
    WithJSONBody(map[string]string{"name": "demo"})
req, err = req.Build()

key := mediafixture.AudioObjectKey("user-1", "rec-1")
payload := mediafixture.SamplePNG()
```

领域数据（行情、财报等）仍使用 `dataprovider/stub`；跨模块测试通用造数走 `fixture`。
