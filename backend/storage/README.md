# storage

对象存储 Port + 可插拔 Adapter（对标 GCS / S3 / MinIO / OSS / COS）。

## 状态

| Adapter | 状态 | SDK |
|---------|------|-----|
| `memory` | ✅ 契约 + 单测 | 内存 fake（含版本/生命周期/通知） |
| `fs` | ✅ | 本地目录 |
| `minio` | ✅ | minio-go（S3 兼容：MinIO/R2/COS endpoint） |
| `s3` | ✅ | aws-sdk-go-v2 |
| `oss` | ✅ | 阿里云 OSS |
| `cos` | ✅ | 腾讯云 COS |

## 接口分层

- `ObjectStore` — 基础 CRUD、预签名 URL
- `MultipartUploader` / `LargeObjectStore` — 大文件分片
- `VersionedObjectStore` / `StorageAdmin` — 版本、生命周期、桶通知

产品只依赖所需接口；领域 key 规范留在产品层。

## 选型

| 场景 | 推荐 |
|------|------|
| 本地 / 单测 | `memory` |
| MinIO / R2 / COS S3 API | `minio` 或 `s3` + 自定义 Endpoint |
| AWS 原生 | `s3` |
| 阿里云非 S3 特性 | `oss` |
| 腾讯云非 S3 特性 | `cos` |

## COS / R2 配置示例（minio adapter）

```go
store, err := minio.New(storage.Config{
    Endpoint:  "cos.ap-guangzhou.myqcloud.com",
    AccessKey: os.Getenv("COS_SECRET_ID"),
    SecretKey: os.Getenv("COS_SECRET_KEY"),
    Bucket:    "my-bucket",
    Region:    "ap-guangzhou",
    UseSSL:    true,
})
```

## 测试

```bash
cd backend/storage
go test ./... -race
# 集成测试（需 MinIO，见 docs/TESTING_INFRA_PLAN.md Wave T1）
go test -tags=integration ./... 
```

## 说明

- **cosign**（容器镜像签名）与腾讯云 **COS**（对象存储）无关。
