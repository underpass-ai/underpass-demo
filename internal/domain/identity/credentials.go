// Package identity holds mTLS credential management,
// following the same pattern as fleetctl's identity package.
package identity

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"
)

// Credentials holds the mTLS identity for secure communication.
type Credentials struct {
	clientCert tls.Certificate
	caPool     *x509.CertPool
	expiresAt  time.Time
	serverName string
}

// NewCredentials creates credentials from PEM-encoded cert, key, and CA.
func NewCredentials(certPEM, keyPEM, caPEM []byte, serverName string) (Credentials, error) {
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return Credentials{}, fmt.Errorf("parse key pair: %w", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return Credentials{}, fmt.Errorf("parse CA certificate")
	}

	var expiresAt time.Time
	if len(cert.Certificate) > 0 {
		parsed, err := x509.ParseCertificate(cert.Certificate[0])
		if err == nil {
			expiresAt = parsed.NotAfter
		}
	}

	return Credentials{
		clientCert: cert,
		caPool:     pool,
		expiresAt:  expiresAt,
		serverName: serverName,
	}, nil
}

// TLSConfig returns a *tls.Config suitable for gRPC transport credentials.
func (c Credentials) TLSConfig() *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{c.clientCert},
		RootCAs:      c.caPool,
		ServerName:   c.serverName,
		MinVersion:   tls.VersionTLS13,
	}
}

// ExpiresAt returns the certificate expiration time.
func (c Credentials) ExpiresAt() time.Time {
	return c.expiresAt
}

// ServerName returns the TLS server name override.
func (c Credentials) ServerName() string {
	return c.serverName
}
