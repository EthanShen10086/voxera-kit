package tls_test

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/crypto/tls"
)

func TestNewServerTLSConfig_MissingFiles(t *testing.T) {
	_, err := tls.NewServerTLSConfig(tls.Config{
		CertFile: "/no/such/cert.pem",
		KeyFile:  "/no/such/key.pem",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewClientTLSConfig_MissingFiles(t *testing.T) {
	_, err := tls.NewClientTLSConfig(tls.Config{
		CertFile: "/no/such/cert.pem",
		KeyFile:  "/no/such/key.pem",
		CAFile:   "/no/such/ca.pem",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
