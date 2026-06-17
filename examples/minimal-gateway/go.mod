module github.com/EthanShen10086/voxera-kit/examples/minimal-gateway

go 1.22

require (
	github.com/EthanShen10086/voxera-kit/circuitbreaker v0.0.0
	github.com/EthanShen10086/voxera-kit/middleware v0.0.0
)

replace (
	github.com/EthanShen10086/voxera-kit/circuitbreaker => ../../backend/circuitbreaker
	github.com/EthanShen10086/voxera-kit/middleware => ../../backend/middleware
)
