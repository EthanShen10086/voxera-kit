package contract

import (
	"context"
	"errors"
	"testing"

	"github.com/EthanShen10086/voxera-kit/database"
)

type mockDatabase struct {
	pingErr error
	closed  bool
}

func (m *mockDatabase) Connect(_ context.Context, _ database.Config) error { return nil }

func (m *mockDatabase) Close() error {
	m.closed = true
	return nil
}

func (m *mockDatabase) Ping(_ context.Context) error {
	return m.pingErr
}

func (m *mockDatabase) Transaction() database.Transaction {
	return &mockTransaction{}
}

type mockTransaction struct{}

func (m *mockTransaction) Begin(_ context.Context) (database.Transaction, error) {
	return m, nil
}

func (m *mockTransaction) Commit() error { return nil }

func (m *mockTransaction) Rollback() error { return nil }

func TestDatabaseContract_Mock(t *testing.T) {
	RunDatabaseContract(t, func(t *testing.T) (database.Database, func()) {
		return &mockDatabase{}, nil
	})
}

func TestDatabasePingError(t *testing.T) {
	want := errors.New("connection refused")
	var db database.Database = &mockDatabase{pingErr: want}
	if err := db.Ping(context.Background()); !errors.Is(err, want) {
		t.Fatalf("Ping() = %v, want %v", err, want)
	}
}

func TestDatabaseClose(t *testing.T) {
	mock := &mockDatabase{}
	if err := mock.Close(); err != nil {
		t.Fatalf("Close() = %v, want nil", err)
	}
	if !mock.closed {
		t.Fatal("Close() did not mark database as closed")
	}
}
