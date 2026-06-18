module github.com/EthanShen10086/voxera-kit/testkit

go 1.22.0

require (
	github.com/EthanShen10086/voxera-kit/cache v0.0.0
	github.com/EthanShen10086/voxera-kit/database v0.0.0
	github.com/EthanShen10086/voxera-kit/mq v0.0.0
	github.com/EthanShen10086/voxera-kit/secret v0.0.0
	github.com/EthanShen10086/voxera-kit/storage v0.0.0
	github.com/EthanShen10086/voxera-kit/task v0.0.0
	github.com/testcontainers/testcontainers-go v0.35.0
	github.com/testcontainers/testcontainers-go/modules/minio v0.35.0
	github.com/testcontainers/testcontainers-go/modules/nats v0.35.0
	github.com/testcontainers/testcontainers-go/modules/postgres v0.35.0
	github.com/testcontainers/testcontainers-go/modules/redis v0.35.0
)

replace (
	github.com/EthanShen10086/voxera-kit/cache => ../cache
	github.com/EthanShen10086/voxera-kit/database => ../database
	github.com/EthanShen10086/voxera-kit/mq => ../mq
	github.com/EthanShen10086/voxera-kit/secret => ../secret
	github.com/EthanShen10086/voxera-kit/storage => ../storage
	github.com/EthanShen10086/voxera-kit/task => ../task
)
