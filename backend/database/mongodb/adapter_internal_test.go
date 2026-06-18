package mongodb

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/database"
)

func TestURIDefaults(t *testing.T) {
	u := uri(database.Config{Host: "localhost", Database: "app"})
	if u != "mongodb://localhost:27017/app" {
		t.Fatalf("uri = %q", u)
	}
}

func TestURIWithAuth(t *testing.T) {
	u := uri(database.Config{
		User: "u", Password: "p", Host: "db", Port: 27018, Database: "",
	})
	if u != "mongodb://u:p@db:27018/admin" {
		t.Fatalf("uri = %q", u)
	}
}
