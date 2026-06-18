package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/database"
	postgresadapter "github.com/EthanShen10086/voxera-kit/database/postgres"
)

func TestAdapterNotConnected(t *testing.T) {
	a := postgresadapter.New()
	ctx := context.Background()

	if err := a.Ping(ctx); err == nil {
		t.Fatal("Ping() expected error when not connected")
	}
	if err := a.Close(); err != nil {
		t.Fatalf("Close() = %v", err)
	}

	tx := a.Transaction()
	if _, err := tx.Begin(ctx); err == nil {
		t.Fatal("Begin() expected error when pool is nil")
	}
	if err := tx.Commit(); err == nil {
		t.Fatal("Commit() expected error when tx not started")
	}
	if err := tx.Rollback(); err == nil {
		t.Fatal("Rollback() expected error when tx not started")
	}
}

func TestConnectUnreachable(t *testing.T) {
	a := postgresadapter.New()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := a.Connect(ctx, database.Config{
		Host:            "127.0.0.1",
		Port:            1,
		User:            "voxera",
		Password:        "voxera",
		Database:        "voxera_test",
		SSLMode:         "disable",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
	})
	if err == nil {
		_ = a.Close()
		t.Fatal("Connect() expected error for unreachable host")
	}
}
