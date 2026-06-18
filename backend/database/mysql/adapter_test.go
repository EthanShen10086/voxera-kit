package mysql_test

import (
	"context"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/database"
	mysqladapter "github.com/EthanShen10086/voxera-kit/database/mysql"
)

func TestAdapterNotConnected(t *testing.T) {
	a := mysqladapter.New()
	ctx := context.Background()
	if err := a.Ping(ctx); err == nil {
		t.Fatal("expected ping error")
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
	tx := a.Transaction()
	if _, err := tx.Begin(ctx); err == nil {
		t.Fatal("expected begin error")
	}
}

func TestConnectUnreachable(t *testing.T) {
	a := mysqladapter.New()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := a.Connect(ctx, database.Config{
		Host: "127.0.0.1", Port: 1, User: "u", Password: "p", Database: "db",
		SSLMode: "preferred", MaxOpenConns: 3, MaxIdleConns: 1, ConnMaxLifetime: time.Minute,
	})
	if err == nil {
		_ = a.Close()
		t.Fatal("expected connect error")
	}
}
