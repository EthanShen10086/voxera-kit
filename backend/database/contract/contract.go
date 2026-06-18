package contract

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/database"
)

// Factory creates a connected database.Database and an optional cleanup function.
type Factory func(t *testing.T) (database.Database, func())

// RunDatabaseContract exercises Ping and transaction lifecycle for database adapters.
func RunDatabaseContract(t *testing.T, factory Factory) {
	t.Helper()
	ctx := context.Background()

	db, cleanup := factory(t)
	if cleanup != nil {
		defer cleanup()
	}
	defer func() { _ = db.Close() }()

	t.Run("Ping", func(t *testing.T) {
		if err := db.Ping(ctx); err != nil {
			t.Fatalf("Ping() = %v, want nil", err)
		}
	})

	t.Run("TransactionCommit", func(t *testing.T) {
		tx := db.Transaction()
		if tx == nil {
			t.Fatal("Transaction() returned nil")
		}
		active, err := tx.Begin(ctx)
		if err != nil {
			t.Fatalf("Begin() = %v", err)
		}
		if active == nil {
			t.Fatal("Begin() returned nil transaction")
		}
		if err := active.Commit(); err != nil {
			t.Fatalf("Commit() = %v", err)
		}
	})

	t.Run("TransactionRollback", func(t *testing.T) {
		tx := db.Transaction()
		if tx == nil {
			t.Fatal("Transaction() returned nil")
		}
		active, err := tx.Begin(ctx)
		if err != nil {
			t.Fatalf("Begin() = %v", err)
		}
		if err := active.Rollback(); err != nil {
			t.Fatalf("Rollback() = %v", err)
		}
	})
}
