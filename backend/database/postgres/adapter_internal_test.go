package postgres

import (
	"strings"
	"testing"

	"github.com/EthanShen10086/voxera-kit/database"
)

func TestDSNDefaults(t *testing.T) {
	d := dsn(database.Config{
		User: "u", Password: "p", Host: "localhost", Database: "db",
	})
	if !strings.Contains(d, "postgres://u:p@localhost:5432/db") {
		t.Fatalf("dsn = %q", d)
	}
	if !strings.Contains(d, "sslmode=disable") {
		t.Fatalf("expected sslmode=disable in %q", d)
	}
}

func TestDSNCustomPortAndSSL(t *testing.T) {
	d := dsn(database.Config{
		User: "u", Password: "p", Host: "pg", Port: 5433, Database: "app", SSLMode: "require",
	})
	if !strings.Contains(d, "pg:5433") || !strings.Contains(d, "sslmode=require") {
		t.Fatalf("dsn = %q", d)
	}
}
