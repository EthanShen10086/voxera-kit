//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/cache"
	cachecontract "github.com/EthanShen10086/voxera-kit/cache/contract"
	cacheredis "github.com/EthanShen10086/voxera-kit/cache/redis"
	"github.com/EthanShen10086/voxera-kit/database"
	"github.com/EthanShen10086/voxera-kit/mq"
	mqcontract "github.com/EthanShen10086/voxera-kit/mq/contract"
	mqnats "github.com/EthanShen10086/voxera-kit/mq/nats"
	"github.com/EthanShen10086/voxera-kit/storage"
	storagecontract "github.com/EthanShen10086/voxera-kit/storage/contract"
	miniostore "github.com/EthanShen10086/voxera-kit/storage/minio"
	"github.com/EthanShen10086/voxera-kit/testkit/containers"

	postgresadapter "github.com/EthanShen10086/voxera-kit/database/postgres"
)

func TestRedisCacheContract(t *testing.T) {
	ctx := context.Background()
	c, err := containers.StartRedis(ctx)
	if err != nil {
		t.Fatalf("StartRedis: %v", err)
	}
	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	cachecontract.RunCacheContract(t, func(t *testing.T) (cache.Cache, func()) {
		t.Helper()
		return cacheredis.New(cache.Config{
			Address:     c.Address,
			DialTimeout: 5 * time.Second,
			ReadTimeout: 3 * time.Second,
			WriteTimeout: 3 * time.Second,
		}), nil
	})
}

func TestNATSMQContract(t *testing.T) {
	ctx := context.Background()
	c, err := containers.StartNATS(ctx)
	if err != nil {
		t.Fatalf("StartNATS: %v", err)
	}
	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	mqcontract.RunMQContract(t, func(t *testing.T) (mq.Publisher, mq.Subscriber, func()) {
		t.Helper()
		cfg := mq.Config{Brokers: []string{c.URL}}
		pub, err := mqnats.NewPublisher(cfg)
		if err != nil {
			t.Fatalf("NewPublisher: %v", err)
		}
		sub, err := mqnats.NewSubscriber(cfg)
		if err != nil {
			_ = pub.Close()
			t.Fatalf("NewSubscriber: %v", err)
		}
		return pub, sub, nil
	})
}

func TestPostgresPing(t *testing.T) {
	ctx := context.Background()
	c, err := containers.StartPostgres(ctx)
	if err != nil {
		t.Fatalf("StartPostgres: %v", err)
	}
	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	db := postgresadapter.New()
	if err := db.Connect(ctx, c.Config); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Ping: %v", err)
	}

	tx := db.Transaction()
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
	_ = tx // compile-time check database.Transaction is returned
}

func TestMinIOStorageContract(t *testing.T) {
	ctx := context.Background()
	c, err := containers.StartMinIO(ctx, "voxera-test")
	if err != nil {
		t.Fatalf("StartMinIO: %v", err)
	}
	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	storagecontract.RunObjectStoreContract(t, func(t *testing.T) storage.ObjectStore {
		t.Helper()
		store, err := miniostore.New(c.Config)
		if err != nil {
			t.Fatalf("miniostore.New: %v", err)
		}
		return store
	})
}

// Ensure postgres adapter implements database.Database.
var _ database.Database = (*postgresadapter.Adapter)(nil)
