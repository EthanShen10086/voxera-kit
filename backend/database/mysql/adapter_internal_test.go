package mysql

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/database"
)

func TestDSNDefaults(t *testing.T) {
	d := dsn(database.Config{
		User:     "u",
		Password: "p",
		Host:     "localhost",
		Database: "db",
	})
	if d != "u:p@tcp(localhost:3306)/db?parseTime=true&tls=false" {
		t.Fatalf("dsn = %q", d)
	}
}

func TestDSNTLSRequire(t *testing.T) {
	d := dsn(database.Config{
		User: "u", Password: "p", Host: "h", Port: 3307, Database: "d", SSLMode: "require",
	})
	if d != "u:p@tcp(h:3307)/d?parseTime=true&tls=true" {
		t.Fatalf("dsn = %q", d)
	}
}
