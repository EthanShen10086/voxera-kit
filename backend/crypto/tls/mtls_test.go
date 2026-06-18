package tls_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	tlskit "github.com/EthanShen10086/voxera-kit/crypto/tls"
)

func writeCertFiles(t *testing.T) (certFile, keyFile, caFile string) {
	t.Helper()
	dir := t.TempDir()

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	caTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	caFile = filepath.Join(dir, "ca.pem")
	writePEM(t, caFile, "CERTIFICATE", caDER)

	leafKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	leafTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTmpl, caTmpl, &leafKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	certFile = filepath.Join(dir, "cert.pem")
	keyFile = filepath.Join(dir, "key.pem")
	writePEM(t, certFile, "CERTIFICATE", leafDER)
	writeKeyPEM(t, keyFile, leafKey)
	return certFile, keyFile, caFile
}

func writePEM(t *testing.T, path, typ string, der []byte) {
	t.Helper()
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: typ, Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
}

func writeKeyPEM(t *testing.T, path string, key *rsa.PrivateKey) {
	t.Helper()
	der := x509.MarshalPKCS1PrivateKey(key)
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestNewServerTLSConfig_MissingFiles(t *testing.T) {
	_, err := tlskit.NewServerTLSConfig(tlskit.Config{
		CertFile: "/no/such/cert.pem",
		KeyFile:  "/no/such/key.pem",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewClientTLSConfig_MissingFiles(t *testing.T) {
	_, err := tlskit.NewClientTLSConfig(tlskit.Config{
		CertFile: "/no/such/cert.pem",
		KeyFile:  "/no/such/key.pem",
		CAFile:   "/no/such/ca.pem",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewServerTLSConfig_WithClientAuth(t *testing.T) {
	certFile, keyFile, caFile := writeCertFiles(t)
	tc, err := tlskit.NewServerTLSConfig(tlskit.Config{
		CertFile:          certFile,
		KeyFile:           keyFile,
		CAFile:            caFile,
		RequireClientCert: true,
		MinVersion:        tls.VersionTLS10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if tc.MinVersion != tls.VersionTLS12 {
		t.Fatalf("MinVersion = %x", tc.MinVersion)
	}
	if tc.ClientAuth != tls.RequireAndVerifyClientCert {
		t.Fatalf("ClientAuth = %v", tc.ClientAuth)
	}
}

func TestNewClientTLSConfig_WithCA(t *testing.T) {
	certFile, keyFile, caFile := writeCertFiles(t)
	tc, err := tlskit.NewClientTLSConfig(tlskit.Config{
		CertFile: certFile,
		KeyFile:  keyFile,
		CAFile:   caFile,
	})
	if err != nil {
		t.Fatal(err)
	}
	if tc.RootCAs == nil {
		t.Fatal("expected RootCAs")
	}
}

func TestLoadCAInvalidPEM(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.pem")
	if err := os.WriteFile(bad, []byte("not a cert"), 0o600); err != nil {
		t.Fatal(err)
	}
	certFile, keyFile, _ := writeCertFiles(t)
	_, err := tlskit.NewClientTLSConfig(tlskit.Config{
		CertFile: certFile,
		KeyFile:  keyFile,
		CAFile:   bad,
	})
	if err == nil {
		t.Fatal("expected invalid CA error")
	}
}

func TestNewServerTLSConfig_RequireClientWithoutCA(t *testing.T) {
	certFile, keyFile, _ := writeCertFiles(t)
	tc, err := tlskit.NewServerTLSConfig(tlskit.Config{
		CertFile:          certFile,
		KeyFile:           keyFile,
		RequireClientCert: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if tc.ClientCAs != nil {
		t.Fatal("expected nil ClientCAs when CAFile empty")
	}
}
