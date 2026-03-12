// Package embedded provides in-memory adapters for zero-infrastructure demos.
package embedded

import (
	"context"
	"strings"
	"time"

	"github.com/underpass-ai/underpass-demo/internal/domain"
)

// PolicyStore is an in-memory PolicyReader seeded with starship subsystem data.
type PolicyStore struct {
	policies []domain.ToolPolicy
}

// NewPolicyStore returns a store pre-loaded with USS Underpass ship systems.
func NewPolicyStore() *PolicyStore {
	now := time.Now()
	return &PolicyStore{
		policies: []domain.ToolPolicy{
			{ToolID: "nav.plot", ContextSignature: "deep_space", Alpha: 142, Beta: 8, Confidence: 0.9467, ErrorRate: 0.053, P95LatencyMs: 120, P95Cost: 0.02, NSamples: 3800, FreshnessTs: now},
			{ToolID: "scan.deep", ContextSignature: "deep_space", Alpha: 198, Beta: 12, Confidence: 0.9429, ErrorRate: 0.057, P95LatencyMs: 340, P95Cost: 0.08, NSamples: 4200, FreshnessTs: now},
			{ToolID: "comm.burst", ContextSignature: "deep_space", Alpha: 189, Beta: 11, Confidence: 0.9450, ErrorRate: 0.055, P95LatencyMs: 200, P95Cost: 0.03, NSamples: 4000, FreshnessTs: now},
			{ToolID: "life.recycle", ContextSignature: "deep_space", Alpha: 167, Beta: 13, Confidence: 0.9278, ErrorRate: 0.072, P95LatencyMs: 560, P95Cost: 0.04, NSamples: 3600, FreshnessTs: now},
			{ToolID: "hull.seal", ContextSignature: "deep_space", Alpha: 156, Beta: 14, Confidence: 0.9176, ErrorRate: 0.082, P95LatencyMs: 450, P95Cost: 0.05, NSamples: 3400, FreshnessTs: now},
			{ToolID: "power.reroute", ContextSignature: "deep_space", Alpha: 145, Beta: 15, Confidence: 0.9063, ErrorRate: 0.094, P95LatencyMs: 670, P95Cost: 0.10, NSamples: 3200, FreshnessTs: now},
			{ToolID: "shield.mod", ContextSignature: "deep_space", Alpha: 134, Beta: 16, Confidence: 0.8933, ErrorRate: 0.107, P95LatencyMs: 780, P95Cost: 0.12, NSamples: 3000, FreshnessTs: now},
			{ToolID: "eng.thrust", ContextSignature: "deep_space", Alpha: 175, Beta: 25, Confidence: 0.8750, ErrorRate: 0.125, P95LatencyMs: 890, P95Cost: 0.15, NSamples: 4000, FreshnessTs: now},
		},
	}
}

func (s *PolicyStore) ReadAll(_ context.Context) ([]domain.ToolPolicy, error) {
	out := make([]domain.ToolPolicy, len(s.policies))
	copy(out, s.policies)
	return out, nil
}

func (s *PolicyStore) ReadByContext(_ context.Context, contextSig string) ([]domain.ToolPolicy, error) {
	var out []domain.ToolPolicy
	for _, p := range s.policies {
		if strings.EqualFold(p.ContextSignature, contextSig) {
			out = append(out, p)
		}
	}
	return out, nil
}

func (s *PolicyStore) Close() error { return nil }
