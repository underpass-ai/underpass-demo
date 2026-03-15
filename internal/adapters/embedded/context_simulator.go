package embedded

import (
	"context"

	"github.com/underpass-ai/underpass-demo/internal/domain"
)

// ContextSimulator implements ports.ContextProvider with hardcoded graph data
// matching the USS Underpass mission scenario (phases 7-8).
type ContextSimulator struct{}

// NewContextSimulator returns a zero-infra context provider.
func NewContextSimulator() *ContextSimulator {
	return &ContextSimulator{}
}

func (s *ContextSimulator) GetContext(_ context.Context, req domain.ContextRequest) (*domain.ContextResult, error) {
	return &domain.ContextResult{
		RootNodeID: req.RootNodeID,
		Nodes: []domain.GraphNode{
			{ID: "mission:engine-core-failure", Kind: "mission", Label: "Engine Core Failure", Status: "active"},
			{ID: "task:diagnose-anomaly", Kind: "task", Label: "Diagnose anomaly", Status: "done"},
			{ID: "task:assess-cascade", Kind: "task", Label: "Assess cascade damage", Status: "done"},
			{ID: "task:direct-engine-repair", Kind: "task", Label: "Direct engine repair", Status: "abandoned"},
			{ID: "task:seal-hull", Kind: "task", Label: "Seal hull breaches", Status: "active"},
			{ID: "task:stabilize-power", Kind: "task", Label: "Stabilize power grid", Status: "pending"},
			{ID: "task:repair-engine-safe", Kind: "task", Label: "Repair engine (safe conditions)", Status: "pending"},
		},
		Relationships: []domain.GraphRelationship{
			{SourceID: "mission:engine-core-failure", TargetID: "task:diagnose-anomaly", Kind: "has_subtask"},
			{SourceID: "task:diagnose-anomaly", TargetID: "task:assess-cascade", Kind: "depends_on"},
			{SourceID: "task:assess-cascade", TargetID: "task:direct-engine-repair", Kind: "has_subtask", Label: "Path A"},
			{SourceID: "task:assess-cascade", TargetID: "task:seal-hull", Kind: "has_subtask", Label: "Path B"},
			{SourceID: "task:seal-hull", TargetID: "task:stabilize-power", Kind: "depends_on"},
			{SourceID: "task:stabilize-power", TargetID: "task:repair-engine-safe", Kind: "depends_on"},
		},
		Details: []domain.NodeDetail{
			{NodeID: "task:direct-engine-repair", Title: "Direct engine repair",
				Description: "3 attempts. Hull stress +12%. Counterproductive.",
				Properties:  map[string]string{"model": "qwen3-8b", "escalated_to": "claude-opus"}},
			{NodeID: "task:seal-hull", Title: "Seal hull breaches",
				Description: "Hull-first protocol. Structural integrity target: 98%.",
				Properties:  map[string]string{"model": "qwen3-8b"}},
			{NodeID: "task:repair-engine-safe", Title: "Repair engine under safe conditions",
				Description: "Engine repair after hull stabilized.",
				Properties:  map[string]string{"model": "qwen3-8b"}},
		},
		Sections: []domain.BundleSection{
			{Heading: "Mission Context", Content: "USS Underpass engine core failure. Cascade to power and shields.", TokenCount: 42},
			{Heading: "Failed Approach", Content: "Path A: Direct engine repair failed 3 times. Hull stress +12%.", TokenCount: 38},
			{Heading: "Active Strategy", Content: "Path B: Hull-first protocol. Seal → stabilize → repair.", TokenCount: 35},
		},
		TokenCount:  394,
		ContentHash: "sha256:a1b2c3d4e5f6",
		SnapshotID:  "snap_uss_20260312T154230Z",
	}, nil
}

func (s *ContextSimulator) GetGraphRelationships(_ context.Context, nodeID, _ string, _ uint32) (*domain.GraphResult, error) {
	ctx, _ := s.GetContext(context.Background(), domain.ContextRequest{RootNodeID: nodeID})
	root := domain.GraphNode{ID: nodeID, Kind: "mission", Label: "Engine Core Failure", Status: "active"}
	return &domain.GraphResult{
		Root:          root,
		Neighbors:     ctx.Nodes[1:], // skip root
		Relationships: ctx.Relationships,
	}, nil
}
