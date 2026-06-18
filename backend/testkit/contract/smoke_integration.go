//go:build integration

package contract

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/cache"
	cacheredis "github.com/EthanShen10086/voxera-kit/cache/redis"
	"github.com/EthanShen10086/voxera-kit/database"
	"github.com/EthanShen10086/voxera-kit/mq"
	mqnats "github.com/EthanShen10086/voxera-kit/mq/nats"
	"github.com/EthanShen10086/voxera-kit/secret"
	"github.com/EthanShen10086/voxera-kit/secret/env"
	"github.com/EthanShen10086/voxera-kit/storage"
	miniostore "github.com/EthanShen10086/voxera-kit/storage/minio"
	s3store "github.com/EthanShen10086/voxera-kit/storage/s3"
	"github.com/EthanShen10086/voxera-kit/task"
	"github.com/EthanShen10086/voxera-kit/task/memory"
	"github.com/EthanShen10086/voxera-kit/testkit/containers"

	postgresadapter "github.com/EthanShen10086/voxera-kit/database/postgres"
)

// RunDataPlaneSmoke runs container-backed contract tests for cache, mq, database, and storage.
func RunDataPlaneSmoke(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	t.Run("CacheRedis", func(t *testing.T) {
		c, err := containers.StartRedis(ctx)
		if err != nil {
			t.Fatalf("StartRedis: %v", err)
		}
		t.Cleanup(func() { _ = c.Terminate(context.Background()) })

		RunCacheContract(t, func(t *testing.T) (cache.Cache, func()) {
			return cacheredis.New(cache.Config{
				Address:      c.Address,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
			}), nil
		})
	})

	t.Run("MQNATS", func(t *testing.T) {
		c, err := containers.StartNATS(ctx)
		if err != nil {
			t.Fatalf("StartNATS: %v", err)
		}
		t.Cleanup(func() { _ = c.Terminate(context.Background()) })

		RunMQContract(t, func(t *testing.T) (mq.Publisher, mq.Subscriber, func()) {
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
	})

	t.Run("DatabasePostgres", func(t *testing.T) {
		c, err := containers.StartPostgres(ctx)
		if err != nil {
			t.Fatalf("StartPostgres: %v", err)
		}
		t.Cleanup(func() { _ = c.Terminate(context.Background()) })

		RunDatabaseContract(t, func(t *testing.T) (database.Database, func()) {
			db := postgresadapter.New()
			if err := db.Connect(ctx, c.Config); err != nil {
				t.Fatalf("Connect: %v", err)
			}
			return db, nil
		})
	})

	t.Run("StorageMinIO", func(t *testing.T) {
		c, err := containers.StartMinIO(ctx, "voxera-test")
		if err != nil {
			t.Fatalf("StartMinIO: %v", err)
		}
		t.Cleanup(func() { _ = c.Terminate(context.Background()) })

		RunObjectStoreContract(t, func(t *testing.T) storage.ObjectStore {
			store, err := miniostore.New(c.Config)
			if err != nil {
				t.Fatalf("miniostore.New: %v", err)
			}
			return store
		})
	})

	t.Run("StorageS3Compat", func(t *testing.T) {
		c, err := containers.StartMinIO(ctx, "voxera-s3-test")
		if err != nil {
			t.Fatalf("StartMinIO: %v", err)
		}
		t.Cleanup(func() { _ = c.Terminate(context.Background()) })

		RunObjectStoreContract(t, func(t *testing.T) storage.ObjectStore {
			store, err := s3store.New(c.S3CompatConfig())
			if err != nil {
				t.Fatalf("s3store.New: %v", err)
			}
			return store
		})
	})

	t.Run("SecretEnv", func(t *testing.T) {
		RunSecretContract(t, func(t *testing.T) (secret.Manager, func()) {
			prefix := "VOXERA_SMOKE_" + strings.ReplaceAll(t.Name(), "/", "_")
			mgr := env.NewManager(prefix)
			return mgr, func() {
				for _, key := range []string{"api-key", "delete-me", "prefix/a", "prefix/b", "other/c"} {
					_ = mgr.Delete(context.Background(), key)
				}
			}
		})
	})

	t.Run("TaskMemory", func(t *testing.T) {
		RunTaskContract(t, func(t *testing.T, handler task.Handler) (task.TaskQueue, func()) {
			return memory.New(memory.Config{Handler: handler}), nil
		})
	})
}
