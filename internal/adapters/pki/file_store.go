// Package pki persists mTLS credentials to the local filesystem.
// Pattern borrowed from fleetctl — four PEM files under a config directory.
package pki

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/underpass-ai/underpass-demo/internal/domain/identity"
)

const (
	fileClientKey  = "client.key"
	fileClientCert = "client.crt"
	fileCACert     = "ca.crt"
	fileServerName = "server_name"
	dirPerm        = 0700
	filePerm       = 0600
)

// FileStore persists credentials as PEM files in a local directory.
type FileStore struct {
	dir string
}

// NewFileStore creates a file-based credential store.
func NewFileStore(configDir string) *FileStore {
	return &FileStore{dir: filepath.Join(configDir, "pki")}
}

// Save writes all credential files to disk with secure permissions.
func (s *FileStore) Save(certPEM, keyPEM, caPEM []byte, serverName string) error {
	if err := os.MkdirAll(s.dir, dirPerm); err != nil {
		return fmt.Errorf("create pki dir: %w", err)
	}

	files := []struct {
		name string
		data []byte
	}{
		{fileClientKey, keyPEM},
		{fileClientCert, certPEM},
		{fileCACert, caPEM},
		{fileServerName, []byte(serverName)},
	}

	for _, f := range files {
		if err := os.WriteFile(filepath.Join(s.dir, f.name), f.data, filePerm); err != nil {
			return fmt.Errorf("write %s: %w", f.name, err)
		}
	}
	return nil
}

// Load reads credentials from disk.
func (s *FileStore) Load() (identity.Credentials, error) {
	certPEM, err := os.ReadFile(filepath.Join(s.dir, fileClientCert))
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("read client cert: %w", err)
	}
	keyPEM, err := os.ReadFile(filepath.Join(s.dir, fileClientKey))
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("read client key: %w", err)
	}
	caPEM, err := os.ReadFile(filepath.Join(s.dir, fileCACert))
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("read CA cert: %w", err)
	}
	serverName, err := os.ReadFile(filepath.Join(s.dir, fileServerName))
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("read server name: %w", err)
	}
	return identity.NewCredentials(certPEM, keyPEM, caPEM, string(serverName))
}

// Exists returns true if credential files exist.
func (s *FileStore) Exists() bool {
	_, err := os.Stat(filepath.Join(s.dir, fileClientCert))
	return err == nil
}
