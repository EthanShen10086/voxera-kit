// Package tls provides helpers for building mutual-TLS (mTLS) configurations
// used by both servers and clients in the voxera-kit ecosystem.
package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// Config holds the file paths and options needed to construct a [tls.Config].
type Config struct {
	// CertFile is the path to the PEM-encoded certificate.
	CertFile string
	// KeyFile is the path to the PEM-encoded private key.
	KeyFile string
	// CAFile is the path to the PEM-encoded CA bundle used to verify the
	// peer's certificate. When empty, the system root CA pool is used.
	CAFile string
	// RequireClientCert controls whether the server demands a valid client
	// certificate during the TLS handshake.
	RequireClientCert bool
	// MinVersion is the minimum TLS version to accept. Values below
	// [tls.VersionTLS12] are silently raised to TLS 1.2.
	MinVersion uint16
}

// effectiveMinVersion returns at least TLS 1.2.
func effectiveMinVersion(v uint16) uint16 {
	if v < tls.VersionTLS12 {
		return tls.VersionTLS12
	}
	return v
}

// loadCA reads a PEM CA bundle from disk and returns a certificate pool.
func loadCA(path string) (*x509.CertPool, error) {
	pem, err := os.ReadFile(path) // #nosec G304 -- path is caller-controlled configuration
	if err != nil {
		return nil, fmt.Errorf("tls: read CA file: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("tls: no valid certificates found in %s", path)
	}
	return pool, nil
}

// NewServerTLSConfig builds a [tls.Config] suitable for an HTTPS server.
//
// If [Config.RequireClientCert] is true and [Config.CAFile] is set the
// returned config verifies client certificates against the supplied CA.
// The minimum TLS version is always at least TLS 1.2.
func NewServerTLSConfig(cfg Config) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("tls: load server key pair: %w", err)
	}

	tc := &tls.Config{ // #nosec G402 -- effectiveMinVersion enforces >= TLS 1.2
		Certificates: []tls.Certificate{cert},
		MinVersion:   effectiveMinVersion(cfg.MinVersion),
	}

	if cfg.RequireClientCert {
		tc.ClientAuth = tls.RequireAndVerifyClientCert
		if cfg.CAFile != "" {
			pool, err := loadCA(cfg.CAFile)
			if err != nil {
				return nil, err
			}
			tc.ClientCAs = pool
		}
	}

	return tc, nil
}

// NewClientTLSConfig builds a [tls.Config] suitable for an mTLS HTTP client.
//
// The client presents its own certificate and verifies the server against the
// CA bundle specified in [Config.CAFile]. The minimum TLS version is always at
// least TLS 1.2.
func NewClientTLSConfig(cfg Config) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("tls: load client key pair: %w", err)
	}

	tc := &tls.Config{ // #nosec G402 -- effectiveMinVersion enforces >= TLS 1.2
		Certificates: []tls.Certificate{cert},
		MinVersion:   effectiveMinVersion(cfg.MinVersion),
	}

	if cfg.CAFile != "" {
		pool, err := loadCA(cfg.CAFile)
		if err != nil {
			return nil, err
		}
		tc.RootCAs = pool
	}

	return tc, nil
}
