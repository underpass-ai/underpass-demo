// Package ports defines the inbound interfaces for the demo client.
package ports

import (
	"context"

	"github.com/underpass-ai/underpass-demo/internal/domain"
	"github.com/underpass-ai/underpass-demo/internal/domain/identity"
)

// PolicyReader reads tool policies from the policy store (Valkey).
type PolicyReader interface {
	// ReadAll returns all policies matching the key prefix.
	ReadAll(ctx context.Context) ([]domain.ToolPolicy, error)
	// ReadByContext returns policies for a specific context signature.
	ReadByContext(ctx context.Context, contextSig string) ([]domain.ToolPolicy, error)
}

// EventSubscriber subscribes to real-time policy update events (NATS).
type EventSubscriber interface {
	// Subscribe returns a channel of policy update events.
	Subscribe(ctx context.Context) (<-chan domain.PolicyUpdateEvent, error)
	// Close unsubscribes and cleans up.
	Close()
}

// CredentialStore persists mTLS credentials.
type CredentialStore interface {
	Save(certPEM, keyPEM, caPEM []byte, serverName string) error
	Load() (identity.Credentials, error)
	Exists() bool
}
