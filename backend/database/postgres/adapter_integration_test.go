//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/database"
	"github.com/EthanShen10086/voxera-kit/database/contract"
	postgresadapter "github.com/EthanShen10086/voxera-kit/database/postgres"
	"github.com/EthanShen10086/voxera-kit/testkit/containers"
)

func TestPostgresDatabaseContract(t *testing.T) {
	ctx := context.Background()
	c, err := containers.StartPostgres(ctx)
	if err != nil {
		t.Fatalf("StartPostgres: %v", err)
	}
	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	contract.RunDatabaseContract(t, func(t *testing.T) (database.Database, func()) {
		db := postgresadapter.New()
		if err := db.Connect(ctx, c.Config); err != nil {
			t.Fatalf("Connect: %v", err)
		}
		return db, nil
	})
}

func TestPostgresNestedTransaction(t *testing.T) {
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
	defer func() { _ = db.Close() }()

	outer := db.Transaction()
	tx, err := outer.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin outer: %v", err)
	}
	inner, err := tx.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin savepoint: %v", err)
	}
	if err := inner.Rollback(); err != nil {
		t.Fatalf("Rollback savepoint: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit outer: %v", err)
	}
}
